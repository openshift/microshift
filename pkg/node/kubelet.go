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
		"--node-ip=" + ip,
		"--volume-plugin-dir=/etc/kubernetes/kubelet-plugins/volume/exec",
		"--log-dir=/var/log",
		"--v=3",
	}
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

func StartKubeProxy() {
	command := kubeproxy.NewProxyCommand()
	args := []string{
		"--config=/etc/kubernetes/ushift-resources/kube-proxy/config/config.yaml",
		"--master=https://127.0.0.1:6443",
		"--log-dir=/var/log",
		"-v=3",
	}
	if err := command.ParseFlags(args); err != nil {
		logrus.Fatalf("failed to parse flags:%v", err)
	}
	logrus.Infof("starting kube-proxy, args: %v", args)

	go func() {
		command.Run(command, args)
		logrus.Fatalf("kube-proxy exited")
	}()

}
