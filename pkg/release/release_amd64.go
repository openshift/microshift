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
		"haproxy_router":            "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e8ab526e34493eb6cac4be5c3be867ab992c1e82d3d415b7c0fcc67392282124",
		"kube_rbac_proxy":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:3b75e8e24ef2235370454dd557ae59f93bf47c3e17653109ab7d869633728347",
		"odf_topolvm":               "quay.io/rh-storage-partners/microshift-topolvm@sha256:616fe64c9f2d1315cec655d482e7b26596594e879e07017e0e610d37c72bacd0",
		"openssl":                   "registry.access.redhat.com/ubi8/openssl@sha256:9e743d947be073808f7f1750a791a3dbd81e694e37161e8c6c6057c2c342d671",
		"csi_external_provisioner":  "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:60a5bf9f22b3f86c831c885863ce1d6928afb18e997967b46e6a8f97323c46e1",
		"csi_external_resizer":      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:ace72d90893a700dbecb1c94bf666b8832e1b6c15d3d6f6f8bcfbdc575cdaa7e",
		"csi_node_driver_registrar": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:3eb045e477f60d68daa738654b8d9c14aee77c6b15d92c6744d1db36422ac54f",
		"csi_livenessprobe":         "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:eea41d47c843567370789a15361d7624dd6e50770b51e9fd60f887d3172b27dd",
		"ovn_kubernetes_microshift": "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:d7bfd8e664d20e6ba259ba2677b542ff65a554ef326cc49bdbbfdecf89916635",
		"pod":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2b761fd6cc77514f421e2216ccb8f625c1e216f99cf65b922ccce1cf25f9773a",
		"service_ca_operator":       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2741fb3a1088349c089868bb57bbd1d0416e4425f4e8f95b62df9bdc267a19d7",
	}
}
