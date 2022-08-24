/*
Copyright © 2021 Microshift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"os"
	"path/filepath"
	"text/template"
)

// Kubeconfig creates a kubeconfig
func Kubeconfig(path string, clusterTrustBundle []byte, common string, svcName []string, clusterURL string) error {
	cert, key, err := GenCertsBuff(common, svcName)
	if err != nil {
		return err
	}

	return KubeConfigWithClientCerts(path, clusterURL, clusterTrustBundle, cert, key)
}

func KubeConfigWithClientCerts(path string, clusterURL string, clusterTrustBundle []byte, clientCertPEM []byte, clientKeyPEM []byte) error {
	kubeconfigTemplate := template.Must(template.New("kubeconfig").Parse(`
apiVersion: v1
kind: Config
current-context: microshift
preferences: {}
contexts:
- context:
    cluster: microshift
    namespace: default
    user: user
  name: microshift
clusters:
- cluster:
    server: {{.ClusterURL}}
    certificate-authority-data: {{.ClusterCA}}
  name: microshift
users:
- name: user
  user:
    client-certificate-data: {{.ClientCert}}
    client-key-data: {{.ClientKey}}
`))

	data := struct {
		ClusterURL string
		ClusterCA  string
		ClientCert string
		ClientKey  string
	}{
		ClusterURL: clusterURL,
		ClusterCA:  Base64(clusterTrustBundle),
		ClientCert: Base64(clientCertPEM),
		ClientKey:  Base64(clientKeyPEM),
	}
	os.MkdirAll(filepath.Dir(path), os.FileMode(0700))

	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	return kubeconfigTemplate.Execute(output, &data)
}
