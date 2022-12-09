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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:9945c3f5475a37e145160d2fe6bb21948f1024a856827bc9e7d5bc882f44a750",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:82cfef91557f9a70cff5a90accba45841a37524e9b93f98a97b20f6b2b69e5db",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:a66f5038b4499e3c069067365c8426388d09bf3ac4dd8eb8bcbd39cd5f6c6ed0",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:dac19e13041b0bb5d60ecaab219a43c2b4ea57a082cf13bc6305acf86254432e",
		"odf_topolvm":               "quay.io/rh-storage-partners/microshift-topolvm@sha256:616fe64c9f2d1315cec655d482e7b26596594e879e07017e0e610d37c72bacd0",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:9e743d947be073808f7f1750a791a3dbd81e694e37161e8c6c6057c2c342d671",
		"csi_external_provisioner":  "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:82983bdae16b3cadc78539f687ea39f6bb7af1ed99f3382fdbcb61500ed30398",
		"csi_external_resizer":      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b951d8b5f88142bc1a7f5ca8529e3a6a89c8ff2dd78c9c06e8f6194e3d681f",
		"csi_node_driver_registrar": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f8f85305bd4a9fc8796a05cb27676e084e80562c8af421fa99b44ef0441beff9",
		"csi_livenessprobe":         "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:46bfcb0440620f12a91be6493039b07bf7c96d104a1a59b3e8ed3caaec2dda5c",
		"ovn_kubernetes_microshift": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:34faa03671a950607aec06a2ba515426d09f80a38101d2e91a508b5fd5316116",
		"pod":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f8979cde8702f9e3318f5b0a6165bc659536248d5a1ce87aa24bf265062215c0",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:20fd4b23b852db6f9f0d0134ad8d9128cb771af16723a8e239bcd97e5cd874b4",
	}
}
