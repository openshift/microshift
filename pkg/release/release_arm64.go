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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:1987b9d295ee9b7488750861ba85c761923237496547a1c847b835165a03a586",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:1b7be165834025aed05f874e83f7393a434f1e4e52ae7814941a67818d6a529f",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:7d39cd15a32d083e3afcda7b836df17c17fa20ad4dc17f639834c7b8ffa41028",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f4d4b35f8e4a9bf073ea40e71bb53c063e29cc9ab30fcbcb54e03a29781a6cd6",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:3f781a07e59d164eba065dba7d8e7661ab2494b21199c379b65b0ff514a1b8d0",
		"ovn_kubernetes_microshift": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:63a74084d869365e562fe5bed75ca0ce897436e2506c9ab22dc83fb90283ed67",
		"pod":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:df143bf90b526df87adec0328ad94daaa5877e055eba5ec3a044f03719798dad",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:35cca3568753a576a93197d58fba68e5ab9a2c864fb330153b734d41733cffa6",
	}
}
