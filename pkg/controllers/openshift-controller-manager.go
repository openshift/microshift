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
	"time"

	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/library-go/pkg/config/helpers"
	"k8s.io/klog/v2"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	openshift_controller_manager "github.com/openshift/openshift-controller-manager/pkg/cmd/openshift-controller-manager"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

type OCPControllerManager struct {
	kubeconfig string
	config     openshiftcontrolplanev1.OpenShiftControllerManagerConfig
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
	return []string{"kube-apiserver", "openshift-crd-manager"}
}

func (s *OCPControllerManager) configure(cfg *config.MicroshiftConfig) {
	s.kubeconfig = cfg.KubeConfigPath(config.KubeAdmin)
	servingCertDir := cryptomaterial.OpenshiftControllerManagerServingCertDir(cryptomaterial.CertsDirectory(cfg.DataDir))

	ocmconfig := openshiftcontrolplanev1.OpenShiftControllerManagerConfig{
		KubeClientConfig: configv1.KubeClientConfig{
			KubeConfig: s.kubeconfig,
		},
		Controllers: []string{
			"*",
			"-openshift.io/build",
			"-openshift.io/build-config-change",
			"-openshift.io/default-rolebindings",
			"-openshift.io/deployer",
			"-openshift.io/deploymentconfig",
			"-openshift.io/image-import",
			"-openshift.io/image-signature-import",
			"-openshift.io/image-trigger",
			"-openshift.io/origin-namespace",
			"-openshift.io/serviceaccount",
			"-openshift.io/serviceaccount-pull-secrets",
			"-openshift.io/templateinstance",
			"-openshift.io/templateinstancefinalizer",
			"-openshift.io/unidling",
		},
		LeaderElection: configv1.LeaderElection{
			Disable:       true,
			LeaseDuration: metav1.Duration{Duration: 270 * time.Second},
			RenewDeadline: metav1.Duration{Duration: 240 * time.Second},
			RetryPeriod:   metav1.Duration{Duration: 60 * time.Second},
		},
		ServingInfo: &configv1.HTTPServingInfo{
			ServingInfo: configv1.ServingInfo{
				BindAddress: "0.0.0.0:8445",
				BindNetwork: "tcp4",
				ClientCA:    cryptomaterial.TotalClientCABundlePath(cryptomaterial.CertsDirectory(cfg.DataDir)),
				CertInfo: configv1.CertInfo{
					CertFile: cryptomaterial.ServingCertPath(servingCertDir),
					KeyFile:  cryptomaterial.ServingKeyPath(servingCertDir),
				},
			},
		},
	}

	s.config = ocmconfig

}

func (s *OCPControllerManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	errorChannel := make(chan error, 1)

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

	clientConfig, err := helpers.GetKubeClientConfig(configv1.KubeClientConfig{KubeConfig: s.kubeconfig})
	if err != nil {
		return err
	}

	go func(ctx context.Context) error {

		errorChannel <- openshift_controller_manager.RunOpenShiftControllerManager(&s.config, clientConfig)

		select {
		case <-ctx.Done():
			klog.Infof("OpenShift Controller Manager received stop signal. exiting.")
			return nil
		}
	}(ctx)

	return <-errorChannel
}
