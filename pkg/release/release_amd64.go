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
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2e14189c0096c2294368a1a8edd7dec5f30c93f8bbd614da0e78127c8b194ab7",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:3b75e8e24ef2235370454dd557ae59f93bf47c3e17653109ab7d869633728347",
		"odf_topolvm":               "registry.redhat.io/odf4/odf-topolvm-rhel8@sha256:362c41177d086fc7c8d4fa4ac3bbedb18b1902e950feead9219ea59d1ad0e7ad",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:8b41865d30b7947de68a9c1747616bce4efab4f60f68f8b7016cd84d7708af6b",
		"ose_csi_ext_provisioner":   "registry.redhat.io/openshift4/ose-csi-external-provisioner@sha256:4b7d8035055a867b14265495bd2787db608b9ff39ed4e6f65ff24488a2e488d2",
		"ose_csi_ext_resizer":       "registry.redhat.io/openshift4/ose-csi-external-resizer@sha256:ca34c46c4a4c1a4462b8aa89d1dbb5427114da098517954895ff797146392898",
		"ose_csi_node_registrar":    "registry.redhat.io/openshift4/ose-csi-node-driver-registrar@sha256:3babcf219371017d92f8bc3301de6c63681fcfaa8c344ec7891c8e84f31420eb",
		"ose_csi_livenessprobe":     "registry.redhat.io/openshift4/ose-csi-livenessprobe@sha256:e4b0f6c89a12d26babdc2feae7d13d3f281ac4d38c24614c13c230b4a29ec56e",
		"ovn_kubernetes_microshift": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:cd9365d7ab0a70fd0d67937853bed13eaece6f49895aed34d7eca038f5e0aa77",
		"pod":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:f2e139e96869aa0f807c42cc557b8f857de9cf749baa00c881f57b10679f77e2",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2741fb3a1088349c089868bb57bbd1d0416e4425f4e8f95b62df9bdc267a19d7",
	}
}
