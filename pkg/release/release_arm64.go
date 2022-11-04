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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:6aed4b901ff2456c803ff159b50021bd86c6473c0c65f76c39269fb330640d1b",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:be116df9f46266e604f1028fbde3742e3d63fef7c6421fb5a2a3673ce4f971e7",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:4f5f51a8ba723590e550bc2024eaea6554bb4b749ec4356229060d3bd92fc9eb",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:51ce83de9a43dfe8731c532f4ed0402c70da8df1e056292190c2cb43a45232a7",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:3f781a07e59d164eba065dba7d8e7661ab2494b21199c379b65b0ff514a1b8d0",
		"ovn_kubernetes_microshift": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:071a394e5d43dbe7cb2470fc570b4d6cd607b799614d02b0ecadfe28c2c00138",
		"pod":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f310397d0a61f0ffefbb5753cb72f271d483eac82f4dd676d592fe2ea8487a30",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f41255221e5dafd5688684bc0446fedf4a7526b3a02e05ca787bd166449088f8",
	}
}
