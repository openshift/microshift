package util

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/openshift/library-go/pkg/crypto"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	rootCA *crypto.CA
)

const (
	defaultDurationDays = 365
	defaultDuration     = defaultDurationDays * 24 * time.Hour
	defaultHostname     = "localhost"
)

func BuildCA(hostname string, duration time.Duration) (*crypto.CA, error) {
	rootCaConfig, err := crypto.MakeSelfSignedCAConfigForDuration(hostname, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create root-signer CA: %v", err)
	}
	ca := &crypto.CA{
		Config:          rootCaConfig,
		SerialGenerator: &crypto.RandomSerialGenerator{},
	}
	return ca, nil
}

// GenCerts creates certs and keys
// GenCerts("/var/lib/openshift/service-ca/key", "tls.crt", "tls.key", "example.com")
func GenCerts(dir, certFilename, keyFilename string, svcName ...string) error {
	var err error
	if rootCA == nil {
		rootCA, err = BuildCA(defaultHostname, defaultDuration)
		if err != nil {
			return err
		}
	}
	certPath := filepath.Join(dir, certFilename)
	keyPath := filepath.Join(dir, keyFilename)
	s := sets.NewString(svcName...)
	_, err = rootCA.MakeAndWriteServerCert(certPath, keyPath, s, defaultDurationDays)
	return err
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
