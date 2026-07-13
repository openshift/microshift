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
	"time"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	securityv1 "github.com/openshift/api/security/v1"
	clusterpolicycontroller "github.com/openshift/cluster-policy-controller/pkg/cmd/cluster-policy-controller"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

type ClusterPolicyController struct {
	run             func(context.Context) error
	kubeconfig      string
	adminKubeconfig string

	configErr error
}

func NewClusterPolicyController(cfg *config.Config) *ClusterPolicyController {
	s := &ClusterPolicyController{}
	s.configErr = s.configure(cfg)
	return s
}

func (s *ClusterPolicyController) Name() string { return "cluster-policy-controller" }
func (s *ClusterPolicyController) Dependencies() []string {
	return []string{"kube-apiserver"}
}

func (s *ClusterPolicyController) configure(cfg *config.Config) error {
	s.kubeconfig = cfg.KubeConfigPath(config.ClusterPolicyController)
	s.adminKubeconfig = cfg.KubeConfigPath(config.KubeAdmin)

	scheme := runtime.NewScheme()
	if err := openshiftcontrolplanev1.AddToScheme(scheme); err != nil {
		return err
	}

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
	builder := controllercmd.NewController(s.Name(), clusterpolicycontroller.RunClusterPolicyController, clock.RealClock{}).
		WithKubeConfigFile(s.kubeconfig, nil).
		WithComponentNamespace(namespace).
		// Without an explicit owner reference, the builder will try using POD_NAME or the
		// first pod in the target namespace (and fail because we have no pod).
		WithComponentOwnerReference(&corev1.ObjectReference{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
			Name:       namespace,
			Namespace:  namespace,
		}).
		WithEventRecorderOptions(events.RecommendedClusterSingletonCorrelatorOptions())

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

	if err := applyClusterPolicyControllerRBAC(ctx, s.adminKubeconfig); err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.run(ctx)
	}()

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- waitForNamespaceSecurityAllocation(ctx, s.kubeconfig)
	}()

	select {
	case err := <-errCh:
		return err
	case err := <-waitCh:
		if err != nil {
			return err
		}
	}

	klog.Infof("%s is ready", s.Name())
	close(ready)

	return <-errCh
}

func applyClusterPolicyControllerRBAC(ctx context.Context, kubeconfigPath string) error {
	cr := []string{
		"controllers/cluster-policy-controller/namespace-security-allocation-controller-clusterrole.yaml",
		"controllers/cluster-policy-controller/podsecurity-admission-label-syncer-controller-clusterrole.yaml",
		"controllers/cluster-policy-controller/podsecurity-admission-label-privileged-namespaces-syncer-controller-clusterrole.yaml",
	}
	crb := []string{
		"controllers/cluster-policy-controller/namespace-security-allocation-controller-clusterrolebinding.yaml",
		"controllers/cluster-policy-controller/podsecurity-admission-label-syncer-controller-clusterrolebinding.yaml",
		"controllers/cluster-policy-controller/podsecurity-admission-label-privileged-namespaces-syncer-controller-clusterrolebinding.yaml",
	}
	if err := assets.ApplyClusterRoles(ctx, cr, kubeconfigPath); err != nil {
		return fmt.Errorf("failed to apply cluster-policy-controller cluster roles: %w", err)
	}
	if err := assets.ApplyClusterRoleBindings(ctx, crb, kubeconfigPath); err != nil {
		return fmt.Errorf("failed to apply cluster-policy-controller cluster role bindings: %w", err)
	}
	return nil
}

// waitForNamespaceSecurityAllocation waits until the UID range annotation exists on the
// default namespace. On first boot this annotation is set by the namespace-security-allocation-controller;
// on restart it persists from the previous boot. Either way, its presence guarantees that
// SCC admission will not block when infrastructure-services-manager creates deployments.
func waitForNamespaceSecurityAllocation(ctx context.Context, kubeconfigPath string) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig for readiness check: %w", err)
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create kube client for readiness check: %w", err)
	}

	attempt := 0
	return wait.PollUntilContextCancel(ctx, time.Second, true, func(ctx context.Context) (bool, error) {
		attempt++
		ns, err := kubeClient.CoreV1().Namespaces().Get(ctx, "default", metav1.GetOptions{})
		if err != nil {
			if attempt%5 == 0 {
				klog.Infof("cluster-policy-controller: still waiting for namespace security allocation (attempt %d): %v", attempt, err)
			}
			return false, nil
		}
		if _, ok := ns.Annotations[securityv1.UIDRangeAnnotation]; !ok {
			if attempt%5 == 0 {
				klog.Infof("cluster-policy-controller: still waiting for %s annotation on default namespace (attempt %d)", securityv1.UIDRangeAnnotation, attempt)
			}
			return false, nil
		}
		return true, nil
	})
}
