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

	klog "k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/components"
	"github.com/openshift/microshift/pkg/config"
)

type InfrastructureServicesManager struct {
	cfg *config.MicroshiftConfig
}

func NewInfrastructureServices(cfg *config.MicroshiftConfig) *InfrastructureServicesManager {
	s := &InfrastructureServicesManager{}
	s.cfg = cfg
	return s
}

func (s *InfrastructureServicesManager) Name() string { return "infrastructure-services-manager" }
func (s *InfrastructureServicesManager) Dependencies() []string {
	return []string{"kube-apiserver", "openshift-crd-manager", "route-controller-manager"}
}

func (s *InfrastructureServicesManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(ready)

	if err := applyDefaultRBACs(s.cfg); err != nil {
		klog.Errorf("%s unable to apply default RBACs: %v", s.Name(), err)
		return err
	}

	priorityClasses := []string{"core/priority-class-openshift-user-critical.yaml"}
	if err := assets.ApplyPriorityClasses(priorityClasses, s.cfg.KubeConfigPath(config.KubeAdmin)); err != nil {
		klog.Errorf("%s unable to apply PriorityClasses: %v", s.Name(), err)
		return err
	}

	// TO-DO add readiness check
	if err := components.StartComponents(s.cfg); err != nil {
		return err
	}
	klog.Infof("%s launched ocp componets", s.Name())
	return ctx.Err()
}

func applyDefaultRBACs(cfg *config.MicroshiftConfig) error {
	kubeconfigPath := cfg.KubeConfigPath(config.KubeAdmin)
	var (
		cr = []string{
			"core/csr_approver_clusterrole.yaml",
			"core/namespace-security-allocation-controller-clusterrole.yaml",
			"core/podsecurity-admission-label-syncer-controller-clusterrole.yaml",
		}
		crb = []string{
			"core/csr_approver_clusterrolebinding.yaml",
			"core/namespace-security-allocation-controller-clusterrolebinding.yaml",
			"core/podsecurity-admission-label-syncer-controller-clusterrolebinding.yaml",
		}
	)
	if err := assets.ApplyClusterRoles(cr, kubeconfigPath); err != nil {
		klog.Warningf("failed to apply cluster roles %v", err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(crb, kubeconfigPath); err != nil {
		klog.Warningf("failed to apply cluster roles %v", err)
		return err
	}
	return nil
}
