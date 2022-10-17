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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:4d182d11a30e6c3c1420502bec5b1192c43c32977060c4def96ea160172f71e7",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e5f97df4705b6f3a222491197000b887d541e9f3a440a7456f94c82523193760",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:72c751aa148bf498839e6f37b304e3265f85af1e00578e637332a13ed9545ece",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:dd49360368f93bbe1a11b8d1ce6f0f98eeb0c9230d9801a2b08a714a92e1f655",
		"odf_topolvm":               "registry.redhat.io/odf4/odf-topolvm-rhel8@sha256:362c41177d086fc7c8d4fa4ac3bbedb18b1902e950feead9219ea59d1ad0e7ad",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:8b41865d30b7947de68a9c1747616bce4efab4f60f68f8b7016cd84d7708af6b",
		"ose_csi_ext_provisioner":   "registry.redhat.io/openshift4/ose-csi-external-provisioner@sha256:4b7d8035055a867b14265495bd2787db608b9ff39ed4e6f65ff24488a2e488d2",
		"ose_csi_ext_resizer":       "registry.redhat.io/openshift4/ose-csi-external-resizer@sha256:ca34c46c4a4c1a4462b8aa89d1dbb5427114da098517954895ff797146392898",
		"ose_csi_node_registrar":    "registry.redhat.io/openshift4/ose-csi-node-driver-registrar@sha256:3babcf219371017d92f8bc3301de6c63681fcfaa8c344ec7891c8e84f31420eb",
		"ose_csi_livenessprobe":     "registry.redhat.io/openshift4/ose-csi-livenessprobe@sha256:e4b0f6c89a12d26babdc2feae7d13d3f281ac4d38c24614c13c230b4a29ec56e",
		"ovn_kubernetes_microshift": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:b91accd2ceab9e8aa2fecb71b04c66914a92b192f90aeb96e52ef93e797e6306",
		"pod":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:c66675912585b7da1f5e33d4249cb2424e6d155a9038bbd393f3e4456329b827",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:afcc1f59015b394e6da7d73eba32de407807da45018e3c4ecc25e5741aaae2dd",
	}
}
