package certchains

import (
	"fmt"

	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

type CertificateChainsBuilder interface {
	WithSigners(signers ...CertificateSignerBuilder) CertificateChainsBuilder
	WithCABundle(bundlePath string, signerNames ...string) CertificateChainsBuilder
	Complete() (*CertificateChains, error)
}

type certificateChains struct {
	signers []CertificateSignerBuilder

	// fileBundles maps fileName -> signers, where fileName is the filename of a CA bundle
	// where PEM certificates should be stored
	fileBundles map[string][]string
}

func NewCertificateChains(signers ...CertificateSignerBuilder) CertificateChainsBuilder {
	return &certificateChains{
		signers: signers,

		fileBundles: make(map[string][]string),
	}
}

func (cs *certificateChains) WithSigners(signers ...CertificateSignerBuilder) CertificateChainsBuilder {
	cs.signers = append(cs.signers, signers...)
	return cs
}

func (cs *certificateChains) WithCABundle(bundlePath string, signerNames ...string) CertificateChainsBuilder {
	cs.fileBundles[bundlePath] = signerNames
	return cs
}

func (cs *certificateChains) Complete() (*CertificateChains, error) {
	completeChains := &CertificateChains{
		signers: make(map[string]*CertificateSigner),
	}

	for _, signer := range cs.signers {
		signer := signer
		if _, ok := completeChains.signers[signer.Name()]; ok {
			return nil, fmt.Errorf("signer name clash: %s", signer.Name())
		}

		completedSigner, err := signer.Complete()
		if err != nil {
			return nil, fmt.Errorf("failed to complete signer %q: %w", signer.Name(), err)
		}
		completeChains.signers[completedSigner.signerName] = completedSigner
	}

	bundlePreWrite := make(map[string][]byte, len(cs.fileBundles))
	for bundlePath, signers := range cs.fileBundles {
		for _, s := range signers {
			signerCACertPEM, err := completeChains.GetSigner(s).GetSignerCertPEM()
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve cert PEM for signer %q: %w", s, err)
			}
			bundlePreWrite[bundlePath] = append(bundlePreWrite[bundlePath], append(signerCACertPEM, byte('\n'))...)
		}
	}

	for bundlePath, pemChain := range bundlePreWrite {
		if err := cryptomaterial.AppendCertsToFile(bundlePath, pemChain); err != nil {
			return nil, err
		}
	}

	return completeChains, nil
}
