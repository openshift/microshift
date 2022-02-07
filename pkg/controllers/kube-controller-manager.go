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
package controllers

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"

	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/version/verflag"
	kubecm "k8s.io/kubernetes/cmd/kube-controller-manager/app"
	kubecmoptions "k8s.io/kubernetes/cmd/kube-controller-manager/app/options"

	klog "k8s.io/klog/v2"
)

type KubeControllerManager struct {
	kubecmOptions *kubecmoptions.KubeControllerManagerOptions
	kubeconfig    string
}

func NewKubeControllerManager(cfg *config.MicroshiftConfig) *KubeControllerManager {
	s := &KubeControllerManager{}
	s.configure(cfg)
	return s
}

func (s *KubeControllerManager) Name() string           { return "kube-controller-manager" }
func (s *KubeControllerManager) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *KubeControllerManager) configure(cfg *config.MicroshiftConfig) {
	caCertFile := filepath.Join(cfg.DataDir, "certs", "ca-bundle", "ca-bundle.crt")
	kubeconfig := filepath.Join(cfg.DataDir, "resources", "kube-controller-manager", "kubeconfig")

	opts, err := kubecmoptions.NewKubeControllerManagerOptions()
	if err != nil {
		klog.Fatalf("%s initialization error command options: %v", s.Name(), err)
	}
	s.kubecmOptions = opts
	s.kubeconfig = kubeconfig

	args := []string{
		"--kubeconfig=" + kubeconfig,
		"--service-account-private-key-file=" + cfg.DataDir + "/resources/kube-apiserver/secrets/service-account-key/service-account.key",
		"--allocate-node-cidrs=true",
		"--cluster-cidr=" + cfg.Cluster.ClusterCIDR,
		"--authorization-kubeconfig=" + kubeconfig,
		"--authentication-kubeconfig=" + kubeconfig,
		"--root-ca-file=" + caCertFile,
		"--bind-address=127.0.0.1",
		"--secure-port=10257",
		"--leader-elect=false",
		"--use-service-account-credentials=true",
		"--cluster-signing-cert-file=" + caCertFile,
		"--cluster-signing-key-file=" + cfg.DataDir + "/certs/ca-bundle/ca-bundle.key",
	}

	// fake the kube-controller-manager cobra command to parse args into controllermanager options
	cmd := &cobra.Command{
		Use:          "kube-controller-manager",
		Long:         `kube-controller-manager`,
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return nil },
	}

	namedFlagSets := s.kubecmOptions.Flags(kubecm.KnownControllers(), kubecm.ControllersDisabledByDefault.List())
	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	for _, f := range namedFlagSets.FlagSets {
		cmd.Flags().AddFlagSet(f)
	}
	if err := cmd.ParseFlags(args); err != nil {
		klog.Fatalf("%s failed to parse flags: %v", s.Name(), err)
	}
}

func (s *KubeControllerManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	errorChannel := make(chan error, 1)

	// run readiness check
	go func() {
		healthcheckStatus := util.RetryInsecureHttpsGet("https://127.0.0.1:10257/healthz")
		if healthcheckStatus != 200 {
			klog.Errorf("kube-controller-manager failed to start")
			errorChannel <- errors.New("kube-controller-manager failed to start")
		}

		klog.Infof("%s is ready", s.Name())
		close(ready)
	}()

	c, err := s.kubecmOptions.Config(kubecm.KnownControllers(), kubecm.ControllersDisabledByDefault.List())
	if err != nil {
		return err
	}

	// TODO: OpenShift's kubecm patch, uncomment if OpenShiftContext added
	//if err := kubecm.ShimForOpenShift(s.kubecmOptions, c); err != nil {
	//	return err
	//}

	go func() {
		errorChannel <- kubecm.Run(c.Complete(), ctx.Done())
	}()

	return <-errorChannel
}
