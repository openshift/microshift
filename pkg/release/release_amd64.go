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

package release

// For the amd64 architecture we use the existing and tested and
// published OCP or other component upstream images

func init() {
	Image = map[string]string{
		"cli":                           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:4dcacb7ae5c5a5c0cb2fbc8e132b14a72ee937b3dd7d201c7a66ded05e5d6790",
		"coredns":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:6e963275041b08eafa47fa119287e1d4428b5af642c3befc6f0c90900d3516a4",
		"haproxy_router":                "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:608a8c9f893b3fae358028b44238f3184af2f3bd5cde844b8b2688583b8468ea",
		"kube_flannel":                  "quay.io/coreos/flannel:v0.14.0",
		"kube_flannel_cni":              "quay.io/microshift/flannel-cni:v0.14.0",
		"kube_rbac_proxy":               "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:08e8b4004edaeeb125ced09ab2c4cd6d690afaf3a86309c91a994dec8e3ccbf3",
		"kubevirt_hostpath_provisioner": "quay.io/kubevirt/hostpath-provisioner:v0.8.0",
		"pause":                         "k8s.gcr.io/pause:3.6",
		"service_ca_operator":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:94a4cd55cb067f6d3747843b5532114806b6c81e8421e370bfe92ccafcc6bca7",
	}
}
