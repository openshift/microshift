package util

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

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

// Kubeconfig creates a kubeconfig
func Kubeconfig(dir, filename, endpoint, clusterCA string) error {
	kubeconfigTemplate := template.Must(template.New("kubeconfig").Parse(`
apiVersion: v1
kind: Config
preferences:
  colors: true
current-context: ushift-ctx
contexts:
- context:
    cluster: ushift
    namespace: default
    user: ushift
  name: ushift
clusters:
- cluster:
    server: ${Endpoint}
    certificate-authority-data: ${ClusterCA}
  name: ushift
users:
- name: ushift
  user:
    client-certificate-data: ${ClientCert}
    client-key-data: ${ClientKey}
`))
	ca, err := crypto.MakeSelfSignedCAConfig("kubeconfig", 790)
	if err != nil {
		return err
	}

	certBuff := &bytes.Buffer{}
	keyBuff := &bytes.Buffer{}
	if err := ca.WriteCertConfig(certBuff, keyBuff); err != nil {
		return err
	}
	clientCert := certBuff.String()
	clientKey := keyBuff.String()

	data := struct {
		Endpoint   string
		ClusterCA  string
		ClientCert string
		ClientKey  string
	}{
		Endpoint:   endpoint,
		ClusterCA:  clusterCA,
		ClientCert: clientCert,
		ClientKey:  clientKey,
	}

	os.MkdirAll(dir, 0700)
	path := filepath.Join(dir, filename)

	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	return kubeconfigTemplate.Execute(output, &data)
}
