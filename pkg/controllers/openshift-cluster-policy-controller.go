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
	"github.com/openshift/library-go/pkg/config/client"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/microshift/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (
	podNameEnv         = "POD_NAME"
	podNamespaceEnv    = "POD_NAMESPACE"
	componentName      = "cluster-policy-controller"
	componentNamespace = "openshift-kube-controller-manager"
)

// OpenShift ClusterPolicyController Service
type OpenShiftClusterPolicyController struct {
	controllerFlags *controllercmd.ControllerFlags
	kubeconfig      string
}

func NewOpenShiftClusterPolicyController(cfg *config.MicroshiftConfig) *OpenShiftClusterPolicyController {
	s := &OpenShiftClusterPolicyController{}
	s.configure(cfg)
	return s
}

func (s *OpenShiftClusterPolicyController) Name() string {
	return "openshift-cluster-policy-controller"
}
func (s *OpenShiftClusterPolicyController) Dependencies() []string {
	return []string{"kube-apiserver", "openshift-controller-manager", "ocp-apiserver"}
}

func (s *OpenShiftClusterPolicyController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	klog.Infof("starting openshift-cluster-policy-controller")
	_, unstructuredConfig, err := s.controllerFlags.ToConfigObj()
	if err != nil {
		return err
	}
	clientConfig, err := client.GetKubeConfigOrInClusterConfig(s.kubeconfig, nil)
	if err != nil {
		return err
	}
	controllerRef := &corev1.ObjectReference{
		Kind:      "Pod",
		Name:      os.Getenv(podNameEnv),
		Namespace: os.Getenv(podNamespaceEnv),
	}
	kubeClient := kubernetes.NewForConfigOrDie(clientConfig)
	eventRecorder := events.NewKubeRecorderWithOptions(kubeClient.CoreV1().Events(componentNamespace), events.RecommendedClusterSingletonCorrelatorOptions(), componentName, controllerRef)
	protoConfig := rest.CopyConfig(clientConfig)
	protoConfig.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"
	protoConfig.ContentType = "application/vnd.kubernetes.protobuf"
	controllerContext := &controllercmd.ControllerContext{
		ComponentConfig: unstructuredConfig,
		KubeConfig:      clientConfig,
		ProtoKubeConfig: protoConfig,
		EventRecorder:   eventRecorder,
	}
	if err := clusterpolicycontroller.RunClusterPolicyController(ctx, controllerContext); err != nil {
		return err
	}
	return ctx.Err()
}

func (s *OpenShiftClusterPolicyController) configure(cfg *config.MicroshiftConfig) error {
	controllerFlags := controllercmd.NewControllerFlags()
	kubeconfig := filepath.Join(cfg.DataDir, "resources", "kubeadmin", "kubeconfig")
	s.kubeconfig = kubeconfig
	path := filepath.Join(cfg.DataDir, "resources", "openshift-cluster-policy-controller", "config", "config.yaml")
	controllerFlags.ConfigFile = path
	controllerFlags.KubeConfigFile = kubeconfig
	controllerFlags.BindAddress = "0.0.0.0:10357"
	s.controllerFlags = controllerFlags

	data := []byte(`apiVersion: openshiftcontrolplane.config.openshift.io/v1
kind: OpenShiftControllerManagerConfig
kubeClientConfig:
  kubeConfig: ` + cfg.DataDir + `/resources/kubeadmin/kubeconfig
servingInfo:
  bindAddress: "0.0.0.0:10357"
  certFile: ` + cfg.DataDir + `/resources/openshift-cluster-policy-controller/secrets/tls.crt
  keyFile:  ` + cfg.DataDir + `/resources/openshift-cluster-policy-controller/secrets/tls.key
  clientCA: ` + cfg.DataDir + `/certs/ca-bundle/ca-bundle.crt`)

	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}
