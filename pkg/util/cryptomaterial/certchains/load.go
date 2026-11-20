package certchains

import (
	"fmt"
	"strings"

	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Load reads existing chains without creating, repairing or reissuing material.
// Use after ValidateRenewal when activating a prepared generation: Complete's
// Ensure methods may reissue certificates instead of retaining the saved set.
func (cs *certificateChains) Load() (*CertificateChains, error) {
	chains := &CertificateChains{signers: make(map[string]*CertificateSigner, len(cs.signers))}
	for _, builder := range cs.signers {
		signer, err := loadSigner(builder, false)
		if err != nil {
			return nil, err
		}
		chains.signers[signer.signerName] = signer
	}
	for bundle, paths := range cs.fileBundles {
		for _, path := range paths {
			signer := chains.GetSigner(path...)
			if signer == nil {
				return nil, fmt.Errorf("unknown bundle signer %v", path)
			}
			signer.caBundlePaths.Insert(bundle)
		}
	}
	return chains, nil
}

func loadSigner(builder CertificateSignerBuilder, intermediate bool) (*CertificateSigner, error) {
	s, ok := builder.(*certificateSigner)
	if !ok {
		return nil, fmt.Errorf("unsupported certificate signer builder %T", builder)
	}
	certPath := cryptomaterial.CACertPath(s.signerDir)
	if intermediate {
		certPath = cryptomaterial.CABundlePath(s.signerDir)
	}
	// Consumption does not need an issuance counter. Passing its file would let
	// GetCA repair a zero serial, violating this loader's read-only contract.
	ca, err := crypto.GetCA(certPath, cryptomaterial.CAKeyPath(s.signerDir), "")
	if err != nil {
		return nil, fmt.Errorf("cannot load signer %q: %w", s.signerName, err)
	}
	loaded := &CertificateSigner{
		signerName: s.signerName, service: s.service, signerDir: s.signerDir,
		signerValidity: s.signerValidity, signerConfig: ca,
		subCAs:             make(map[string]*CertificateSigner, len(s.subCAs)),
		signedCertificates: make(map[string]*signedCertificateInfo, len(s.certificatesToSign)),
		caBundlePaths:      sets.New(s.caBundlePaths...),
	}
	for _, child := range s.subCAs {
		subCA, err := loadSigner(child, true)
		if err != nil {
			return nil, err
		}
		loaded.subCAs[subCA.signerName] = subCA
	}
	for _, info := range s.certificatesToSign {
		path, err := certificatePath(s.signerDir, info)
		if err != nil {
			return nil, err
		}
		material, err := crypto.GetTLSCertificateConfig(path, strings.TrimSuffix(path, ".crt")+".key")
		if err != nil {
			return nil, fmt.Errorf("cannot load certificate %q: %w", info.GetMeta().Name, err)
		}
		loaded.signedCertificates[info.GetMeta().Name] = &signedCertificateInfo{CSRInfo: info, tlsConfig: material}
	}
	return loaded, nil
}
