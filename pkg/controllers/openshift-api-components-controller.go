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

	"github.com/openshift/microshift/pkg/config"

	"github.com/sirupsen/logrus"
)

type OpenShiftAPIComponentsControllerManager struct {
	cfg *config.MicroshiftConfig
}

func NewOpenShiftAPIComponents(cfg *config.MicroshiftConfig) *OpenShiftAPIComponentsControllerManager {
	s := &OpenShiftAPIComponentsControllerManager{}
	s.cfg = cfg
	return s
}

func (s *OpenShiftAPIComponentsControllerManager) Name() string {
	return "openshift-api-components-manager"
}
func (s *OpenShiftAPIComponentsControllerManager) Dependencies() []string {
	return []string{"kube-apiserver", "openshift-prepjob-manager", "oauth-apiserver"}
}

func (s *OpenShiftAPIComponentsControllerManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(ready)
	// TO-DO add readiness check
	if err := StartOCPAPIComponents(s.cfg); err != nil {
		logrus.Errorf("%s unable to prepare ocp componets: %v", s.Name(), err)
	}
	logrus.Infof("%s launched ocp componets", s.Name())
	return ctx.Err()
}
