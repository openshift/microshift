/*
Copyright 2019 The Kubernetes Authors.

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

package trigger

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	migrationv1alpha1 "sigs.k8s.io/kube-storage-version-migrator/pkg/apis/migration/v1alpha1"
	"sigs.k8s.io/kube-storage-version-migrator/pkg/controller"
)

func storageStateName(resource migrationv1alpha1.GroupVersionResource) string {
	// TODO: add this rule to the CRD validation
	// TODO: we might use ResourceID as the name in the future.
	if resource.Group == "" {
		return resource.Resource
	}
	return resource.Resource + "." + resource.Group
}

func (mt *MigrationTrigger) markStorageStateSucceeded(ctx context.Context, resource migrationv1alpha1.GroupVersionResource) error {
	// We will retry on any error. Migrating a resource takes a long time.
	// It would be a pity to give up just because of an update error.
	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		ss, err := mt.client.MigrationV1alpha1().StorageStates().Get(ctx, storageStateName(resource), metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			utilruntime.HandleError(err)
			return false, nil
		}
		if err != nil && errors.IsNotFound(err) {
			utilruntime.HandleError(err)
			// if the storage state does not exist, there is not
			// much we can do. We will leave it to the next
			// discovery routine to create the storage state.
			return true, nil
		}
		ss.Status.PersistedStorageVersionHashes = []string{ss.Status.CurrentStorageVersionHash}
		_, err = mt.client.MigrationV1alpha1().StorageStates().UpdateStatus(ctx, ss, metav1.UpdateOptions{})
		if err != nil {
			utilruntime.HandleError(err)
			return false, nil
		}
		return true, nil
	})
}

func (mt *MigrationTrigger) processMigration(ctx context.Context, m *migrationv1alpha1.StorageVersionMigration) error {
	klog.V(2).Infof("processing migration %#v", m)
	switch {
	case controller.HasCondition(m, migrationv1alpha1.MigrationSucceeded):
		return mt.markStorageStateSucceeded(ctx, m.Spec.Resource)
	case controller.HasCondition(m, migrationv1alpha1.MigrationFailed):
		// The migration controller should have already tried its best
		// to complete the migration before marking the migration as
		// failed. There is nothing the triggering controller can do.
		return nil
	default:
		return nil
	}
}

func (mt *MigrationTrigger) processQueue(ctx context.Context, obj interface{}) error {
	item, ok := obj.(*queueItem)
	if !ok {
		return fmt.Errorf("expected queueItem, got %#v", reflect.TypeOf(obj))
	}
	// historic migrations are cleaned up when the controller observes
	// storage version changes in the discovery doc.
	m, err := mt.client.MigrationV1alpha1().StorageVersionMigrations().Get(ctx, item.name, metav1.GetOptions{})
	if err == nil {
		return mt.processMigration(ctx, m)
	}

	if err != nil && errors.IsNotFound(err) {
		// Likely the migration is deleted because mt is cleaning up
		// migration history, in which case there is nothing to be
		// done. If the migration is mistakenly removed by other
		// clients, the periodic discovery routine will restart the
		// migration.
		return nil
	}
	return err
}
