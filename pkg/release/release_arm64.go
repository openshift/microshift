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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:fe65a036a65af078f6f61017ae96e141dbb203f3602ecaca7f63ec8f58a1f6c6",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:b5b3d024b2586bd0bf7b1315b2866f36a9b8b0acd23f0a9c6459371234dc8429",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:349e73813f432203920ae9ed04fc33a4026507e26ecc23ff2ab609d5b95b4206",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:c19226019fe605b5ab10496fb0b7cb4712cb694a7ee1e26642d63d515ca6b7cc",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:3f781a07e59d164eba065dba7d8e7661ab2494b21199c379b65b0ff514a1b8d0",
		"ovn_kubernetes_microshift": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:8634637031e8b7eea5cb2d5f3bb76c0f40d6328871839eae8f215c3e524d9ab7",
		"pod":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:9f9b6f1aa2770195c726ef4186306813b3c3c038cc488ff05aa5873297de94d0",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2fe468f25881e7b5ae8118c7d54b41a7fbb132a186f0156bbe46df0fd6a2f1f8",
	}
}
