/*
Copyright © 2021 Microshift Contributors

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
		"cli":                     "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:ba44dead03ea74107f90d58525106fb52d27a120b73c6cc8e2be31d37043ca1c",
		"coredns":                 "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2b423e88cdd37f307aff51cbb0f53fc45deff9618f5b4f12bfb78bea7aff51a2",
		"haproxy_router":          "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:5f9ee3afa744e790dbb61d08f44e30370c9a5ff041054bf99dc1afe58792cd7b",
		"kube_rbac_proxy":         "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:3df935837634adb5df8080ac7263c7fb4c4f9d8fd45b36e32ca4fb802bdeaecc",
		"pause":                   "k8s.gcr.io/pause:3.6",
		"service_ca_operator":     "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:8b0e194c1cb6babeb2d8326091deb2bba63c430938abe056eb89398c78733eb6",
		"ovn_kubernetes":          "quay.io/microshift/ovn-kubernetes-singlenode:4.10.18",
		"odf_lvm_topolvm":         "registry.redhat.io/odf4/odf-topolvm-rhel8@sha256:bd9fb330fc35f88fae65f1598b802923c8a9716eeec8432bdf05d16bd4eced64",
		"odf_lvm_operator":        "registry.redhat.io/odf4/odf-lvm-rhel8-operator@sha256:4f486e6f92a4810ceebeb053bb2848728da36ba1285123407e308ef9ef6dbfbb",
		"ose_csi_ext_provisioner": "registry.redhat.io/openshift4/ose-csi-external-provisioner@sha256:42563eb25efb2b6f277944b627bea420fa58fe950b46a1bd1487122b8a387e75",
		"ose_csi_ext_resizer":     "registry.redhat.io/openshift4/ose-csi-external-resizer@sha256:75017593988025df444c8b3849b6ba867c3a7f6fc83212aeff2dfc3de4fabd21",
		"ose_csi_node_registrar":  "registry.redhat.io/openshift4/ose-csi-node-driver-registrar@sha256:376f21cfa8308dc1b61a3e8401b7023d903eda768912699f39403de742ab88b1",
		"ose_csi_livenessprobe":   "registry.redhat.io/openshift4/ose-csi-livenessprobe@sha256:058fd6f949218cd3a76d8974ff1ea27fd45cba4662d14e3561285c779f0f0de5",
	}
}
