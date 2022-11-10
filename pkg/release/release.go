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

var Base = "4.12.0-0.nightly-2022-11-07-181244"

var Image = map[string]string{
	"cli":                       "quay.io/microshift/cli:" + Base,
	"coredns":                   "quay.io/microshift/coredns:" + Base,
	"haproxy_router":            "quay.io/microshift/haproxy-router:" + Base,
	"kube_rbac_proxy":           "quay.io/microshift/kube-rbac-proxy:" + Base,
	"odf_topolvm":               "quay.io/microshift/odf-topolvm-rhel8" + Base,
	"openssl":                   "quay.io/microshift/openssl" + Base,
	"ose_csi_ext_provisioner":   "quay.io/microshift/ose-csi-external-provisioner" + Base,
	"ose_csi_ext_resizer":       "quay.io/microshift/ose-csi-external-resizer" + Base,
	"ose_csi_node_registrar":    "quay.io/microshift/ose-csi-node-driver-registrar" + Base,
	"ose_csi_livenessprobe":     "quay.io/microshift/ose-csi-livenessprobe" + Base,
	"ovn_kubernetes_microshift": "quay.io/microshift/ovn-kubernetes-microshift:" + Base,
	"pod":                       "quay.io/microshift/pause:" + Base,
	"service_ca_operator":       "quay.io/microshift/service-ca-operator:" + Base,
}
