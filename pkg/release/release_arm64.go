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

package release

// For the amd64 architecture we use the existing and tested and
// published OCP or other component upstream images

func init() {
	Image = map[string]string{
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:cfcd14ed3a3bdfcf094646b6ca1b9ebd4ca705a10e280dba8a959b70dbb1e1d6",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:a1ab5409d46f7aadd2bcbcdcf1d6f45a360964f9805423479908630d53f1731c",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:95ad731ee6c16477809a1ef752aa260bac0e9f35133b141aeb3c4f3886864449",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f1828611410f0d6f6e26ba7ed1fbbb9ffa623b8d2bb1d24f3a0d58983ab088cd",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:3f781a07e59d164eba065dba7d8e7661ab2494b21199c379b65b0ff514a1b8d0",
		"ovn_kubernetes_microshift": "quay.io/microshift/ovn-kubernetes-singlenode@sha256:012e743363b5f15f442c238099d35a0c70343fd1d4dc15b0a57a7340a338ffdb",
		"pause":                     "k8s.gcr.io/pause:3.6",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:8efa0863945884ae8c7103b02f69c04b0060786041ccf96f90bf5884527d2bc4",
	}
}
