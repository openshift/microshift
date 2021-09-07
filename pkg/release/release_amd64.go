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
	"cli":                           "quay.io/openshift/okd-content@sha256:b20d195c721cd3b6215e5716b5569cbabbe861559af7dce07b5f8f3d38e6d701", // for dns-node-resolver
	"coredns":                       "quay.io/openshift/okd-content@sha256:a51b153b85c92c76017eee22582630b3ba6585ee839cf9b23a18f8598c326254",
	"haproxy_router":                "quay.io/openshift/okd-content@sha256:d09ab1bfce0ec273183cbe822aa5c6bdceeaee1685753de27b0ab9c15b54d8c0",
	"kube_flannel":                  "quay.io/coreos/flannel:v0.14.0",
	"kube_rbac_proxy":               "quay.io/openshift/okd-content@sha256:fcce680899a37d6bdc621a58b6da0587d01cbb49a2d7b713e0d606dffc9f685a",
	"kubevirt_hostpath_provisioner": "quay.io/kubevirt/hostpath-provisioner:v0.8.0",
	"pause":                         "k8s.gcr.io/pause",
	"service_ca_operator":           "quay.io/openshift/okd-content@sha256:9f32b0d4b6a08f8f40ff32772caa6ad152e8a045cfbf34b857b02d4cfa4e3fc7",
}
