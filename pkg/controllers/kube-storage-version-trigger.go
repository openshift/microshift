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

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"

	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"
	migrationclient "sigs.k8s.io/kube-storage-version-migrator/pkg/clients/clientset"
	"sigs.k8s.io/kube-storage-version-migrator/pkg/trigger"
)

const (
	kubeStorageVersionTrigger = "storage-version-migration-trigger"
)

type KubeStorageVersionTrigger struct {
	kubeconfig string
	healthPath string
	healthPort string
}

func NewKubeStorageVersionTrigger(cfg *config.Config) *KubeStorageVersionTrigger {
	s := &KubeStorageVersionTrigger{}
	s.kubeconfig = cfg.KubeConfigAdminPath("")
	s.healthPort = "2113"
	s.healthPath = "storage-trigger-healthz"
	return s
}

func (s *KubeStorageVersionTrigger) Name() string { return kubeStorageVersionTrigger }
func (s *KubeStorageVersionTrigger) Dependencies() []string {
	return []string{kubeStorageVersionMigrator}
}

func (s *KubeStorageVersionTrigger) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
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

	go func() {
		errorChannel <- s.runTrigger(ctx)
	}()

	return <-errorChannel
}

func (s *KubeStorageVersionTrigger) runTrigger(ctx context.Context) error {
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

	config.UserAgent = kubeStorageVersionTrigger + "/v0.1"
	migration, err := migrationclient.NewForConfig(config)
	if err != nil {
		return err
	}

	// There seems to be a bug with the new discovery client
	// where the StorageVersionHash data is empty in the APIResources
	// TODO: investigate issue upstream, for now use legacy
	migration.DiscoveryClient.UseLegacyDiscovery = true

	c := trigger.NewMigrationTrigger(migration)
	c.Run(ctx)

	select {
	case <-ctx.Done():
		return nil
	default:
		return fmt.Errorf("%s shutdown unexpectedly", kubeStorageVersionTrigger)
	}
}
