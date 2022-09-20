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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:26153cfac611418881dad62d9bb2ea46b5c93ba586a68e752335b657dd79032a",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:cdb8a4c24c404c0c03b0ed2a8221f5a71cf735c1969ff38a3cc3631a754fe7ad",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:b008d8526ffdf997d62309ec12e7b6ba673f76235a5ffe7236db57f407dc24ba",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:174b4f8995157d0f1b5533c9d179c1eb681415a0bd092b4c1b14c2ed1f28083c",
		"pause":                     "k8s.gcr.io/pause:3.6",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:0fde163f17d82d39b75b5ec6a5f2b36652e02b29d6d339c52017de135e4be47a",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:8b41865d30b7947de68a9c1747616bce4efab4f60f68f8b7016cd84d7708af6b",
		"ovn_kubernetes_microshift": "quay.io/microshift/ovn-kubernetes-singlenode@sha256:e97d6035754fad1660b522b8afa4dea2502d5189c8490832e762ae2afb4cf142",
		"odf_topolvm":               "registry.redhat.io/odf4/odf-topolvm-rhel8@sha256:bd9fb330fc35f88fae65f1598b802923c8a9716eeec8432bdf05d16bd4eced64",
		"ose_csi_ext_provisioner":   "registry.redhat.io/openshift4/ose-csi-external-provisioner@sha256:42563eb25efb2b6f277944b627bea420fa58fe950b46a1bd1487122b8a387e75",
		"ose_csi_ext_resizer":       "registry.redhat.io/openshift4/ose-csi-external-resizer@sha256:75017593988025df444c8b3849b6ba867c3a7f6fc83212aeff2dfc3de4fabd21",
		"ose_csi_node_registrar":    "registry.redhat.io/openshift4/ose-csi-node-driver-registrar@sha256:376f21cfa8308dc1b61a3e8401b7023d903eda768912699f39403de742ab88b1",
		"ose_csi_livenessprobe":     "registry.redhat.io/openshift4/ose-csi-livenessprobe@sha256:058fd6f949218cd3a76d8974ff1ea27fd45cba4662d14e3561285c779f0f0de5",
	}
}
