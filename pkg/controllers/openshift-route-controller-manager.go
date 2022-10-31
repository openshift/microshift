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

	"k8s.io/klog/v2"

	configv1 "github.com/openshift/api/config/v1"
	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	"github.com/openshift/library-go/pkg/config/configdefaults"
	"github.com/openshift/library-go/pkg/config/helpers"
	"github.com/openshift/library-go/pkg/config/leaderelection"
	route_controller_manager "github.com/openshift/route-controller-manager/pkg/cmd/route-controller-manager"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

type OCPRouteControllerManager struct {
	kubeconfig string
	config     *openshiftcontrolplanev1.OpenShiftControllerManagerConfig
}

const (
	// OCPRouteControllerManager component name
	componentRCM = "route-controller-manager"
)

func NewRouteControllerManager(cfg *config.MicroshiftConfig) *OCPRouteControllerManager {
	s := &OCPRouteControllerManager{}
	s.configure(cfg)
	return s
}

func (s *OCPRouteControllerManager) Name() string { return componentRCM }
func (s *OCPRouteControllerManager) Dependencies() []string {
	return []string{"kube-apiserver", "openshift-crd-manager"}
}

func (s *OCPRouteControllerManager) configure(cfg *config.MicroshiftConfig) {
	s.kubeconfig = cfg.KubeConfigPath(config.KubeAdmin)
	s.config = s.writeConfig(cfg)
}

func (s *OCPRouteControllerManager) writeConfig(cfg *config.MicroshiftConfig) *openshiftcontrolplanev1.OpenShiftControllerManagerConfig {
	servingCertDir := cryptomaterial.RouteControllerManagerServingCertDir(cryptomaterial.CertsDirectory(microshiftDataDir))

	c := &openshiftcontrolplanev1.OpenShiftControllerManagerConfig{
		KubeClientConfig: configv1.KubeClientConfig{
			KubeConfig: s.kubeconfig,
			ConnectionOverrides: configv1.ClientConnectionOverrides{
				ContentType: "application/json",
			},
		},
		ServingInfo: &configv1.HTTPServingInfo{
			ServingInfo: configv1.ServingInfo{
				BindAddress: "0.0.0.0:8445",
				BindNetwork: "tcp4",
				CertInfo: configv1.CertInfo{
					CertFile: cryptomaterial.ServingCertPath(servingCertDir),
					KeyFile:  cryptomaterial.ServingKeyPath(servingCertDir),
				},
				ClientCA: cryptomaterial.TotalClientCABundlePath(cryptomaterial.CertsDirectory(microshiftDataDir)),
			},
		},
		Controllers: []string{
			"openshift.io/ingress-to-route",
			"-openshift.io/ingress-ip",
		},
	}

	// https://github.com/openshift/route-controller-manager/blob/main/pkg/cmd/route-controller-manager/openshiftcontrolplane_default.go
	configdefaults.SetRecommendedHTTPServingInfoDefaults(c.ServingInfo)
	configdefaults.SetRecommendedKubeClientConfigDefaults(&c.KubeClientConfig)
	c.LeaderElection = leaderelection.LeaderElectionDefaulting(c.LeaderElection, "kube-system", "openshift-route-controllers")
	return c
}

func (s *OCPRouteControllerManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
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
		"core/0000_50_cluster-openshift-route-controller-manager_00_namespace.yaml",
	}, s.kubeconfig); err != nil {
		klog.Fatalf("failed to apply openshift namespaces %v", err)
	}
	clientConfig, err := helpers.GetKubeClientConfig(s.config.KubeClientConfig)
	if err != nil {
		return err
	}

	return route_controller_manager.RunRouteControllerManager(s.config, clientConfig, ctx)
}
