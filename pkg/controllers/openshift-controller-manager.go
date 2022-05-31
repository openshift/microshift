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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	openshift_controller_manager "github.com/openshift/openshift-controller-manager/pkg/cmd/openshift-controller-manager"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
)

type OCPControllerManager struct {
	kubeconfig     string
	ConfigFilePath string
	Output         io.Writer
}

const (
	// OCPControllerManager component name
	componentOCM = "openshift-controller-manager"
)

func NewOpenShiftControllerManager(cfg *config.MicroshiftConfig) *OCPControllerManager {
	s := &OCPControllerManager{}
	s.configure(cfg)
	return s
}

func (s *OCPControllerManager) Name() string { return componentOCM }
func (s *OCPControllerManager) Dependencies() []string {
	return []string{"kube-apiserver", "openshift-apiserver"}
}

func (s *OCPControllerManager) configure(cfg *config.MicroshiftConfig) {
	if err := s.writeConfig(cfg); err != nil {
		klog.Fatalf("Failed to write openshift-controller-manager config %v", err)
	}

	var configFilePath = cfg.DataDir + "/resources/openshift-controller-manager/config/config.yaml"
	args := []string{
		"--config=" + configFilePath,
	}

	options := openshift_controller_manager.OpenShiftControllerManager{Output: os.Stdout}
	options.ConfigFilePath = configFilePath

	cmd := &cobra.Command{
		Use:          componentOCM,
		Long:         componentOCM,
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return nil },
	}

	flags := cmd.Flags()
	cmd.SetArgs(args)
	flags.StringVar(&options.ConfigFilePath, "config", options.ConfigFilePath, "Location of the master configuration file to run from.")
	cmd.MarkFlagFilename("config", "yaml", "yml")
	cmd.MarkFlagRequired("config")

	s.kubeconfig = filepath.Join(cfg.DataDir, "resources", "kubeadmin", "kubeconfig")
	s.ConfigFilePath = options.ConfigFilePath
	s.Output = options.Output
}

func (s *OCPControllerManager) writeConfig(cfg *config.MicroshiftConfig) error {
	// OCM config contains a list of controllers to enable/disable.
	// If no list is specified, all controllers are started (default).
	// If a non-zero length list is specified, only controllers enabled in the list are started.  Unlisted controllers
	// are therefore disabled.  Enable controllers by appending their name to `controllers:`. Disable a controller by
	// prepending "-" to the name, e.g. `controllers: ["-openshift.io/build"]
	// Disabled OCM controllers are included in the list for documentary purposes.
	data := []byte(`apiVersion: openshiftcontrolplane.config.openshift.io/v1
kind: OpenShiftControllerManagerConfig
kubeClientConfig:
  kubeConfig: ` + cfg.DataDir + `/resources/kubeadmin/kubeconfig
servingInfo:
  bindAddress: "0.0.0.0:8445"
  certFile: ` + cfg.DataDir + `/resources/openshift-controller-manager/secrets/tls.crt
  keyFile:  ` + cfg.DataDir + `/resources/openshift-controller-manager/secrets/tls.key
  clientCA: ` + cfg.DataDir + `/certs/ca-bundle/ca-bundle.crt
controllers:
- "openshift.io/ingress-ip"
- "openshift.io/ingress-to-route"
- "-openshift.io/build"
- "-openshift.io/build-config-change"
- "-openshift.io/default-rolebindings"
- "-openshift.io/deployer"
- "-openshift.io/deploymentconfig"
- "-openshift.io/image-import"
- "-openshift.io/image-signature-import"
- "-openshift.io/image-trigger"
- "-openshift.io/origin-namespace"
- "-openshift.io/serviceaccount"
- "-openshift.io/serviceaccount-pull-secrets"
- "-openshift.io/templateinstance"
- "-openshift.io/templateinstancefinalizer"
- "-openshift.io/unidling"
`)

	path := filepath.Join(cfg.DataDir, "resources", "openshift-controller-manager", "config", "config.yaml")
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}

func (s *OCPControllerManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	// run readiness check
	go func() {
		healthcheckStatus := util.RetryTCPConnection("127.0.0.1", "8445")
		if !healthcheckStatus {
			klog.Fatalf("initial healthcheck on %s failed", s.Name())
		}
		klog.Infof("%s is ready", s.Name())
		close(ready)
	}()

	if err := assets.ApplyNamespaces([]string{
		"assets/core/0000_50_cluster-openshift-controller-manager_00_namespace.yaml",
	}, s.kubeconfig); err != nil {
		klog.Fatalf("failed to apply openshift namespaces %v", err)
	}

	options := openshift_controller_manager.OpenShiftControllerManager{Output: os.Stdout}
	options.ConfigFilePath = s.ConfigFilePath

	go func() {
		if err := options.StartControllerManager(); err != nil {
			klog.Fatalf("Failed to start openshift-controller-manager %v", err)
		}
	}()

	<-ctx.Done()
	return ctx.Err()
}
