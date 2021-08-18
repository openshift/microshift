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
	"context"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openshift/microshift/pkg/config"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	cliflag "k8s.io/component-base/cli/flag"
	kubeproxy "k8s.io/kubernetes/cmd/kube-proxy/app"
	kubelet "k8s.io/kubernetes/cmd/kubelet/app"
	kubeletoptions "k8s.io/kubernetes/cmd/kubelet/app/options"
	kubeletconfig "k8s.io/kubernetes/pkg/kubelet/apis/config"
)

const (
	// Kubelet component name
	componentKubelet = "kubelet"
)

type Kubelet struct {
	kubeletflags   *kubeletoptions.KubeletFlags
	kubeconfig     *kubeletconfig.KubeletConfiguration
	kubeconfigfile string
}

func NewKubelet(cfg *config.MicroshiftConfig) (*Kubelet, error) {
	s := &Kubelet{}
	err := s.configure(cfg)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Kubelet) Name() string           { return "kubelet" }
func (s *Kubelet) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *Kubelet) configure(cfg *config.MicroshiftConfig) error {
	if err := config.KubeletConfig(cfg); err != nil {
		return err
	}

	kubeletConfig, err := kubeletoptions.NewKubeletConfiguration()

	if err != nil {
		logrus.Fatalf("Failed to create a new kubelet configuration: %v", err)
		return err
	}

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

	cmd := &cobra.Command{
		Use:          "kubelet",
		Long:         `kubelet`,
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return nil },
	}

	cleanFlagSet := pflag.NewFlagSet(componentKubelet, pflag.ContinueOnError)
	cleanFlagSet.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	kubeletFlags := kubeletoptions.NewKubeletFlags()
	kubeletFlags.AddFlags(cleanFlagSet)
	kubeletoptions.AddKubeletConfigFlags(cleanFlagSet, kubeletConfig)
	kubeletoptions.AddGlobalFlags(cleanFlagSet)

	if err := cmd.ParseFlags(args); err != nil {
		logrus.Fatalf("failed to parse flags:%v", err)
		return err
	}
	s.kubeletflags = kubeletFlags
	s.kubeconfig = kubeletConfig
	s.kubeconfigfile = filepath.Join(cfg.DataDir, "resources", "kubelet", "kubeconfig")

	logrus.Infof("Starting kubelet %s, args: %v", cfg.NodeIP, args)
	return nil
}

func (s *Kubelet) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	// construct a KubeletServer from kubeletFlags and kubeletConfig
	kubeletServer := &kubeletoptions.KubeletServer{
		KubeletFlags:         *s.kubeletflags,
		KubeletConfiguration: *s.kubeconfig,
	}

	kubeletDeps, err := kubelet.UnsecuredDependencies(kubeletServer, utilfeature.DefaultFeatureGate)
	if err != nil {
		logrus.Fatalf("Error in fetching depenedencies %v", err)
	}
	if err := kubelet.Run(ctx, kubeletServer, kubeletDeps, utilfeature.DefaultFeatureGate); err != nil {
		logrus.Error("Kubelet failed to start %v", err)
	}
	return ctx.Err()
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
