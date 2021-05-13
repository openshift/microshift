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
package node

import (
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/openshift/microshift/pkg/constant"
	"github.com/openshift/microshift/pkg/util"

	kubeproxy "k8s.io/kubernetes/cmd/kube-proxy/app"
	kubelet "k8s.io/kubernetes/cmd/kubelet/app"
)

/*
kubelet
--config=/etc/kubernetes/kubelet.conf
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
func StartKubelet() error {
	ip, err := util.GetHostIP()
	if err != nil {
		return fmt.Errorf("failed to get host IP: %v", err)
	}
	command := kubelet.NewKubeletCommand()
	args := []string{
		"--config=/etc/kubernetes/ushift-resources/kubelet/config/config.yaml",
		"--bootstrap-kubeconfig=" + constant.AdminKubeconfigPath,
		"--kubeconfig=" + constant.AdminKubeconfigPath,
		"--container-runtime=remote",
		"--container-runtime-endpoint=/var/run/crio/crio.sock",
		"--runtime-cgroups=/system.slice/crio.service",
		"--node-labels=node-role.kubernetes.io/master,node.openshift.io/os_id=rhcos",
		"--node-ip=" + ip,
		"--minimum-container-ttl-duration=6m0s",
		"--volume-plugin-dir=/etc/kubernetes/kubelet-plugins/volume/exec",
		"--register-with-taints=node-role.kubernetes.io/master=:NoSchedule",
		"--fail-swap-on=false",
		//"--pod-infra-container-image=quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:6eedefd9c899f7bd95978594d3a7f18fc3d9b54a53b70f58b29a3fb97bb65511
		"--v=2",
	}
	//command.DisableFlagParsing = false
	if err := command.ParseFlags(args); err != nil {
		logrus.Fatalf("failed to parse flags:%v", err)
	}
	logrus.Infof("starting kubelet %s, args: %v", ip, args)

	go func() {
		command.Run(command, args)
		logrus.Fatalf("kubelet exited")
	}()

	return nil
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
