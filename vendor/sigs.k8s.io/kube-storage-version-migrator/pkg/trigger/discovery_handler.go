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
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/klog/v2"

	migrationv1alpha1 "sigs.k8s.io/kube-storage-version-migrator/pkg/apis/migration/v1alpha1"
	"sigs.k8s.io/kube-storage-version-migrator/pkg/controller"
)

func (mt *MigrationTrigger) processDiscovery(ctx context.Context) {
	var resources []*metav1.APIResourceList
	var err2 error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		resources, err2 = mt.client.Discovery().ServerPreferredResources()
		if err2 != nil {
			utilruntime.HandleError(err2)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		if discovery.IsGroupDiscoveryFailedError(err2) {
			// process the partial discovery result, and update the heartbeat for
			// resources that do have a valid discovery document
			klog.Warningf("failed to discover some groups: %v; processing partial result", err2.(*discovery.ErrGroupDiscoveryFailed).Groups)
		} else {
			klog.Warningf("failed to discover preferred resources: %v", err2)
		}
	}
	mt.heartbeat = metav1.Now()
	for _, l := range resources {
		gv, err := schema.ParseGroupVersion(l.GroupVersion)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("unexpected group version error: %v", err))
			continue
		}
		for _, r := range l.APIResources {
			if r.Group == "" {
				r.Group = gv.Group
			}
			if r.Version == "" {
				r.Version = gv.Version
			}
			mt.processDiscoveryResource(ctx, r)
		}
	}
}

func toGroupResource(r metav1.APIResource) migrationv1alpha1.GroupVersionResource {
	return migrationv1alpha1.GroupVersionResource{
		Group:    r.Group,
		Version:  r.Version,
		Resource: r.Name,
	}
}

// cleanMigrations removes all storageVersionMigrations whose .spec.resource == r.
func (mt *MigrationTrigger) cleanMigrations(ctx context.Context, r metav1.APIResource) error {
	// Using the cache to find all matching migrations.
	// The delay of the cache shouldn't matter in practice, because
	// existing migrations are created by previous discovery cycles, they
	// have at least discoveryPeriod to enter the informer's cache.
	idx := mt.migrationInformer.GetIndexer()
	l, err := idx.ByIndex(controller.ResourceIndex, controller.ToIndex(toGroupResource(r)))
	if err != nil {
		return err
	}
	for _, m := range l {
		mm, ok := m.(*migrationv1alpha1.StorageVersionMigration)
		if !ok {
			return fmt.Errorf("expected StorageVersionMigration, got %#v", reflect.TypeOf(m))
		}
		err := mt.client.MigrationV1alpha1().StorageVersionMigrations().Delete(ctx, mm.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("unexpected error deleting migration %s, %v", mm.Name, err)
		}
	}
	return nil
}

func (mt *MigrationTrigger) launchMigration(ctx context.Context, resource migrationv1alpha1.GroupVersionResource) error {
	m := &migrationv1alpha1.StorageVersionMigration{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: storageStateName(resource) + "-",
		},
		Spec: migrationv1alpha1.StorageVersionMigrationSpec{
			Resource: resource,
		},
	}
	_, err := mt.client.MigrationV1alpha1().StorageVersionMigrations().Create(ctx, m, metav1.CreateOptions{})
	return err
}

// relaunchMigration cleans existing migrations for the resource, and launch a new one.
func (mt *MigrationTrigger) relaunchMigration(ctx context.Context, r metav1.APIResource) error {
	if err := mt.cleanMigrations(ctx, r); err != nil {
		return err
	}
	return mt.launchMigration(ctx, toGroupResource(r))

}

func (mt *MigrationTrigger) newStorageState(r metav1.APIResource) *migrationv1alpha1.StorageState {
	return &migrationv1alpha1.StorageState{
		ObjectMeta: metav1.ObjectMeta{
			Name: storageStateName(toGroupResource(r)),
		},
		Spec: migrationv1alpha1.StorageStateSpec{
			Resource: migrationv1alpha1.GroupResource{
				Group:    r.Group,
				Resource: r.Name,
			},
		},
	}
}

