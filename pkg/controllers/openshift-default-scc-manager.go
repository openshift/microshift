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

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"k8s.io/klog/v2"
)

type OpenShiftDefaultSCCManager struct {
	cfg *config.MicroshiftConfig
}

func NewOpenShiftDefaultSCCManager(cfg *config.MicroshiftConfig) *OpenShiftDefaultSCCManager {
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
	defer close(ready)
	// TO-DO add readiness check
	if err := ApplyDefaultSCCs(s.cfg); err != nil {
		klog.Errorf("%s unable to apply default SCCs: %v", s.Name(), err)
		return err
	}
	klog.Infof("%s applied default SCCs", s.Name())
	return ctx.Err()
}

func ApplyDefaultSCCs(cfg *config.MicroshiftConfig) error {
	kubeconfigPath := cfg.DataDir + "/resources/kubeadmin/kubeconfig"
	var (
		sccs = []string{
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml",
		}
	)
	if err := assets.ApplySCCs(sccs, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("failed to apply sccs %v", err)
		return err
	}
	return nil
}
