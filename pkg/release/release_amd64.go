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
		"cli":                           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:ba44dead03ea74107f90d58525106fb52d27a120b73c6cc8e2be31d37043ca1c",
		"coredns":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2b423e88cdd37f307aff51cbb0f53fc45deff9618f5b4f12bfb78bea7aff51a2",
		"haproxy_router":                "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:5f9ee3afa744e790dbb61d08f44e30370c9a5ff041054bf99dc1afe58792cd7b",
		"kube_flannel":                  "quay.io/coreos/flannel:v0.14.0",
		"kube_flannel_cni":              "quay.io/microshift/flannel-cni:v0.14.0",
		"kube_rbac_proxy":               "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:3df935837634adb5df8080ac7263c7fb4c4f9d8fd45b36e32ca4fb802bdeaecc",
		"kubevirt_hostpath_provisioner": "quay.io/kubevirt/hostpath-provisioner:v0.8.0",
		"pause":                         "k8s.gcr.io/pause:3.6",
		"service_ca_operator":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:8b0e194c1cb6babeb2d8326091deb2bba63c430938abe056eb89398c78733eb6",
	}
}
