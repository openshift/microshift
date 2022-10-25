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

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/version"
	"k8s.io/klog/v2"
)

type VersionManager struct {
	cfg *config.MicroshiftConfig
}

func NewVersionManager(cfg *config.MicroshiftConfig) *VersionManager {
	s := &VersionManager{}
	s.cfg = cfg
	return s
}

func (s *VersionManager) Name() string { return "version-manager" }
func (s *VersionManager) Dependencies() []string {
	return []string{"kube-apiserver"}
}

func (s *VersionManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {

	var cm = "version/microshift-version.yaml"

	defer close(stopped)
	defer close(ready)

	versionInfo := version.Get()
	var data = map[string]string{
		"major":   versionInfo.Major,
		"minor":   versionInfo.Minor,
		"version": versionInfo.String(),
	}

	kubeConfigPath := s.cfg.KubeConfigPath(config.KubeAdmin)
	if err := assets.ApplyConfigMapWithData(cm, data, kubeConfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v, %v", cm, err)
		return err
	}

	return ctx.Err()
}
