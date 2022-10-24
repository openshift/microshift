package certchains

import (
	"crypto/x509"
	"fmt"
	"time"
)

type CertificateChains struct {
	signers map[string]*CertificateSigner
}

func (cs *CertificateChains) GetSignerNames() []string {
	return certificateSignersMapKeysOrdered(cs.signers)
}

func (cs *CertificateChains) GetSigner(signerPath ...string) *CertificateSigner {
	if len(signerPath) == 0 {
		return nil
	}

	currentSigner := cs.signers[signerPath[0]]
	for _, fragment := range signerPath[1:] {
		if currentSigner != nil {
			currentSigner = currentSigner.GetSubCA(fragment)
		} else {
			return nil
		}
	}

	return currentSigner
}

func (cs *CertificateChains) GetCertKey(certPath ...string) ([]byte, []byte, error) {
	if len(certPath) == 0 {
		return nil, nil, fmt.Errorf("empty certificate path")
	}
	if len(certPath) == 1 {
		return nil, nil, fmt.Errorf("the CertificateChains struct only stores signers, the path must be at least 1 level deep")
	}

	signerPath := certPath[:len(certPath)-1]
	signer := cs.GetSigner(signerPath...)
	if signer == nil {
		return nil, nil, fmt.Errorf("no such signer in the path: %v", signerPath)
	}

	return signer.GetCertKey(certPath[len(certPath)-1])
}

type CertWalkFunc func(certPath []string, c x509.Certificate) error

// WalkChains traverses through the trust chain starting at `rootPath` and applies
// `fn` on all the certificates in the chain tree
func (cs *CertificateChains) WalkChains(rootPath []string, fn CertWalkFunc) error {
	if len(rootPath) == 0 {
		for _, signerName := range cs.GetSignerNames() {
			if err := cs.WalkChains([]string{signerName}, fn); err != nil {
				return err
			}
		}
		return nil
	}

	if signer := cs.GetSigner(rootPath...); signer != nil {
		// the path points to a signer
		if err := fn(rootPath, *signer.signerConfig.Config.Certs[0]); err != nil {
			return fmt.Errorf("failed to execute walk function on %v: %v", rootPath, err)
		}

		nextNames := append(signer.GetSubCANames(), signer.GetCertNames()...)
		for _, name := range nextNames {
			if err := cs.WalkChains(append(rootPath, name), fn); err != nil {
				return err
			}
		}
		return nil

	} else if len(rootPath) == 1 {
		// the path is a single element but no such signer exists
		return fmt.Errorf("%v is not a path to a signer", rootPath)
	} else {
		// the path points to a leaf certificate
		signerPath := rootPath[:len(rootPath)-1]
		if signer := cs.GetSigner(signerPath...); signer != nil {
			cert := signer.signedCertificates[rootPath[len(rootPath)-1]]
			if cert == nil {
				return fmt.Errorf("the requested element does not exist")
			}
			return fn(rootPath, *cert.tlsConfig.Certs[0])
		}

		return fmt.Errorf("a non-leaf fragment of the path '%v' either is not a signer or it doesn't exist", rootPath)
	}
}

func WhenToRotateAtEarliest(cs *CertificateChains) ([]string, time.Time, error) {
	const rotateAtLifetime = 0.7
	var (
		certPath     []string
		rotationDate time.Time
	)

	err := cs.WalkChains(nil, func(currentPath []string, c x509.Certificate) error {
		totalTime := c.NotAfter.Sub(c.NotBefore).Seconds()
		rotateAt := c.NotBefore.Add(time.Duration(totalTime*rotateAtLifetime) * time.Second)

		if rotationDate.IsZero() {
			rotationDate = rotateAt
			certPath = currentPath
			return nil
		}

		if rotateAt.Before(rotationDate) {
			rotationDate = rotateAt
			certPath = currentPath
		}

		return nil
	})

	return certPath, rotationDate, err
}
