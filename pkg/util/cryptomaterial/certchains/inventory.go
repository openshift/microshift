package certchains

import "crypto/x509"

// CertificateRole identifies how a certificate is used by MicroShift.
type CertificateRole string

const (
	CertificateRoleUnknown CertificateRole = "unknown"
	CertificateRoleCA      CertificateRole = "ca"
	CertificateRoleClient  CertificateRole = "client"
	CertificateRoleServing CertificateRole = "serving"
	CertificateRolePeer    CertificateRole = "peer"
)

// CertificateInventoryEntry describes one certificate managed by the chain.
// Path can be passed to CertificateChains methods such as Regenerate.
type CertificateInventoryEntry struct {
	Path        []string
	Role        CertificateRole
	Certificate x509.Certificate
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
		Path:        append([]string(nil), path...),
		Role:        CertificateRoleCA,
		Certificate: *s.signerConfig.Config.Certs[0],
	})

	for _, subCAName := range s.GetSubCANames() {
		subCAPath := append(append([]string(nil), path...), subCAName)
		entries = append(entries, s.GetSubCA(subCAName).inventory(subCAPath)...)
	}

	for _, certName := range s.GetCertNames() {
		cert := s.signedCertificates[certName]
		certPath := append(append([]string(nil), path...), certName)
		entries = append(entries, CertificateInventoryEntry{
			Path:        certPath,
			Role:        certificateRole(cert.CSRInfo),
			Certificate: *cert.tlsConfig.Certs[0],
		})
	}

	return entries
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
