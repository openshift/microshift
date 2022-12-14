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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:caa0fe9b53e4b0d2c9988fc20fa745ad80817a7cec36ba098d1812fab03e2add",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:cbe32c3940f369eb9093d2b6669a22ce4fd3b1c0781c2afd74f1b0b1e6bd3a9d",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e19d3bcfc797cd879db6043ef59c1d9e9c8c199181f5f2b6cca5e4c7cfed0a5d",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:329566d40a19ff6914c4d584c7526c2093917a1437eb32c9f299f1c62350d035",
		"odf_topolvm":               "quay.io/rh-storage-partners/microshift-topolvm@sha256:616fe64c9f2d1315cec655d482e7b26596594e879e07017e0e610d37c72bacd0",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:9e743d947be073808f7f1750a791a3dbd81e694e37161e8c6c6057c2c342d671",
		"csi_external_provisioner":  "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f8a246885c509a113cbd7ce43f78ea764752fad2f1bf2b61849abcaa77baacff",
		"csi_external_resizer":      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:b9524eb63c3408c2889ec926f2ebdf9d4ab4689c3ad50594eb8d80a9bdd0dbc9",
		"csi_node_driver_registrar": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:66d99027850fac4ed6f6f9cef8f6247c08881f75648492b76a9e7f50ff9dc115",
		"csi_livenessprobe":         "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:ea75863f09d2a45ef549e08b6c75fe2058ca142ecc53d793408d5a80982c90d7",
		"ovn_kubernetes_microshift": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e025283685bfc3dca0c7c1ae0453baaa1c9e06a196863edac8b3f4d9e8c4c972",
		"pod":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:fc1bb41085a8bb979c1d5242c8925174a1265483a0295943fe3f7c90a3939b2d",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2b601e2889b25a6175c9d65c834c310262696094a9bcb5f49d6c4e2682392727",
	}
}
