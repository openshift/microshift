/*
Copyright © 2021 MicroShift Contributors

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

	corev1 "k8s.io/api/core/v1"
	unstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"

	configv1 "github.com/openshift/api/config/v1"
	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	"github.com/openshift/api/operator/v1alpha1"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	route_controller_manager "github.com/openshift/route-controller-manager/pkg/cmd/route-controller-manager"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

type OCPRouteControllerManager struct {
	run           func(context.Context) error
	kubeconfig    string
	kubeadmconfig string

	configErr error
}

const (
	// OCPRouteControllerManager component name
	componentRCM = "route-controller-manager"
)

func NewRouteControllerManager(cfg *config.Config) *OCPRouteControllerManager {
	s := &OCPRouteControllerManager{}
	s.configErr = s.configure(cfg)
	return s
}

func (s *OCPRouteControllerManager) Name() string { return componentRCM }
func (s *OCPRouteControllerManager) Dependencies() []string {
	return []string{"kube-apiserver", "openshift-crd-manager"}
}

func (s *OCPRouteControllerManager) configure(cfg *config.Config) error {
	s.kubeconfig = cfg.KubeConfigPath(config.RouteControllerManager)
	s.kubeadmconfig = cfg.KubeConfigPath(config.KubeAdmin)

	servingCertDir := cryptomaterial.RouteControllerManagerServingCertDir(cryptomaterial.CertsDirectory(config.DataDir))
	rcmConfig := &openshiftcontrolplanev1.OpenShiftControllerManagerConfig{
		ServingInfo: &configv1.HTTPServingInfo{
			ServingInfo: configv1.ServingInfo{
				BindAddress: "0.0.0.0:8445",
				BindNetwork: "tcp",
				CertInfo: configv1.CertInfo{
					CertFile: cryptomaterial.ServingCertPath(servingCertDir),
					KeyFile:  cryptomaterial.ServingKeyPath(servingCertDir),
				},
				ClientCA: cryptomaterial.TotalClientCABundlePath(cryptomaterial.CertsDirectory(config.DataDir)),
			},
		},
		Controllers: []string{
			"openshift.io/ingress-to-route",
			"-openshift.io/ingress-ip",
		},
	}

	scheme := runtime.NewScheme()
	if err := openshiftcontrolplanev1.AddToScheme(scheme); err != nil {
		return err
	}
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(openshiftcontrolplanev1.GroupVersion)
	encodedConfig, err := runtime.Encode(codec, rcmConfig)
	if err != nil {
		return err
	}
	ctrlConfig := &unstructuredv1.Unstructured{}
	if err := runtime.DecodeInto(codec, encodedConfig, ctrlConfig); err != nil {
		return err
	}

	const namespace = "openshift-route-controller-manager"
	builder := controllercmd.NewController(s.Name(), route_controller_manager.RunRouteControllerManager).
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
		WithServer(*rcmConfig.ServingInfo, v1alpha1.DelegatedAuthentication{}, v1alpha1.DelegatedAuthorization{})

	s.run = func(ctx context.Context) error {
		return builder.Run(ctx, ctrlConfig)
	}

	return nil
}

func (s *OCPRouteControllerManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	if s.configErr != nil {
		return fmt.Errorf("configuration failed: %w", s.configErr)
	}

	if err := assets.ApplyNamespaces(ctx, []string{
		"controllers/route-controller-manager/route-controller-manager-ns.yaml",
		"controllers/route-controller-manager/ns.yaml",
	}, s.kubeadmconfig); err != nil {
		return fmt.Errorf("failed to apply openshift namespaces: %w", err)
	}
	if err := assets.ApplyClusterRoles(ctx, []string{
		"controllers/route-controller-manager/informer-clusterrole.yaml",
		"controllers/route-controller-manager/route-controller-manager-tokenreview-clusterrole.yaml",
		"controllers/route-controller-manager/route-controller-manager-ingress-to-route-controller-clusterrole.yaml",
		"controllers/route-controller-manager/route-controller-manager-clusterrole.yaml",
	}, s.kubeadmconfig); err != nil {
		return fmt.Errorf("failed to apply route controller manager cluster roles: %w", err)
	}

	if err := assets.ApplyClusterRoleBindings(ctx, []string{
		"controllers/route-controller-manager/informer-clusterrolebinding.yaml",
		"controllers/route-controller-manager/route-controller-manager-tokenreview-clusterrolebinding.yaml",
		"controllers/route-controller-manager/route-controller-manager-ingress-to-route-controller-clusterrolebinding.yaml",
		"controllers/route-controller-manager/route-controller-manager-clusterrolebinding.yaml",
	}, s.kubeadmconfig); err != nil {
		return fmt.Errorf("failed to apply route controller manager cluster role bindings: %w", err)
	}

	if err := assets.ApplyRoles(ctx, []string{
		"controllers/route-controller-manager/route-controller-manager-separate-sa-role.yaml",
	}, s.kubeadmconfig); err != nil {
		return fmt.Errorf("failed to apply route controller manager roles: %w", err)
	}

	if err := assets.ApplyRoleBindings(ctx, []string{
		"controllers/route-controller-manager/route-controller-manager-authentication-reader-rolebinding.yaml",
		"controllers/route-controller-manager/route-controller-manager-separate-sa-rolebinding.yaml",
	}, s.kubeadmconfig); err != nil {
		return fmt.Errorf("failed to apply route controller manager role bindings: %w", err)
	}

	if err := assets.ApplyServiceAccounts(ctx, []string{
		"controllers/route-controller-manager/route-controller-manager-sa.yaml",
		"controllers/route-controller-manager/sa.yaml",
	}, s.kubeadmconfig); err != nil {
		return fmt.Errorf("failed to apply route controller manager service account: %w", err)
	}

	// Run healthcheck probe and controller in parallel.
	// No matter which ends first - if it ends with an error,
	// it'll cause ServiceManager to trigger graceful shutdown.

	errc := make(chan error)

	go func() {
		healthcheckStatus := util.RetryTCPConnection(ctx, "localhost", "8445")
		if !healthcheckStatus {
			e := fmt.Errorf("initial healthcheck on %s failed", s.Name())
			klog.Error(e)
			errc <- e
			return
		}
		klog.Infof("%s is ready", s.Name())
		close(ready)
	}()

	panicChannel := make(chan any, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicChannel <- r
			}
		}()
		errc <- s.run(ctx)
	}()

	select {
	case err := <-errc:
		return err
	case perr := <-panicChannel:
		panic(perr)
	}
}
