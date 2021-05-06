package util

import (
	"os"
	"path/filepath"
	"text/template"
)

// Kubeconfig creates a kubeconfig
func Kubeconfig(path, endpoint string) error {
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
    server: ${Endpoint}
    certificate-authority-data: ${ClusterCA}
  name: ushift
users:
- name: ushift
  user:
    client-certificate-data: ${ClientCert}
    client-key-data: ${ClientKey}
`))
	clusterCA := Base64(CertToPem(rootCA))
	clientCert := ""
	clientKey := ""
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
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))

	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	return kubeconfigTemplate.Execute(output, &data)
}
