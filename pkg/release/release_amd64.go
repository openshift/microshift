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
// published OKD or other component upstream images

func init() {
	Image = map[string]string{
		"cli":                           "quay.io/openshift/okd-content@sha256:5bca6c5df37e9191bda7e265d47473b4cd59c941a3e70966f0f84470d1f65a24",
		"coredns":                       "quay.io/openshift/okd-content@sha256:82723e1d41ab68c0fde2e2f8bfa22ea470a11dd3101d7d214cdf9dd63171788d",
		"haproxy_router":                "quay.io/openshift/okd-content@sha256:891559e2a2fc2cace2043d39f6a14d0e5e923ff127325b38267aa79eb663c5bd",
		"kube_flannel":                  "quay.io/coreos/flannel:v0.14.0",
		"kube_flannel_cni":              "quay.io/microshift/flannel-cni:" + Base,
		"kube_rbac_proxy":               "quay.io/openshift/okd-content@sha256:baedb268ac66456018fb30af395bb3d69af5fff3252ff5d549f0231b1ebb6901",
		"kubevirt_hostpath_provisioner": "quay.io/kubevirt/hostpath-provisioner:v0.8.0",
		"pause":                         "k8s.gcr.io/pause:3.6",
		"service_ca_operator":           "quay.io/openshift/okd-content@sha256:0692de34bfd9f40455a5afc69fb47be56b1f386c55183b569d9d71d2969ff2e6",
	}
}
