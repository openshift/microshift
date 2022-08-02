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
	"io/ioutil"
	"os"
	"path/filepath"

	clusterpolicycontroller "github.com/openshift/cluster-policy-controller/pkg/cmd/cluster-policy-controller"
	clusterpolicyversion "github.com/openshift/cluster-policy-controller/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/microshift/pkg/config"
	"k8s.io/component-base/cli"
)

type ClusterPolicyController struct {
	config     *controllercmd.ControllerCommandConfig
	flags      *controllercmd.ControllerFlags
	kubeconfig string
}

func NewClusterPolicyController(cfg *config.MicroshiftConfig) *ClusterPolicyController {
	s := &ClusterPolicyController{}
	s.configure(cfg)
	return s
}

func (s *ClusterPolicyController) Name() string           { return "cluster-policy-controller" }
func (s *ClusterPolicyController) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *ClusterPolicyController) configure(cfg *config.MicroshiftConfig) {

	s.kubeconfig = filepath.Join(cfg.DataDir, "resources", "kubeadmin", "kubeconfig")
	s.writeConfig(cfg)

	// Use NewControllerCommandConfig to create a controller configuration struct
	// with default values.
	s.config = controllercmd.NewControllerCommandConfig(s.Name(), clusterpolicyversion.Get(), clusterpolicycontroller.RunClusterPolicyController)

	flags := controllercmd.NewControllerFlags()
	flags.ConfigFile = filepath.Join(cfg.DataDir, "resources", "cluster-policy-controller", "config", "config.yaml")
	flags.KubeConfigFile = s.kubeconfig
	flags.Validate()
	s.flags = flags

}

func (s *ClusterPolicyController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	cmd := s.config.NewCommandWithContext(ctx)
	s.flags.AddFlags(cmd)
	cmd.Use = s.Name()
	cmd.Short = "Start the cluster-policy-controller"

	go func() {
		cli.Run(cmd)
	}()

	<-ctx.Done()
	return ctx.Err()
}

func (s *ClusterPolicyController) writeConfig(cfg *config.MicroshiftConfig) error {
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
controllers:
- "-openshift.io/resourcequota"
- "-openshift.io/cluster-quota-reconciliation"
`)

	path := filepath.Join(cfg.DataDir, "resources", "cluster-policy-controller", "config", "config.yaml")
	os.MkdirAll(filepath.Dir(path), os.FileMode(0700))
	return ioutil.WriteFile(path, data, 0644)
}
