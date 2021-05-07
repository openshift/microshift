package util

import (
	"os"
	"path/filepath"
	"text/template"
)

// Kubeconfig creates a kubeconfig
func Kubeconfig(path string, svcName []string) error {
	kubeconfigTemplate := template.Must(template.New("kubeconfig").Parse(`
apiVersion: v1
kind: Config
preferences:
  colors: true
current-context: ushift
contexts:
- context:
    cluster: ushift
    namespace: default
    user: ushift
  name: ushift
clusters:
- cluster:
    server: https://127.0.0.1:6443
    certificate-authority-data: {{.ClusterCA}}
  name: ushift
users:
- name: ushift
  user:
    client-certificate-data: {{.ClientCert}}
    client-key-data: {{.ClientKey}}
`))
	certBuff, keyBuff, err := GenCertsBuff(svcName)
	if err != nil {
		return err
	}
	clusterCA := Base64(CertToPem(GetRootCA()))
	clientCert := Base64(certBuff)
	clientKey := Base64(keyBuff)
	data := struct {
		ClusterCA  string
		ClientCert string
		ClientKey  string
	}{
		ClusterCA:  clusterCA,
		ClientCert: clientCert,
		ClientKey:  clientKey,
	}
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))

	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	return kubeconfigTemplate.Execute(output, &data)
}
