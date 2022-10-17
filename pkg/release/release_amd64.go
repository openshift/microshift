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
		"cli":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:c7ae4cdee454e8aeb8b175696ae47c2f66137f517fca2618c6f34f5b6fcbd687",
		"coredns":                   "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:12b1435f5f1b1f1f0930f6b5571818c3c6840e4dff769833e866844fc97458b3",
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:1751e269e76b2a92b9065aa18ac2ccf3237e87be405c4f1fd5e85bba1e52c130",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:16a2195d4c6485f1051e0e59e7fab40e094f7fe0146f9b2562a744feab8f625c",
		"odf_topolvm":               "registry.redhat.io/odf4/odf-topolvm-rhel8@sha256:362c41177d086fc7c8d4fa4ac3bbedb18b1902e950feead9219ea59d1ad0e7ad",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:8b41865d30b7947de68a9c1747616bce4efab4f60f68f8b7016cd84d7708af6b",
		"ose_csi_ext_provisioner":   "registry.redhat.io/openshift4/ose-csi-external-provisioner@sha256:4b7d8035055a867b14265495bd2787db608b9ff39ed4e6f65ff24488a2e488d2",
		"ose_csi_ext_resizer":       "registry.redhat.io/openshift4/ose-csi-external-resizer@sha256:ca34c46c4a4c1a4462b8aa89d1dbb5427114da098517954895ff797146392898",
		"ose_csi_node_registrar":    "registry.redhat.io/openshift4/ose-csi-node-driver-registrar@sha256:3babcf219371017d92f8bc3301de6c63681fcfaa8c344ec7891c8e84f31420eb",
		"ose_csi_livenessprobe":     "registry.redhat.io/openshift4/ose-csi-livenessprobe@sha256:e4b0f6c89a12d26babdc2feae7d13d3f281ac4d38c24614c13c230b4a29ec56e",
		"ovn_kubernetes_microshift": "quay.io/microshift/ovn-kubernetes-singlenode@sha256:e97d6035754fad1660b522b8afa4dea2502d5189c8490832e762ae2afb4cf142",
		"pause":                     "k8s.gcr.io/pause:3.6",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2a943123cf987185526ad781450264474ae5008b6b02cdce4c715754605da3c4",
	}
}
