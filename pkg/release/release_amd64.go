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
		"odf_topolvm":               "quay.io/rhceph-dev/odf4-odf-topolvm-rhel8@sha256:2855918d1849c99a835eb03c53ce07170c238111fd15d2fe50cd45611fcd1ceb",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:8b41865d30b7947de68a9c1747616bce4efab4f60f68f8b7016cd84d7708af6b",
		"ose_csi_ext_provisioner":   "quay.io/rhceph-dev/openshift-ose-csi-external-provisioner@sha256:c3b2417f8fcb8883275f0e613037f83133ccc3f91311a30688e4be520544ea4a",
		"ose_csi_ext_resizer":       "quay.io/rhceph-dev/openshift-ose-csi-external-resizer@sha256:213f43d61b3a214a4a433c7132537be082a108d55005f2ba0777c2ea97489799",
		"ose_csi_node_registrar":    "quay.io/rhceph-dev/openshift-ose-csi-node-driver-registrar@sha256:fb0f5e531847db94dcadc61446b9a892f6f92ddf282e192abf2fdef6c6af78f2",
		"ose_csi_livenessprobe":     "quay.io/rhceph-dev/openshift-ose-csi-livenessprobe@sha256:b05559aa038708ab448cfdfed2ca880726aed6cc30371fea4d6a42c972c0c728",
		"topolvm-csi-snapshotter":   "quay.io/rhceph-dev/openshift-ose-csi-external-snapshotter@sha256:734c095670d21b77f18c84670d6c9a7742be1d9151dca0da20f41858ede65ed8",
		"ovn_kubernetes_microshift": "quay.io/microshift/ovn-kubernetes-singlenode@sha256:e97d6035754fad1660b522b8afa4dea2502d5189c8490832e762ae2afb4cf142",
		"pause":                     "k8s.gcr.io/pause:3.6",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:afcc1f59015b394e6da7d73eba32de407807da45018e3c4ecc25e5741aaae2dd",
	}
}
