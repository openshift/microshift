package util

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/openshift/library-go/pkg/crypto"
)

// GenCerts creates TLS or CA bundles
// GenCerts("service-ca", "/var/lib/openshift/service-ca/key", "tls.crt", "tls.key")
// GenCerts("service-ca-signer", "/var/lib/openshift/service-ca/ca-cabundle", "ca-bundle.crt", "ca-bundle.key")
func GenCerts(svcName string, dir, certFilename, keyFilename string) (string, error) {
	ca, err := crypto.MakeSelfSignedCAConfig(svcName, 790)
	if err != nil {
		return "", err
	}

	certBuff := &bytes.Buffer{}
	keyBuff := &bytes.Buffer{}
	if err := ca.WriteCertConfig(certBuff, keyBuff); err != nil {
		return "", err
	}
	os.MkdirAll(dir, 0700)
	certPath := filepath.Join(dir, certFilename)
	ioutil.WriteFile(certPath, certBuff.Bytes(), 0644)
	keyPath := filepath.Join(dir, keyFilename)
	ioutil.WriteFile(keyPath, keyBuff.Bytes(), 0644)

	return dir, nil
}
