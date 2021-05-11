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
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kubeproxy "k8s.io/kubernetes/cmd/kube-proxy/app"
	kubelet "k8s.io/kubernetes/cmd/kubelet/app"
)

/*
kubelet --config=/etc/kubernetes/kubelet.conf
--bootstrap-kubeconfig=/etc/kubernetes/kubeconfig
--kubeconfig=/var/lib/kubelet/kubeconfig
--container-runtime=remote
--container-runtime-endpoint=/var/run/crio/crio.sock
--runtime-cgroups=/system.slice/crio.service
--node-labels=node-role.kubernetes.io/master,node.openshift.io/os_id=rhcos
--node-ip=192.168.126.11
--minimum-container-ttl-duration=6m0s
--cloud-provider=
--volume-plugin-dir=/etc/kubernetes/kubelet-plugins/volume/exec
--register-with-taints=node-role.kubernetes.io/master=:NoSchedule
--pod-infra-container-image=quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:6eedefd9c899f7bd95978594d3a7f18fc3d9b54a53b70f58b29a3fb97bb65511
--v=2
*/
func StartKubelet(args []string) {
	command := kubelet.NewKubeletCommand()
	go func() {
		logrus.Fatalf("kubelet exited: %v", command.Execute())
	}()
}

/*
/usr/bin/openshift-sdn-node
--node-name crc-xl2km-master-0
--node-ip 192.168.126.11
--proxy-config /config/kube-proxy-config.yaml
--v 2
*/
func StartKubeProxy(args []string) {
	command := kubeproxy.NewProxyCommand()
	go func() {
		logrus.Fatalf("kube-proxy exited: %v", command.Execute())
	}()
}
