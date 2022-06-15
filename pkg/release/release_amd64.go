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
		"cli":                           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:c1fd4f50fd54ab05d76675b8ac6658550a5fbec2292656150a41b5a9c04fff3a",
		"coredns":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:df20060da75fe9571ad05f964e98599304dc90a5b008c75eb5aeaddf04b022a6",
		"haproxy_router":                "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:d4be2076f3bb393db613df3de24cfc951f86693008c9f2f232677ebeb4ab8f66",
		"kube_flannel":                  "quay.io/coreos/flannel:v0.14.0",
		"kube_flannel_cni":              "quay.io/microshift/flannel-cni:v0.14.0",
		"kube_rbac_proxy":               "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:aa8d1daf3432d8dedc5c56d94aeb1f25301bce6ccd7d5406fb03a00be97374ad",
		"kubevirt_hostpath_provisioner": "quay.io/kubevirt/hostpath-provisioner:v0.8.0",
		"pause":                         "k8s.gcr.io/pause:3.7",
		"service_ca_operator":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:6679f1fb9aa5e15a4b6224ac92317461ea2730788862445f84e7c78644d3a2ca",
	}
}
