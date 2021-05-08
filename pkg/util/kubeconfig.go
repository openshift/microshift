package util

import (
	"os"
	"path/filepath"
	"text/template"
)

// Kubeconfig creates a kubeconfig
func Kubeconfig(path, common string, svcName []string) error {
	kubeconfigTemplate := template.Must(template.New("kubeconfig").Parse(`
apiVersion: v1
kind: Config
current-context: ushift
preferences: {}
contexts:
- context:
    cluster: ushift
    namespace: default
    user: user
  name: ushift
clusters:
- cluster:
    server: https://127.0.0.1:6443
    certificate-authority-data: {{.ClusterCA}}
  name: ushift
users:
- name: user
  user:
    client-certificate-data: {{.ClientCert}}
    client-key-data: {{.ClientKey}}
`))
	certBuff, keyBuff, err := GenCertsBuff(common, svcName)
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
