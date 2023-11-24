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

type OpenShiftCRDManager struct {
	cfg *config.Config
}

func NewOpenShiftCRDManager(cfg *config.Config) *OpenShiftCRDManager {
	s := &OpenShiftCRDManager{}
	s.cfg = cfg
	return s
}

func (s *OpenShiftCRDManager) Name() string           { return "openshift-crd-manager" }
func (s *OpenShiftCRDManager) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *OpenShiftCRDManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	if err := assets.ApplyCRDs(ctx, s.cfg); err != nil {
		klog.Errorf("%s unable to apply default CRDs: %v", s.Name(), err)
		return err
	}
	klog.Infof("%s applied default CRDs", s.Name())

	klog.Infof("%s waiting for CRDs acceptance before proceeding", s.Name())
	if err := assets.WaitForCrdsEstablished(ctx, s.cfg); err != nil {
		klog.Errorf("%s unable to confirm all CRDs are ready: %v", s.Name(), err)
		return ctx.Err()
	}
	klog.Infof("%s all CRDs are ready", s.Name())
	close(ready)

	return ctx.Err()
}
