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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:3d13d0552be6ea8f593ccce46ea4566a058429ca112f844703a315f51c9020b8",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f7e9edbcaffe940b0f5989a29a80245b4015342f6f4aae42b24caf909e1f2c6f",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:0ed84cfd46b75d0961c7ea98dcf78482758601997d1375d1a56fde8c3a43aefa",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:72f9e579dc1395bb5754550847fac926ab5a87051949c52a3ea8c709dc23fd7d",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:3f781a07e59d164eba065dba7d8e7661ab2494b21199c379b65b0ff514a1b8d0",
		"ovn_kubernetes_microshift": "quay.io/microshift/ovn-kubernetes-singlenode@sha256:012e743363b5f15f442c238099d35a0c70343fd1d4dc15b0a57a7340a338ffdb",
		"pause":                     "k8s.gcr.io/pause:3.6",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:989a43840f16af130f846df494e61e484d42ed679d311db79a6b698d2399b92e",
	}
}
