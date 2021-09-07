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

var Base = "4.7.0-0.okd-2021-06-13-090745"

var Image = map[string]string{
	"cli":                           "quay.io/microshift/coredns:1.6.9", // for dns-node-resolver
	"coredns":                       "quay.io/microshift/coredns:1.6.9",
	"haproxy_router":                "quay.io/microshift/openshift-router:4.5",
	"kube_flannel":                  "quay.io/microshift/flannel:v0.14.0",
	"kube_rbac_proxy":               "quay.io/microshift/kube-rbac-proxy:v0.11.0",
	"kubevirt_hostpath_provisioner": "quay.io/microshift/hostpath-provisioner:v0.9.0",
	"pause":                         "quay.io/microshift/pause:3.2",
	"service_ca_operator":           "quay.io/microshift/service-ca-operator:latest",
}
