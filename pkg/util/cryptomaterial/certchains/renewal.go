package certchains

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

// The existing library-go signer invokes Next after building a certificate
// template and before signing it. Bounding the template here avoids rounding
// races from shortening a duration before key generation takes place.
type boundedSerialGenerator struct {
	crypto.SerialGenerator

	notBefore, notAfter time.Time
}

func (g boundedSerialGenerator) Next(template *x509.Certificate) (int64, error) {
	if template.NotAfter.After(g.notAfter) {
		template.NotAfter = g.notAfter
	}
	if template.NotBefore.Before(g.notBefore) {
		template.NotBefore = g.notBefore
	}
	if !template.NotBefore.Before(template.NotAfter) || !time.Now().Before(template.NotAfter) {
		return 0, fmt.Errorf("signing CA has no remaining validity; renew the CA chain with --ca")
	}
	return g.SerialGenerator.Next(template)
}

func (s *CertificateSigner) boundIssuerValidity() {
	if s.limitValidity {
		chain := s.signerConfig.Config.Certs
		generator := boundedSerialGenerator{s.signerConfig.SerialGenerator, chain[0].NotBefore, chain[0].NotAfter}
		for _, issuer := range chain[1:] {
			if issuer.NotBefore.After(generator.notBefore) {
				generator.notBefore = issuer.NotBefore
			}
			if issuer.NotAfter.Before(generator.notAfter) {
				generator.notAfter = issuer.NotAfter
			}
		}
		s.signerConfig.SerialGenerator = generator
	}
}

// RegenerateForRenewal uses the existing generation path with descendant
// validity bounded by the entire signing chain. Use it only in isolated
// staging, never on the active PKI. Complete may load expired material first;
// the bounds apply when issuing its replacements, not while loading old CAs.
func (cs *CertificateChains) RegenerateForRenewal(path ...string) error {
	if len(path) == 0 {
		return fmt.Errorf("a certificate path is required")
	}
	for _, signer := range cs.signers {
		signer.enableValidityBounds()
	}
	return cs.Regenerate(path...)
}

func (s *CertificateSigner) enableValidityBounds() {
	if s.limitValidity {
		return
	}
	s.limitValidity = true
	s.boundIssuerValidity()
	for _, child := range s.subCAs {
		child.enableValidityBounds()
	}
}

// CertificateRenewalPlanEntry describes one validated replacement. It contains
// public metadata only; private keys are never retained in a plan.
type CertificateRenewalPlanEntry struct {
	CertificateInventoryEntry

	NewNotAfter time.Time
}

// InventoryReadError distinguishes unreadable material from renewal validation
// failures without exposing private key or PEM contents to callers.
type InventoryReadError struct{ Err error }

func (e *InventoryReadError) Error() string { return e.Err.Error() }
func (e *InventoryReadError) Unwrap() error { return e.Err }

// PlanRenewal validates existing material without creating or modifying files.
// Leaf renewal includes serving, client, and peer certificates. Renewing CAs
// selects every CA and descendant. Descendant expiry is bounded by its signer.
func (cs *certificateChains) PlanRenewal(renewCAs bool, now time.Time) ([]CertificateRenewalPlanEntry, error) {
	inventory, err := cs.LoadInventory()
	if err != nil {
		return nil, &InventoryReadError{err}
	}
	entries := make(map[string]CertificateInventoryEntry, len(inventory))
	for _, entry := range inventory {
		entries[strings.Join(entry.Path, "/")] = entry
	}
	plan := make([]CertificateRenewalPlanEntry, 0, len(inventory))
	for _, builder := range cs.signers {
		signer, ok := builder.(*certificateSigner)
		if !ok {
			return nil, fmt.Errorf("unsupported certificate signer builder %T", builder)
		}
		if err := signer.planRenewal(entries, &plan, []string{signer.Name()}, nil, time.Time{}, renewCAs, now); err != nil {
			return nil, err
		}
	}
	if len(plan) == 0 {
		return nil, fmt.Errorf("no managed certificates selected for renewal")
	}
	sort.Slice(plan, func(i, j int) bool {
		if plan[i].Service != plan[j].Service {
			return plan[i].Service < plan[j].Service
		}
		if plan[i].Name != plan[j].Name {
			return plan[i].Name < plan[j].Name
		}
		return strings.Join(plan[i].Path, "/") < strings.Join(plan[j].Path, "/")
	})
	return plan, nil
}

