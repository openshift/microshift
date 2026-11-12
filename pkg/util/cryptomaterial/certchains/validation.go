package certchains

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openshift/library-go/pkg/crypto"
)

// ValidateRenewal reloads staged or committed files; it never repairs them.
func (cs *certificateChains) ValidateRenewal(renewCAs bool, now time.Time) (CertificateInventory, error) {
	if _, err := cs.PlanRenewal(false, now); err != nil {
		return nil, err
	}
	inventory, err := cs.LoadInventory()
	if err != nil {
		return nil, err
	}
	byPath := make(map[string]CertificateInventoryEntry, len(inventory))
	for _, entry := range inventory {
		byPath[strings.Join(entry.Path, "/")] = entry
	}
	chainExpiry := make(map[string]time.Time, len(inventory))
	// LoadInventory visits parents first, allowing an effective expiry to be
	// propagated through legacy CAs without requiring those CAs to be changed.
	for _, entry := range inventory {
		cert := entry.Certificate
		if !now.Before(cert.NotAfter) {
			return nil, fmt.Errorf("renewed certificate %q is not currently valid: %w", entry.Name, ErrCertificateExpired)
		}
		if now.Before(cert.NotBefore) {
			return nil, fmt.Errorf("renewed certificate %q is not currently valid", entry.Name)
		}
		expiry := cert.NotAfter
		if len(entry.Path) > 1 {
			parentExpiry := chainExpiry[strings.Join(entry.Path[:len(entry.Path)-1], "/")]
			if cert.NotAfter.After(parentExpiry) && (renewCAs || entry.Role != CertificateRoleCA) {
				return nil, fmt.Errorf("certificate %q expires after its signing CA", entry.Name)
			}
			if expiry.After(parentExpiry) {
				expiry = parentExpiry
			}
		}
		chainExpiry[strings.Join(entry.Path, "/")] = expiry
	}
	for path, signers := range cs.fileBundles {
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		certs, err := crypto.CertsFromPEM(contents)
		if err != nil {
			return nil, fmt.Errorf("invalid CA bundle %q", path)
		}
		if len(certs) != len(signers) {
			return nil, fmt.Errorf("CA bundle %q has an unexpected certificate count", path)
		}
		for _, signer := range signers {
			expected := byPath[strings.Join(signer, "/")].Certificate
			found := false
			for _, cert := range certs {
				if bytes.Equal(cert.Raw, expected.Raw) {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("CA bundle %q does not contain the current signer %q", path, strings.Join(signer, "/"))
			}
		}
	}
	return inventory, nil
}
