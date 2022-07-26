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
		"cli":                           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:65889fde898f3025462f1344bd93e4a1d8ae5a764224b61b4f0b7e46cc1108e4",
		"coredns":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2619cbcbc5ea990bd8349848d410cc4040aea0265821b9d1e46b364f0519481b",
		"haproxy_router":                "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:972f3fadc1b7eab2eaa02eff4128149eabe9577943dcb3e56f44dad7fc14059e",
		"kube_rbac_proxy":               "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2d8319a33aa781918a3173e2061635d8534cc5e4a39171775830ff7e2ce646ac",
		"kubevirt_hostpath_provisioner": "quay.io/microshift/hostpath-provisioner:4.10.0-0.okd-2022-04-23-131357",
		"pause":                         "k8s.gcr.io/pause:3.6",
		"service_ca_operator":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:97a7d56ac6d6a38f7977b0759bb1b97fe3895e6020364e60a8a2926f3fd0084e",
		"ovn_kubernetes":                "quay.io/microshift/ovn-kubernetes-singlenode:4.10.18",
	}
}
