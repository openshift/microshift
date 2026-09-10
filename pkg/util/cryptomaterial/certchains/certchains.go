package certchains

import (
	"fmt"
	"time"

	"k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/util/cryptomaterial"
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

func (cs *CertificateChains) Regenerate(certPath ...string) error {
	if signer := cs.GetSigner(certPath[0]); signer != nil {
		return signer.Regenerate(certPath[1:]...)
	}

	return fmt.Errorf("no such signer: %s", certPath[0])
}

func WhenToRotateAtEarliest(cs *CertificateChains) ([]string, time.Time, error) {
	var (
		certPath     []string
		rotationDate time.Time
	)

	for _, entry := range cs.Inventory() {
		currentPath := entry.Path
		c := entry.Certificate
		const month = 30 * time.Hour * 24

		rotateAt := c.NotAfter.Add(-4 * month)
		if !cryptomaterial.IsCertShortLived(&c) {
			rotateAt = c.NotAfter.Add(-12 * month)
		}
		klog.Errorf("%v rotate at: %s", currentPath, rotateAt.String())

		if rotationDate.IsZero() {
			rotationDate = rotateAt
			certPath = currentPath
			continue
		}

		if rotateAt.Before(rotationDate) {
			rotationDate = rotateAt
			certPath = currentPath
		}
	}

	return certPath, rotationDate, nil
}
