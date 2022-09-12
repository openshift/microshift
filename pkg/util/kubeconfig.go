/*
Copyright © 2021 MicroShift Contributors

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
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// KubeConfigWithClientCerts creates a kubeconfig authenticating with client cert/key
// at a location provided by `path`
func KubeConfigWithClientCerts(
	path string,
	clusterURL string,
	clusterTrustBundle []byte,
	clientCertPEM []byte,
	clientKeyPEM []byte,
) error {
	const microshiftName = "microshift"

	cluster := clientcmdapi.NewCluster()
	cluster.Server = clusterURL
	cluster.CertificateAuthorityData = clusterTrustBundle

	msContext := clientcmdapi.NewContext()
	msContext.Cluster = microshiftName
	msContext.Namespace = "default"
	msContext.AuthInfo = "user"

	msUser := clientcmdapi.NewAuthInfo()
	msUser.ClientCertificateData = clientCertPEM
	msUser.ClientKeyData = clientKeyPEM

	kubeConfig := clientcmdapi.Config{
		CurrentContext: microshiftName,
		Clusters:       map[string]*clientcmdapi.Cluster{microshiftName: cluster},
		Contexts:       map[string]*clientcmdapi.Context{microshiftName: msContext},
		AuthInfos:      map[string]*clientcmdapi.AuthInfo{"user": msUser},
	}

	return clientcmd.WriteToFile(kubeConfig, path)
}
