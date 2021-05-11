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
package constant

const (
	AdminKubeconfigPath                 = "/etc/kubernetes/ushift-resources/kubeadmin/kubeconfig"
	KubeAPIKubeconfigPath               = "/etc/kubernetes/ushift-resources/kube-apiserver/kubeconfig"
	KubeControllerManagerKubeconfigPath = "/etc/kubernetes/ushift-resources/kube-controller-manager/kubeconfig"
	KubeSchedulerKubeconfigPath         = "/etc/kubernetes/ushift-resources/kube-scheduler/kubeconfig"
	ClusterCIDR                         = "10.42.0.0/16"
	ServiceCIDR                         = "10.43.0.0/16"
	ClusterDNS                          = "10.43.0.10"
	DomainName                          = "ushift.testing"
)
