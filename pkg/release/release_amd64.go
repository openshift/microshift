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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:24bfe9543bd34c6c3124c39c319ef0ec20534aec974126617752b2883d6d6cff",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:efa3a9aae6ad83d0eec44b654e75a11ed4887a4d25f4a7412a102456688a5840",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:4d39e22dc0ba8011189552689408228d9dcd5cce2c382641121afc39f641aa7a",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:3b75e8e24ef2235370454dd557ae59f93bf47c3e17653109ab7d869633728347",
		"odf_topolvm":               "quay.io/rh-storage-partners/microshift-topolvm@sha256:616fe64c9f2d1315cec655d482e7b26596594e879e07017e0e610d37c72bacd0",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:8b41865d30b7947de68a9c1747616bce4efab4f60f68f8b7016cd84d7708af6b",
		"ose_csi_ext_provisioner":   "registry.redhat.io/openshift4/ose-csi-external-provisioner@sha256:4b7d8035055a867b14265495bd2787db608b9ff39ed4e6f65ff24488a2e488d2",
		"ose_csi_ext_resizer":       "registry.redhat.io/openshift4/ose-csi-external-resizer@sha256:ca34c46c4a4c1a4462b8aa89d1dbb5427114da098517954895ff797146392898",
		"ose_csi_node_registrar":    "registry.redhat.io/openshift4/ose-csi-node-driver-registrar@sha256:3babcf219371017d92f8bc3301de6c63681fcfaa8c344ec7891c8e84f31420eb",
		"ose_csi_livenessprobe":     "registry.redhat.io/openshift4/ose-csi-livenessprobe@sha256:e4b0f6c89a12d26babdc2feae7d13d3f281ac4d38c24614c13c230b4a29ec56e",
		"ovn_kubernetes_microshift": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:3d5077683f71fc6d16095b85426115dcc6ed414947dc572d21c6c1cbdb1761be",
		"pod":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:be5ae7ff1af2efa6789c73ff50e0164f6b063e6398a9084105cf1e8f00c8589f",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2741fb3a1088349c089868bb57bbd1d0416e4425f4e8f95b62df9bdc267a19d7",
	}
}
