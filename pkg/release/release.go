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

var Base = "4.10.18"

var Image = map[string]string{
	"cli":                           "quay.io/microshift/cli:" + Base,
	"coredns":                       "quay.io/microshift/coredns:" + Base,
	"haproxy_router":                "quay.io/microshift/haproxy-router:" + Base,
	"kube_rbac_proxy":               "quay.io/microshift/kube-rbac-proxy:" + Base,
	"pause":                         "quay.io/microshift/pause:" + Base,
	"service_ca_operator":           "quay.io/microshift/service-ca-operator:" + Base,
	"ovn_kubernetes":                "quay.io/microshift/ovn-kubernetes-singlenode:" + Base,
	"odf_lvm_topolvm":               "registry.redhat.io/odf4/odf-topolvm-rhel8@sha256:bd9fb330fc35f88fae65f1598b802923c8a9716eeec8432bdf05d16bd4eced64",
	"odf_lvm_operator":              "registry.redhat.io/odf4/odf-lvm-rhel8-operator@sha256:4f486e6f92a4810ceebeb053bb2848728da36ba1285123407e308ef9ef6dbfbb",
	"odf_lvm_ext_provisioner":       "registry.redhat.io/openshift4/ose-csi-external-provisioner@sha256:42563eb25efb2b6f277944b627bea420fa58fe950b46a1bd1487122b8a387e75",
	"odf_lvm_ext_resizer":           "registry.redhat.io/openshift4/ose-csi-external-resizer@sha256:75017593988025df444c8b3849b6ba867c3a7f6fc83212aeff2dfc3de4fabd21",
	"ose_csi_node_driver_registrar": "registry.redhat.io/openshift4/ose-csi-node-driver-registrar@sha256:376f21cfa8308dc1b61a3e8401b7023d903eda768912699f39403de742ab88b1",
	"ose_csi_livenessprobe":         "registry.redhat.io/openshift4/ose-csi-livenessprobe@sha256:058fd6f949218cd3a76d8974ff1ea27fd45cba4662d14e3561285c779f0f0de5",
}
