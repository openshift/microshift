package certchains

import (
	"fmt"
	"os"
)

type CertificateChainsBuilder interface {
	WithSigners(signers ...CertificateSignerBuilder) CertificateChainsBuilder
	WithCABundle(bundlePath string, signerNames ...[]string) CertificateChainsBuilder
	Complete() (*CertificateChains, error)
}

type certificateChains struct {
	signers []CertificateSignerBuilder

	// fileBundles maps fileName -> signers, where fileName is the filename of a CA bundle
	// where PEM certificates should be stored
	fileBundles map[string][][]string
}

func NewCertificateChains(signers ...CertificateSignerBuilder) CertificateChainsBuilder {
	return &certificateChains{
		signers: signers,

		fileBundles: make(map[string][][]string),
	}
}

func (cs *certificateChains) WithSigners(signers ...CertificateSignerBuilder) CertificateChainsBuilder {
	cs.signers = append(cs.signers, signers...)
	return cs
}

func (cs *certificateChains) WithCABundle(bundlePath string, signerNames ...[]string) CertificateChainsBuilder {
	cs.fileBundles[bundlePath] = signerNames
	return cs
}

func (cs *certificateChains) Complete() (*CertificateChains, error) {
	completeChains := &CertificateChains{
		signers: make(map[string]*CertificateSigner),
	}

	// Library-go crypto package warns via stderr prints about CA
	// and cert validity time when they exceed 5 and 2 years
	// respectively. This is not configurable and the introduction
	// of such a possibility involves changing the API in a massively
	// used library across OpenShift. Temporarily disable stderr as
	// a shortcut to clean logs.
	newstderr, err := os.Open("/dev/null")
	if err == nil {
		originalStderr := os.Stderr
		os.Stderr = newstderr
		defer func() { _ = newstderr.Close() }()
		defer func() {
			os.Stderr = originalStderr
		}()
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

	for bundle, signers := range cs.fileBundles {
		for _, s := range signers {
			signerObj := completeChains.GetSigner(s...)
			if signerObj == nil {
				return nil, NewSignerNotFound(signerObj.signerName)
			}

			if err := signerObj.AddToBundles(bundle); err != nil {
				return nil, fmt.Errorf("failed adding the signer %q to CA bundle %q: %v", signerObj.signerName, bundle, err)
			}
		}
	}

	return completeChains, nil
}
