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

var Base = "4.7.0-0.okd-2021-08-22-163618"

var Image = map[string]string{
	"cli":                           "quay.io/openshift/okd-content@sha256:71c912b47f0b1ca79442d207bd3a5c1435c000412423fa89d2d12096d0420bd6",
	"coredns":                       "quay.io/openshift/okd-content@sha256:a51b153b85c92c76017eee22582630b3ba6585ee839cf9b23a18f8598c326254",
	"haproxy_router":                "quay.io/openshift/okd-content@sha256:5be59a445c7187bc7d9e0bf4971277b058b0786485597379969a5c7fdc136409",
	"kube_flannel":                  "quay.io/coreos/flannel:v0.14.0",
	"kube_rbac_proxy":               "quay.io/openshift/okd-content@sha256:fcce680899a37d6bdc621a58b6da0587d01cbb49a2d7b713e0d606dffc9f685a",
	"kubevirt_hostpath_provisioner": "quay.io/kubevirt/hostpath-provisioner:v0.8.0",
	"pause":                         "k8s.gcr.io/pause",
	"service_ca_operator":           "quay.io/openshift/okd-content@sha256:ea4e95cf5345f6eba15721713f94790febe1e09aa464de30944302e88e8a92a4",
}
