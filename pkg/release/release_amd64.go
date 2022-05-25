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
		"cli":                           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:2fb65a85fe67c15dc28ffcb9464cf6b39bc24b97b9795379074e565b31a6659a",
		"coredns":                       "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:26956e07a594b8665740d9cff7d9c30361ce8dbb1523a996c3aadf95ae77363b",
		"haproxy_router":                "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:dfb2fbf7f0716402649a0e41563664901830ef990188ff86642dc8748291ad6a",
		"kube_flannel":                  "quay.io/coreos/flannel:v0.14.0",
		"kube_flannel_cni":              "quay.io/microshift/flannel-cni:v0.14.0",
		"kube_rbac_proxy":               "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:090ae5e7554012e1c0f1925f8dd7a02e110cb98f94d8774d3e17039115b8a109",
		"kubevirt_hostpath_provisioner": "quay.io/kubevirt/hostpath-provisioner:v0.8.0",
		"pause":                         "k8s.gcr.io/pause:3.6",
		"service_ca_operator":           "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:a62049a7a0c3184319db9f8e423de3b45d17c1733d54f059e03ce4fd0f9c346d",
	}
}
