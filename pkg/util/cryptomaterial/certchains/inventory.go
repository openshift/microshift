package certchains

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

// CertificateRole identifies how a certificate is used by MicroShift.
type CertificateRole string

const (
	CertificateRoleUnknown CertificateRole = "unknown"
	CertificateRoleCA      CertificateRole = "ca"
	CertificateRoleClient  CertificateRole = "client"
	CertificateRoleServing CertificateRole = "serving"
	CertificateRolePeer    CertificateRole = "peer"
)

// RotationPolicy identifies the thresholds used to classify certificate zones.
type RotationPolicy string

const (
	RotationPolicyUnknown  RotationPolicy = "unknown"
	RotationPolicyStandard RotationPolicy = "standard"
	RotationPolicyExtended RotationPolicy = "extended"
)

// CertificateZone identifies the current renewal urgency of a certificate.
type CertificateZone string

const (
	CertificateZoneGreen  CertificateZone = "green"
	CertificateZoneYellow CertificateZone = "yellow"
	CertificateZoneRed    CertificateZone = "red"
)

// CertificateInventoryEntry describes one certificate managed by the chain.
// Path can be passed to CertificateChains methods such as Regenerate.
type CertificateInventoryEntry struct {
	Path           []string
	Name           string
	Service        string
	ParentCA       string
	Role           CertificateRole
	RotationPolicy RotationPolicy
	Certificate    x509.Certificate
}

// CertificateInventory is a snapshot of certificates managed by a chain.
type CertificateInventory []CertificateInventoryEntry

// ByRole returns the inventory entries matching role.
func (i CertificateInventory) ByRole(role CertificateRole) CertificateInventory {
	entries := make(CertificateInventory, 0)
	for _, entry := range i {
		if entry.Role == role {
			entries = append(entries, entry)
		}
	}
	return entries
}

// ZoneAt calculates the certificate's renewal zone at the provided time.
func (entry CertificateInventoryEntry) ZoneAt(now time.Time) (CertificateZone, error) {
	certificate := entry.Certificate
	if now.Before(certificate.NotBefore) || !now.Before(certificate.NotAfter) {
		return CertificateZoneRed, nil
	}

	validity := certificate.NotAfter.Sub(certificate.NotBefore)
	if validity <= 0 {
		return CertificateZoneRed, nil
	}
	remaining := certificate.NotAfter.Sub(now)
	ratio := float64(remaining) / float64(validity)

	var greenThreshold, yellowThreshold float64
	switch entry.RotationPolicy {
	case RotationPolicyStandard:
		greenThreshold, yellowThreshold = 0.583, 0.333
	case RotationPolicyExtended:
		greenThreshold, yellowThreshold = 0.15, 0.10
	case RotationPolicyUnknown:
		return "", fmt.Errorf("certificate %q has unknown rotation policy %q", entry.Name, entry.RotationPolicy)
	default:
		return "", fmt.Errorf("certificate %q has unknown rotation policy %q", entry.Name, entry.RotationPolicy)
	}

	switch {
	case ratio > greenThreshold:
		return CertificateZoneGreen, nil
	case ratio > yellowThreshold:
		return CertificateZoneYellow, nil
	default:
		return CertificateZoneRed, nil
	}
}

// Inventory returns a deterministic snapshot of all certificates managed by
// the chain. Signers are listed before their sub-CAs and leaf certificates.
func (cs *CertificateChains) Inventory() CertificateInventory {
	signerNames := cs.GetSignerNames()
	entries := make(CertificateInventory, 0, len(signerNames))
	for _, signerName := range signerNames {
		signer := cs.GetSigner(signerName)
		entries = append(entries, signer.inventory([]string{signerName})...)
	}
	return entries
}

func (s *CertificateSigner) inventory(path []string) CertificateInventory {
	entries := make(CertificateInventory, 0, 1+len(s.subCAs)+len(s.signedCertificates))
	entries = append(entries, CertificateInventoryEntry{
		Path:           append([]string(nil), path...),
		Name:           s.signerName,
		Service:        s.service,
		ParentCA:       parentCA(path),
		Role:           CertificateRoleCA,
		RotationPolicy: RotationPolicyExtended,
		Certificate:    *s.signerConfig.Config.Certs[0],
	})

	for _, subCAName := range s.GetSubCANames() {
		subCAPath := append(append([]string(nil), path...), subCAName)
		entries = append(entries, s.GetSubCA(subCAName).inventory(subCAPath)...)
	}

	for _, certName := range s.GetCertNames() {
		cert := s.signedCertificates[certName]
		meta := cert.GetMeta()
		certPath := append(append([]string(nil), path...), certName)
		entries = append(entries, CertificateInventoryEntry{
			Path:           certPath,
			Name:           certName,
			Service:        meta.Service,
			ParentCA:       s.signerName,
			Role:           certificateRole(cert.CSRInfo),
			RotationPolicy: certificateRotationPolicy(cert.CSRInfo),
			Certificate:    *cert.tlsConfig.Certs[0],
		})
	}

	return entries
}