func (s *certificateSigner) planRenewal(entries map[string]CertificateInventoryEntry, plan *[]CertificateRenewalPlanEntry,
	path []string, parent *x509.Certificate, parentExpiry time.Time, renewCAs bool, now time.Time,
) error {
	entry := entries[strings.Join(path, "/")]
	if err := validateRenewalMaterial(entry, cryptomaterial.CACertPath(s.signerDir), cryptomaterial.CAKeyPath(s.signerDir), parent); err != nil {
		return err
	}
	expiry := entry.Certificate.NotAfter
	if renewCAs {
		expiry = renewalExpiry(now, s.signerValidity, parentExpiry)
		if !now.Before(expiry) {
			return fmt.Errorf("CA %q has no valid renewal interval", entry.Name)
		}
		*plan = append(*plan, CertificateRenewalPlanEntry{entry, expiry})
	} else if now.Before(entry.Certificate.NotBefore) || !now.Before(expiry) {
		return fmt.Errorf("CA %q is not currently valid; use --ca to renew the CA chain", entry.Name)
	}
	// Legacy intermediate CAs can outlive an ancestor. Leaf renewal leaves
	// those CAs untouched but new leaves must fit the entire signing chain.
	if !parentExpiry.IsZero() && expiry.After(parentExpiry) {
		expiry = parentExpiry
	}
	for _, builder := range s.subCAs {
		subCA, ok := builder.(*certificateSigner)
		if !ok {
			return fmt.Errorf("unsupported certificate signer builder %T", builder)
		}
		subPath := append(append([]string(nil), path...), subCA.Name())
		if err := subCA.planRenewal(entries, plan, subPath, &entry.Certificate, expiry, renewCAs, now); err != nil {
			return err
		}
	}
	for _, info := range s.certificatesToSign {
		leaf := entries[strings.Join(append(append([]string(nil), path...), info.GetMeta().Name), "/")]
		certPath, err := certificatePath(s.signerDir, info)
		if err != nil {
			return err
		}
		keyPath := strings.TrimSuffix(certPath, ".crt") + ".key"
		if err := validateRenewalMaterial(leaf, certPath, keyPath, &entry.Certificate); err != nil {
			return err
		}
		newExpiry := renewalExpiry(now, info.GetMeta().Validity, expiry)
		if !now.Before(newExpiry) {
			return fmt.Errorf("certificate %q has no valid renewal interval", leaf.Name)
		}
		*plan = append(*plan, CertificateRenewalPlanEntry{leaf, newExpiry})
	}
	return nil
}

func renewalExpiry(now time.Time, validity time.Duration, parentExpiry time.Time) time.Time {
	expiry := now.Add(validity).UTC().Truncate(time.Second)
	if !parentExpiry.IsZero() && expiry.After(parentExpiry) {
		return parentExpiry.UTC()
	}
	return expiry
}

func validateRenewalMaterial(entry CertificateInventoryEntry, certPath, keyPath string, parent *x509.Certificate) error {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return &InventoryReadError{fmt.Errorf("cannot read certificate %q", entry.Name)}
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return &InventoryReadError{fmt.Errorf("cannot read private key for certificate %q", entry.Name)}
	}
	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return fmt.Errorf("invalid certificate/key pair for %q", entry.Name)
	}
	if parent == nil {
		parent = &entry.Certificate
	}
	if err := entry.Certificate.CheckSignatureFrom(parent); err != nil {
		return fmt.Errorf("invalid signing chain for certificate %q", entry.Name)
	}
	return nil
}
