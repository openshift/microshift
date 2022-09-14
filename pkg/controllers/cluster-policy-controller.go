/*
Copyright © 2022 MicroShift Contributors
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
	"fmt"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	clusterpolicycontroller "github.com/openshift/cluster-policy-controller/pkg/cmd/cluster-policy-controller"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/microshift/pkg/config"
	corev1 "k8s.io/api/core/v1"
	unstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type ClusterPolicyController struct {
	run        func(context.Context) error
	kubeconfig string

	configErr error
}

func NewClusterPolicyController(cfg *config.MicroshiftConfig) *ClusterPolicyController {
	s := &ClusterPolicyController{}
	s.configErr = s.configure(cfg)
	return s
}

func (s *ClusterPolicyController) Name() string           { return "cluster-policy-controller" }
func (s *ClusterPolicyController) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *ClusterPolicyController) configure(cfg *config.MicroshiftConfig) error {
	s.kubeconfig = cfg.KubeConfigPath(config.KubeControllerManager)

	scheme := runtime.NewScheme()
	openshiftcontrolplanev1.AddToScheme(scheme)
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(openshiftcontrolplanev1.GroupVersion)

	encodedConfig, err := runtime.Encode(codec,
		&openshiftcontrolplanev1.OpenShiftControllerManagerConfig{
			Controllers: []string{
				"*",
				"-openshift.io/resourcequota",
				"-openshift.io/cluster-quota-reconciliation",
			},
		})
	if err != nil {
		return err
	}
	ctrlConfig := &unstructuredv1.Unstructured{}
	if err := runtime.DecodeInto(codec, encodedConfig, ctrlConfig); err != nil {
		return err
	}

	const namespace = "openshift-kube-controller-manager"
	builder := controllercmd.NewController(s.Name(), clusterpolicycontroller.RunClusterPolicyController).
		WithKubeConfigFile(s.kubeconfig, nil).
		WithComponentNamespace(namespace).
		// Without an explicit owner reference, the builder will try using POD_NAME or the
		// first pod in the target namespace (and fail because we have no pod).
		WithComponentOwnerReference(&corev1.ObjectReference{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
			Name:       namespace,
			Namespace:  namespace,
		})

	s.run = func(ctx context.Context) error {
		return builder.Run(ctx, ctrlConfig)
	}

	return nil
}

func (s *ClusterPolicyController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	if s.configErr != nil {
		return fmt.Errorf("configuration failed: %w", s.configErr)
	}

	close(ready) // todo
	return s.run(ctx)
}
