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
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/openshift/microshift/pkg/config"

	kubeproxy "k8s.io/kubernetes/cmd/kube-proxy/app"
	kubelet "k8s.io/kubernetes/cmd/kubelet/app"
)

func StartKubelet(cfg *config.MicroshiftConfig) error {
	command := kubelet.NewKubeletCommand()
	args := []string{
		"--config=" + cfg.DataDir + "/resources/kubelet/config/config.yaml",
		"--bootstrap-kubeconfig=" + cfg.DataDir + "/resources/kubelet/kubeconfig",
		"--kubeconfig=" + cfg.DataDir + "/resources/kubelet/kubeconfig",
		"--container-runtime=remote",
		"--container-runtime-endpoint=unix:///var/run/crio/crio.sock",
		"--runtime-cgroups=/system.slice/crio.service",
		"--node-ip=" + cfg.NodeIP,
		"--volume-plugin-dir=" + cfg.DataDir + "/kubelet-plugins/volume/exec",
		"--logtostderr=" + strconv.FormatBool(cfg.LogDir == "" || cfg.LogAlsotostderr),
		"--alsologtostderr=" + strconv.FormatBool(cfg.LogAlsotostderr),
		"--v=" + strconv.Itoa(cfg.LogVLevel),
		"--vmodule=" + cfg.LogVModule,
	}
	if cfg.LogDir != "" {
		args = append(args, "--log-file="+filepath.Join(cfg.LogDir, "kubelet.log"))
	}

	if err := command.ParseFlags(args); err != nil {
		logrus.Fatalf("failed to parse flags:%v", err)
	}
	logrus.Infof("starting kubelet %s, args: %v", cfg.NodeIP, args)

	go func() {
		command.Run(command, args)
		logrus.Fatalf("kubelet exited")
	}()

	return nil
}

func StartKubeProxy(cfg *config.MicroshiftConfig) error {
	command := kubeproxy.NewProxyCommand()
	args := []string{
		"--config=" + cfg.DataDir + "/resources/kube-proxy/config/config.yaml",
		"--logtostderr=" + strconv.FormatBool(cfg.LogDir == "" || cfg.LogAlsotostderr),
		"--alsologtostderr=" + strconv.FormatBool(cfg.LogAlsotostderr),
		"--v=" + strconv.Itoa(cfg.LogVLevel),
		"--vmodule=" + cfg.LogVModule,
	}
	if cfg.LogDir != "" {
		args = append(args, "--log-file="+filepath.Join(cfg.LogDir, "kube-proxy.log"))
	}
	if err := command.ParseFlags(args); err != nil {
		logrus.Fatalf("failed to parse flags:%v", err)
	}
	logrus.Infof("starting kube-proxy, args: %v", args)

	go func() {
		command.Run(command, args)
		logrus.Fatalf("kube-proxy exited")
	}()

	return nil
}
