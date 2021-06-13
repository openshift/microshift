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
func Kubeconfig(path, common string, svcName []string, clusterURL string) error {
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
	certBuff, keyBuff, err := GenCertsBuff(common, svcName)
	if err != nil {
		return err
	}
	clusterCA := Base64(CertToPem(GetRootCA()))
	clientCert := Base64(certBuff)
	clientKey := Base64(keyBuff)
	data := struct {
		ClusterURL string
		ClusterCA  string
		ClientCert string
		ClientKey  string
	}{
		ClusterURL: clusterURL,
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
