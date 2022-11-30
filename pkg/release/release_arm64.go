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
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:aed96aded2e2c981d8d5ea3eaeaf5fb431b79006c395eeeb2cb5545de2f94433",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f4d4b35f8e4a9bf073ea40e71bb53c063e29cc9ab30fcbcb54e03a29781a6cd6",
		"odf_topolvm":               "quay.io/rh-storage-partners/microshift-topolvm@sha256:616fe64c9f2d1315cec655d482e7b26596594e879e07017e0e610d37c72bacd0",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:9e743d947be073808f7f1750a791a3dbd81e694e37161e8c6c6057c2c342d671",
		"csi_external_provisioner":  "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:883633404e41da2d244e3fff5393ace54be59f819232ddf4c05a2e9f60e354b8",
		"csi_external_resizer":      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:05c7a5222d820145c1573f10e4b4ac75381e4dd8d6e8ed1e2cf521f1f9bd8b6a",
		"csi_node_driver_registrar": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:8a6db58bbd146b62dd5c10deaf7f9ca5cb4381e0aac89e229934fc3cc012929f",
		"csi_livenessprobe":         "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:7927bb88a66de756483c1f0ebc871d3d15dd1e88720fdfc6d05b7e6d7aef000b",
		"ovn_kubernetes_microshift": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:111d3d6f938d675e2320d6a59d126efcb4265f708197c16085db749b267e6ee9",
		"pod":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f107c8849e895c9df0ccec5bcffed27f8576544b63373372d654a2aa65908bc6",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:35cca3568753a576a93197d58fba68e5ab9a2c864fb330153b734d41733cffa6",
	}
}