func (mt *MigrationTrigger) updateStorageState(ctx context.Context, currentHash string, r metav1.APIResource) error {
	// We will retry on any error, because failing to update the
	// heartbeat of the storageState can lead to redo migration, which is
	// costly.
	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		ss, err := mt.client.MigrationV1alpha1().StorageStates().Get(ctx, storageStateName(toGroupResource(r)), metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			utilruntime.HandleError(err)
			return false, nil
		}
		if err != nil && errors.IsNotFound(err) {
			// Note that the apiserver resets the status field for
			// the POST request. We need to update via the status
			// endpoint.
			ss, err = mt.client.MigrationV1alpha1().StorageStates().Create(ctx, mt.newStorageState(r), metav1.CreateOptions{})
			if err != nil {
				utilruntime.HandleError(err)
				return false, nil
			}
		}
		if ss.Status.CurrentStorageVersionHash != currentHash {
			ss.Status.CurrentStorageVersionHash = currentHash
			if len(ss.Status.PersistedStorageVersionHashes) == 0 {
				ss.Status.PersistedStorageVersionHashes = []string{migrationv1alpha1.Unknown}
			} else {
				ss.Status.PersistedStorageVersionHashes = append(ss.Status.PersistedStorageVersionHashes, currentHash)
			}
		}
		ss.Status.LastHeartbeatTime = mt.heartbeat
		_, err = mt.client.MigrationV1alpha1().StorageStates().UpdateStatus(ctx, ss, metav1.UpdateOptions{})
		if err != nil {
			utilruntime.HandleError(err)
			return false, nil
		}
		return true, nil
	})
}

func (mt *MigrationTrigger) staleStorageState(ss *migrationv1alpha1.StorageState) bool {
	return ss.Status.LastHeartbeatTime.Add(2 * discoveryPeriod).Before(mt.heartbeat.Time)
}

func (mt *MigrationTrigger) processDiscoveryResource(ctx context.Context, r metav1.APIResource) {
	klog.V(4).Infof("processing %#v", r)
	if r.StorageVersionHash == "" {
		klog.V(2).Infof("ignored resource %s/%s because its storageVersionHash is empty", r.Group, r.Name)
		return
	}
	ss, getErr := mt.client.MigrationV1alpha1().StorageStates().Get(ctx, storageStateName(toGroupResource(r)), metav1.GetOptions{})
	if getErr != nil && !errors.IsNotFound(getErr) {
		utilruntime.HandleError(getErr)
		return
	}
	found := getErr == nil
	stale := found && mt.staleStorageState(ss)
	storageVersionChanged := found && ss.Status.CurrentStorageVersionHash != r.StorageVersionHash
	needsMigration := found && !mt.isMigrated(ss) && !mt.hasPendingOrRunningMigration(r)
	relaunchMigration := stale || !found || storageVersionChanged || needsMigration

	if stale {
		if err := mt.client.MigrationV1alpha1().StorageStates().Delete(ctx, storageStateName(toGroupResource(r)), metav1.DeleteOptions{}); err != nil {
			utilruntime.HandleError(err)
			return
		}
	}

	if relaunchMigration {
		// Note that this means historical migration objects are deleted.
		if err := mt.relaunchMigration(ctx, r); err != nil {
			utilruntime.HandleError(err)
		}
	}

	// always update status.heartbeat, sometimes update the version hashes.
	mt.updateStorageState(ctx, r.StorageVersionHash, r)
}
func (mt *MigrationTrigger) isMigrated(ss *migrationv1alpha1.StorageState) bool {
	if len(ss.Status.PersistedStorageVersionHashes) != 1 {
		return false
	}
	return ss.Status.CurrentStorageVersionHash == ss.Status.PersistedStorageVersionHashes[0]
}

func (mt *MigrationTrigger) hasPendingOrRunningMigration(r metav1.APIResource) bool {
	// get the corresponding StorageVersionMigration resource
	migrations, err := mt.migrationInformer.GetIndexer().ByIndex(controller.ResourceIndex, controller.ToIndex(toGroupResource(r)))
	if err != nil {
		utilruntime.HandleError(err)
		return false
	}
	for _, migration := range migrations {
		m := migration.(*migrationv1alpha1.StorageVersionMigration)
		if controller.HasCondition(m, migrationv1alpha1.MigrationSucceeded) || controller.HasCondition(m, migrationv1alpha1.MigrationFailed) {
			continue
		}
		// migration is running or pending
		return true
	}
	return false
}
