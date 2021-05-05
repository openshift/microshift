package util

import (
	"os"
	"path/filepath"
	"text/template"
)

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

	os.MkdirAll(dir, 0700)
	path := filepath.Join(dir, filename)

	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	return kubeconfigTemplate.Execute(output, &data)
}
