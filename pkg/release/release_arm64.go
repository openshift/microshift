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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:b611040d5cc03ab7de6f33171272ab585c2d9faec04b8608eb86b95127fdf41e",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e4badf698ae2943089bd505cc9f3425aed78b77b87ccbab8267a72c26df5b4d4",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:86700873067ad570d81bb7bd06713424aade7154c0409413e7bb0e957fd7722e",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:d98acd58b8df042657631ad6c1edf8e2ed51621dd43e7eab3e791e1d293e32e6",
		"pause":                     "k8s.gcr.io/pause:3.6",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:4f6229a4b1a4732a10ee06c222500e57b548bc9cc10e2bc4c01af3a5d6128a79",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:3f781a07e59d164eba065dba7d8e7661ab2494b21199c379b65b0ff514a1b8d0",
		"ovn_kubernetes_microshift": "quay.io/microshift/ovn-kubernetes-singlenode@sha256:012e743363b5f15f442c238099d35a0c70343fd1d4dc15b0a57a7340a338ffdb",
	}
}
