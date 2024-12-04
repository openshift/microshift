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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"

	klog "k8s.io/klog/v2"
	kubescheduler "k8s.io/kubernetes/cmd/kube-scheduler/app"
	schedulerOptions "k8s.io/kubernetes/cmd/kube-scheduler/app/options"
)

type KubeScheduler struct {
	options    *schedulerOptions.Options
	kubeconfig string
}

func NewKubeScheduler(cfg *config.Config) *KubeScheduler {
	s := &KubeScheduler{}
	s.configure(cfg)
	return s
}

func (s *KubeScheduler) Name() string           { return "kube-scheduler" }
func (s *KubeScheduler) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *KubeScheduler) configure(cfg *config.Config) {
	if err := s.writeConfig(cfg); err != nil {
		klog.Fatalf("failed to write kube-scheduler config: %v", err)
	}

	s.options = schedulerOptions.NewOptions()
	s.options.ConfigFile = filepath.Join(config.DataDir, "/resources/kube-scheduler/config/config.yaml")
	s.options.Authentication.RemoteKubeConfigFile = cfg.KubeConfigPath(config.KubeScheduler)
	s.options.Authorization.RemoteKubeConfigFile = cfg.KubeConfigPath(config.KubeScheduler)
	s.options.SecureServing.MinTLSVersion = cfg.ApiServer.TLS.MinVersion
	s.options.SecureServing.CipherSuites = cfg.ApiServer.TLS.CipherSuites
	s.kubeconfig = cfg.KubeConfigPath(config.KubeScheduler)
}

func (s *KubeScheduler) writeConfig(cfg *config.Config) error {
	data := []byte(`apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration
clientConnection:
  kubeconfig: ` + cfg.KubeConfigPath(config.KubeScheduler) + `
leaderElection:
  leaderElect: false`)

	path := filepath.Join(config.DataDir, "resources", "kube-scheduler", "config", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), os.FileMode(0700)); err != nil {
		return fmt.Errorf("creating directory path %s: %w", path, err)
	}
	return os.WriteFile(path, data, 0400)
}

func (s *KubeScheduler) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	errorChannel := make(chan error, 1)

	// run readiness check
	go func() {
		// This endpoint uses a self-signed certificate on purpose, we need to skip verification.
		healthcheckStatus := util.RetryInsecureGet(ctx, "https://localhost:10259/healthz")
		if healthcheckStatus != 200 {
			klog.Errorf("%s healthcheck failed due to kube-scheduler failure to start", s.Name())
			errorChannel <- errors.New("kube-scheduler healthcheck failed")
		}

		klog.Infof("%s is ready", s.Name())
		close(ready)
	}()

	cc, sched, err := kubescheduler.Setup(ctx, s.options)
	if err != nil {
		return err
	}

	panicChannel := make(chan any, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicChannel <- r
			}
		}()
		errorChannel <- kubescheduler.Run(ctx, cc, sched)
	}()

	select {
	case err := <-errorChannel:
		return err
	case perr := <-panicChannel:
		panic(perr)
	}
}
