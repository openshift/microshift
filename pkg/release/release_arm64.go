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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:fe65a036a65af078f6f61017ae96e141dbb203f3602ecaca7f63ec8f58a1f6c6",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:b5b3d024b2586bd0bf7b1315b2866f36a9b8b0acd23f0a9c6459371234dc8429",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:349e73813f432203920ae9ed04fc33a4026507e26ecc23ff2ab609d5b95b4206",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:c19226019fe605b5ab10496fb0b7cb4712cb694a7ee1e26642d63d515ca6b7cc",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:3f781a07e59d164eba065dba7d8e7661ab2494b21199c379b65b0ff514a1b8d0",
		"ovn_kubernetes_microshift": "quay.io/microshift/ovn-kubernetes-singlenode@sha256:012e743363b5f15f442c238099d35a0c70343fd1d4dc15b0a57a7340a338ffdb",
		"pause":                     "k8s.gcr.io/pause:3.6",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2fe468f25881e7b5ae8118c7d54b41a7fbb132a186f0156bbe46df0fd6a2f1f8",
		"odf_topolvm":               "quay.io/rhceph-dev/odf4-odf-topolvm-rhel8@sha256:2855918d1849c99a835eb03c53ce07170c238111fd15d2fe50cd45611fcd1ceb",
		"ose_csi_ext_provisioner":   "quay.io/rhceph-dev/openshift-ose-csi-external-provisioner@sha256:c3b2417f8fcb8883275f0e613037f83133ccc3f91311a30688e4be520544ea4a",
		"ose_csi_ext_resizer":       "quay.io/rhceph-dev/openshift-ose-csi-external-resizer@sha256:213f43d61b3a214a4a433c7132537be082a108d55005f2ba0777c2ea97489799",
		"topolvm-csi-snapshotter":   "quay.io/rhceph-dev/openshift-ose-csi-external-snapshotter@sha256:734c095670d21b77f18c84670d6c9a7742be1d9151dca0da20f41858ede65ed8",
		"ose_csi_livenessprobe":     "quay.io/rhceph-dev/openshift-ose-csi-livenessprobe@sha256:b05559aa038708ab448cfdfed2ca880726aed6cc30371fea4d6a42c972c0c728",
		"ose_csi_node_registrar":    "quay.io/rhceph-dev/openshift-ose-csi-node-driver-registrar@sha256:fb0f5e531847db94dcadc61446b9a892f6f92ddf282e192abf2fdef6c6af78f2",
	}
}