// LoadInventory reads the certificate inventory from disk without creating or
// modifying certificate material.
func (cs *certificateChains) LoadInventory() (CertificateInventory, error) {
	signers := append([]CertificateSignerBuilder(nil), cs.signers...)
	sort.Slice(signers, func(i, j int) bool { return signers[i].Name() < signers[j].Name() })

	entries := make(CertificateInventory, 0, len(signers))
	for _, signerBuilder := range signers {
		signer, ok := signerBuilder.(*certificateSigner)
		if !ok {
			return nil, fmt.Errorf("unsupported certificate signer builder %T", signerBuilder)
		}
		signerEntries, err := signer.loadInventory([]string{signer.Name()})
		if err != nil {
			return nil, err
		}
		entries = append(entries, signerEntries...)
	}
	return entries, nil
}

func (s *certificateSigner) loadInventory(path []string) (CertificateInventory, error) {
	certificate, err := loadCertificate(cryptomaterial.CACertPath(s.signerDir))
	if err != nil {
		return nil, fmt.Errorf("failed to load CA %q: %w", s.signerName, err)
	}

	entries := make(CertificateInventory, 0, 1+len(s.subCAs)+len(s.certificatesToSign))
	entries = append(entries, CertificateInventoryEntry{
		Path:           append([]string(nil), path...),
		Name:           s.signerName,
		Service:        s.service,
		ParentCA:       parentCA(path),
		Role:           CertificateRoleCA,
		RotationPolicy: RotationPolicyExtended,
		Certificate:    certificate,
	})

	subCAs := append([]CertificateSignerBuilder(nil), s.subCAs...)
	sort.Slice(subCAs, func(i, j int) bool { return subCAs[i].Name() < subCAs[j].Name() })
	for _, subCABuilder := range subCAs {
		subCA, ok := subCABuilder.(*certificateSigner)
		if !ok {
			return nil, fmt.Errorf("unsupported certificate signer builder %T", subCABuilder)
		}
		subCAPath := append(append([]string(nil), path...), subCA.Name())
		subCAEntries, err := subCA.loadInventory(subCAPath)
		if err != nil {
			return nil, err
		}
		entries = append(entries, subCAEntries...)
	}

	certificates := append([]CSRInfo(nil), s.certificatesToSign...)
	sort.Slice(certificates, func(i, j int) bool {
		return certificates[i].GetMeta().Name < certificates[j].GetMeta().Name
	})
	for _, info := range certificates {
		meta := info.GetMeta()
		certificateFilePath, err := certificatePath(s.signerDir, info)
		if err != nil {
			return nil, err
		}
		certificate, err := loadCertificate(certificateFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate %q: %w", meta.Name, err)
		}
		entries = append(entries, CertificateInventoryEntry{
			Path:           append(append([]string(nil), path...), meta.Name),
			Name:           meta.Name,
			Service:        meta.Service,
			ParentCA:       s.signerName,
			Role:           certificateRole(info),
			RotationPolicy: certificateRotationPolicy(info),
			Certificate:    certificate,
		})
	}

	return entries, nil
}

func loadCertificate(path string) (x509.Certificate, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return x509.Certificate{}, err
	}
	certificates, err := crypto.CertsFromPEM(contents)
	if err != nil {
		return x509.Certificate{}, err
	}
	if len(certificates) == 0 {
		return x509.Certificate{}, fmt.Errorf("no certificates found in %q", path)
	}
	return *certificates[0], nil
}

func certificatePath(signerDir string, info CSRInfo) (string, error) {
	certificateDir := filepath.Join(signerDir, info.GetMeta().Name)
	switch info.(type) {
	case *ClientCertificateSigningRequestInfo:
		return cryptomaterial.ClientCertPath(certificateDir), nil
	case *ServingCertificateSigningRequestInfo:
		return cryptomaterial.ServingCertPath(certificateDir), nil
	case *PeerCertificateSigningRequestInfo:
		return cryptomaterial.PeerCertPath(certificateDir), nil
	default:
		return "", fmt.Errorf("unsupported certificate request info %T", info)
	}
}

func parentCA(path []string) string {
	if len(path) < 2 {
		return ""
	}
	return path[len(path)-2]
}

func certificateRole(info CSRInfo) CertificateRole {
	switch info.(type) {
	case *ClientCertificateSigningRequestInfo:
		return CertificateRoleClient
	case *ServingCertificateSigningRequestInfo:
		return CertificateRoleServing
	case *PeerCertificateSigningRequestInfo:
		return CertificateRolePeer
	default:
		return CertificateRoleUnknown
	}
}

func certificateRotationPolicy(info CSRInfo) RotationPolicy {
	switch info.(type) {
	case *ServingCertificateSigningRequestInfo:
		return RotationPolicyStandard
	case *ClientCertificateSigningRequestInfo, *PeerCertificateSigningRequestInfo:
		return info.GetMeta().RotationPolicy
	default:
		return RotationPolicyUnknown
	}
}
