/*
Copyright 2018 The Kubernetes Authors.

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

package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	migrationv1alpha1 "sigs.k8s.io/kube-storage-version-migrator/pkg/apis/migration/v1alpha1"
	migrationclient "sigs.k8s.io/kube-storage-version-migrator/pkg/clients/clientset"
	"sigs.k8s.io/kube-storage-version-migrator/pkg/migrator"
	"sigs.k8s.io/kube-storage-version-migrator/pkg/migrator/metrics"
)

// KubeMigrator monitors storageVersionMigraiton objects, fulfills the
// migration, and updates the status of the storageVersionMigration objects.
type KubeMigrator struct {
	dynamic           dynamic.Interface
	migrationClient   migrationclient.Interface
	migrationInformer cache.SharedIndexInformer
}

// NewKubeMigrator creates KubeMigrator.
func NewKubeMigrator(dynamic dynamic.Interface, migrationClient migrationclient.Interface) *KubeMigrator {
	informer := NewStatusIndexedInformer(migrationClient)
	return &KubeMigrator{
		dynamic:           dynamic,
		migrationClient:   migrationClient,
		migrationInformer: informer,
	}
}

func (km *KubeMigrator) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()
	go km.migrationInformer.Run(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), km.migrationInformer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Unable to sync caches"))
		return
	}
	wait.UntilWithContext(ctx, km.process, time.Second)
}

func (km *KubeMigrator) process(ctx context.Context) {
	// KubeMigrator has only one worker, so it doesn't need to use a
	// workqueue to ensure only there is a single thread processing a
	// storageVersionMigration.

	// The already "Running" storageVersionMigrations are the priority.
	runnings, err := km.migrationInformer.GetIndexer().ByIndex(StatusIndex, StatusRunning)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	if len(runnings) != 0 {
		utilruntime.HandleError(km.processOne(ctx, runnings[0]))
		return
	}

	// The next priority is the pending storageVersionMigrations.
	pendings, err := km.migrationInformer.GetIndexer().ByIndex(StatusIndex, StatusPending)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	if len(pendings) != 0 {
		utilruntime.HandleError(km.processOne(ctx, pendings[0]))
		return
	}
}

func (km *KubeMigrator) processOne(ctx context.Context, obj interface{}) error {
	m, ok := obj.(*migrationv1alpha1.StorageVersionMigration)
	if !ok {
		return fmt.Errorf("expected StorageVersionMigration, got %#v", reflect.TypeOf(obj))
	}
	// get the fresh object from the apiserver to make sure the object
	// still exists, and the object is not completed.
	m, err := km.migrationClient.MigrationV1alpha1().StorageVersionMigrations().Get(ctx, m.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if HasCondition(m, migrationv1alpha1.MigrationSucceeded) || HasCondition(m, migrationv1alpha1.MigrationFailed) {
		klog.V(2).Infof("%v: migration has already completed", m.Name)
		return nil
	}
	m, err = km.updateStatus(ctx, m, migrationv1alpha1.MigrationRunning, "")
	klog.V(2).Infof("%v: migration running", m.Name)
	if err != nil {
		return err
	}
	progressTracker := migrator.NewProgressTracker(km.migrationClient.MigrationV1alpha1().StorageVersionMigrations(), m.Name)
	core := migrator.NewMigrator(resource(m), km.dynamic, progressTracker)
	// If the storageVersionMigration object is deleted during Run(), Run()
	// will return an error when it tries to write the continueToken into the
	// migration object. Thus, it's not necessary to register a deletion
	// event handler with the migrationInformer to interrupt the Run().
	err = core.Run(ctx)
	utilruntime.HandleError(err)
	if err == nil {
		if _, err := km.updateStatus(ctx, m, migrationv1alpha1.MigrationSucceeded, ""); err != nil {
			utilruntime.HandleError(err)
		}
		metrics.Metrics.ObserveSucceededMigration(resource(m).String())
		klog.V(2).Infof("%v: migration succeeded", m.Name)
		return err
	}
	klog.Errorf("%v: migration failed: %v", m.Name, err)
	if _, err := km.updateStatus(ctx, m, migrationv1alpha1.MigrationFailed, err.Error()); err != nil {
		utilruntime.HandleError(err)
	}
	metrics.Metrics.ObserveFailedMigration(resource(m).String())
	return err
}

// updateStatus always retries no matter what kind of error is returned by the
// apiserver, because it's a pity to start over the entire migration merely
// because a status update failure.
// updateStatus also removes other KNOWN conditions.
func (km *KubeMigrator) updateStatus(ctx context.Context, m *migrationv1alpha1.StorageVersionMigration, condition migrationv1alpha1.MigrationConditionType, message string) (*migrationv1alpha1.StorageVersionMigration, error) {
	backoff := wait.Backoff{
		Steps:    6,
		Duration: 10 * time.Millisecond,
		Factor:   5.0,
		Jitter:   0.1,
	}
	return m, wait.ExponentialBackoff(backoff, func() (bool, error) {
		var newConditions []migrationv1alpha1.MigrationCondition
		for _, c := range m.Status.Conditions {
			switch c.Type {
			case migrationv1alpha1.MigrationRunning:
			case migrationv1alpha1.MigrationSucceeded:
			case migrationv1alpha1.MigrationFailed:
			default:
				// keeps unknown conditions
				newConditions = append(newConditions, c)
			}
		}
		newCondition := migrationv1alpha1.MigrationCondition{
			Type:           condition,
			Status:         corev1.ConditionTrue,
			LastUpdateTime: metav1.Now(),
			Message:        message,
		}
		newConditions = append(newConditions, newCondition)
		m.Status.Conditions = newConditions

		_, err := km.migrationClient.MigrationV1alpha1().StorageVersionMigrations().UpdateStatus(ctx, m, metav1.UpdateOptions{})
		if err == nil {
			return true, nil
		}
		// Always refresh and retry, no matter what kind of error is returned by the apiserver.
		updated, err := km.migrationClient.MigrationV1alpha1().StorageVersionMigrations().Get(ctx, m.Name, metav1.GetOptions{})
		if err == nil {
			m = updated
		}
		return false, nil
	})
}
