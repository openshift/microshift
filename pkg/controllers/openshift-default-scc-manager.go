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

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"k8s.io/klog/v2"
)

type OpenShiftDefaultSCCManager struct {
	cfg *config.Config
}

func NewOpenShiftDefaultSCCManager(cfg *config.Config) *OpenShiftDefaultSCCManager {
	s := &OpenShiftDefaultSCCManager{}
	s.cfg = cfg
	return s
}

func (s *OpenShiftDefaultSCCManager) Name() string {
	return "openshift-default-scc-manager"
}
func (s *OpenShiftDefaultSCCManager) Dependencies() []string {
	return []string{"kube-apiserver", "openshift-crd-manager"}
}

func (s *OpenShiftDefaultSCCManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	defer close(ready)
	// TO-DO add readiness check
	if err := ApplyDefaultSCCs(ctx, s.cfg); err != nil {
		klog.Errorf("%s unable to apply default SCCs: %v", s.Name(), err)
		return err
	}
	klog.Infof("%s applied default SCCs", s.Name())
	return ctx.Err()
}

func ApplyDefaultSCCs(ctx context.Context, cfg *config.Config) error {
	kubeconfigPath := cfg.KubeConfigPath(config.KubeAdmin)
	var (
		clusterRole = []string{
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_cr-scc-anyuid.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_cr-scc-hostaccess.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_cr-scc-hostmount-anyuid.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_cr-scc-hostnetwork-v2.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_cr-scc-hostnetwork.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_cr-scc-nonroot-v2.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_cr-scc-nonroot.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_cr-scc-privileged.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_cr-scc-restricted-v2.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_cr-scc-restricted.yaml",
		}
		clusterRoleBinding = []string{
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_crb-systemauthenticated-scc-restricted-v2.yaml",
		}
		sccs = []string{
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_scc-hostnetwork-v2.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_scc-nonroot-v2.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_scc-privileged.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_scc-restricted-v2.yaml",
			"controllers/openshift-default-scc-manager/0000_20_kube-apiserver-operator_00_scc-restricted.yaml",
		}
	)
	if err := assets.ApplySCCs(ctx, sccs, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("failed to apply sccs %v", err)
		return err
	}
	if err := assets.ApplyClusterRoles(ctx, clusterRole, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRole %v: %v", clusterRole, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(ctx, clusterRoleBinding, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRolebinding %v: %v", clusterRoleBinding, err)
		return err
	}

	return nil
}
