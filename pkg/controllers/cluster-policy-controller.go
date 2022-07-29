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
	"path/filepath"

	clusterpolicycontroller "github.com/openshift/cluster-policy-controller/pkg/cmd/cluster-policy-controller"
	clusterpolicyversion "github.com/openshift/cluster-policy-controller/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/microshift/pkg/config"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type ClusterPolicyController struct {
	config     *controllercmd.ControllerCommandConfig
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
	// Use NewControllerCommandConfig to create a controller configuration struct
	// with default values.
	s.config = controllercmd.NewControllerCommandConfig(s.Name(), clusterpolicyversion.Get(), clusterpolicycontroller.RunClusterPolicyController)
}

func (s *ClusterPolicyController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	// run readiness check
	restConfig, err := clientcmd.BuildConfigFromFlags("", s.kubeconfig)
	if err != nil {
		return err
	}
	if err := rest.SetKubernetesDefaults(restConfig); err != nil {
		return err
	}
	restConfig.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())

	kubeClient, err := kubernetes.NewForConfig(restConfig)

	if err := clusterpolicycontroller.WaitForHealthyAPIServer(kubeClient.Discovery().RESTClient()); err != nil {
		klog.Fatal(err)
		return err
	}

	go func() {
		s.config.StartController(ctx)
	}()

	<-ctx.Done()
	return ctx.Err()
}
