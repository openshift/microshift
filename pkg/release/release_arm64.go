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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2ccab762a25d6634644f4c3a4c69fc5fb7d31b0560c40e8ba865cddbbbc9bf4f",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:75b085502defd8582f6db6fb82114773dc703614aecd9c50e2d73f6582cbe356",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:4f75d6113004380f47b8ae1c49830fa5ab158f224efd6260e2b882698aa7acaa",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:76717cfcca6f60fc279bf9b0e9ee202020b22d54838aba18a35b98bf007b030b",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:3f781a07e59d164eba065dba7d8e7661ab2494b21199c379b65b0ff514a1b8d0",
		"ovn_kubernetes_microshift": "quay.io/microshift/ovn-kubernetes-singlenode@sha256:012e743363b5f15f442c238099d35a0c70343fd1d4dc15b0a57a7340a338ffdb",
		"pause":                     "k8s.gcr.io/pause:3.6",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:3386be5a4ebbc22fa27445279a5ff9c04042dc14fc31a8a319d245cd25f2f68a",
	}
}
