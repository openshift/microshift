/*
Copyright © 2023 MicroShift Contributors

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
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"
	migrationclient "sigs.k8s.io/kube-storage-version-migrator/pkg/clients/clientset"
	"sigs.k8s.io/kube-storage-version-migrator/pkg/controller"
)

const (
	kubeStorageVersionMigrator = "storage-version-migration-migrator"
	kubeAPIQPS                 = float32(40.0)
	kubeAPIBurst               = 1000
)

type KubeStorageVersionMigrator struct {
	kubeconfig string
	healthPath string
	healthPort string
}

func NewKubeStorageVersionMigrator(cfg *config.Config) *KubeStorageVersionMigrator {
	s := &KubeStorageVersionMigrator{}
	s.kubeconfig = filepath.Join(cfg.KubeConfigRootAdminPath(), "kubeconfig")
	s.healthPort = "2112"
	s.healthPath = "storage-migrator-healthz"
	return s
}

func (s *KubeStorageVersionMigrator) Name() string { return kubeStorageVersionMigrator }
func (s *KubeStorageVersionMigrator) Dependencies() []string {
	return []string{"kube-apiserver", "kubelet"}
}

func (s *KubeStorageVersionMigrator) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	errorChannel := make(chan error, 1)
	defer close(stopped)

	// run readiness check
	go func() {
		healthUrl := fmt.Sprintf("http://localhost:%s/%s", s.healthPort, s.healthPath)
		healthcheckStatus := util.RetryInsecureGet(ctx, healthUrl)
		if healthcheckStatus != 200 {
			klog.ErrorS(fmt.Errorf("%s failed to start", s.Name()), "healthcheck failed", "name", s.Name())
			errorChannel <- fmt.Errorf("%s healthcheck failed", s.Name())
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
		errorChannel <- s.runMigrator(ctx)
	}()

	select {
	case err := <-errorChannel:
		return err
	case perr := <-panicChannel:
		panic(perr)
	}
}

func (s *KubeStorageVersionMigrator) runMigrator(ctx context.Context) error {
	start, shutdown := util.HealthCheckServer(ctx, s.healthPath, s.healthPort)
	go func() {
		err := start()
		if err != nil {
			klog.ErrorS(err, "could not create health endpoint", "controller", s.Name())
		}
	}()
	defer func() {
		err := shutdown()
		if err != nil {
			klog.ErrorS(err, "liveness server did not shutdown gracefully", "controller", s.Name())
		}
	}()

	config, err := clientcmd.BuildConfigFromFlags("", s.kubeconfig)
	if err != nil {
		return fmt.Errorf("error initializing client config: %v for kubeconfig: %v", err.Error(), s.kubeconfig)
	}
	config.QPS = kubeAPIQPS
	config.Burst = kubeAPIBurst
	dynamic, err := dynamic.NewForConfig(rest.AddUserAgent(config, kubeStorageVersionMigrator))
	if err != nil {
		return err
	}
	migration, err := migrationclient.NewForConfig(config)
	if err != nil {
		return err
	}
	c := controller.NewKubeMigrator(
		dynamic,
		migration,
	)

	c.Run(ctx)

	select {
	case <-ctx.Done():
		return nil
	default:
		return fmt.Errorf("%s shutdown unexpectedly", kubeStorageVersionMigrator)
	}
}
