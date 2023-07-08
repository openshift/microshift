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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	migrationv1alpha1 "sigs.k8s.io/kube-storage-version-migrator/pkg/apis/migration/v1alpha1"
	migrationclient "sigs.k8s.io/kube-storage-version-migrator/pkg/clients/clientset"
	"sigs.k8s.io/kube-storage-version-migrator/pkg/controller"
)

var (
	backoff = wait.Backoff{
		Steps:    6,
		Duration: 10 * time.Millisecond,
		Factor:   5.0,
		Jitter:   0.1,
	}
)

const (
	// The migration trigger controller redo the discovery every discoveryPeriod.
	discoveryPeriod = 10 * time.Minute
)

type MigrationTrigger struct {
	client            migrationclient.Interface
	migrationInformer cache.SharedIndexInformer
	queue             workqueue.RateLimitingInterface
	// The timestamp of last time discovery is performed.
	heartbeat metav1.Time
}

func NewMigrationTrigger(c migrationclient.Interface) *MigrationTrigger {
	mt := &MigrationTrigger{
		client: c,
		// TODO: share one with the kubemigrator.go.
		migrationInformer: controller.NewStatusAndResourceIndexedInformer(c),
		queue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "migration_triggering_controller"),
	}
	mt.migrationInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    mt.addResource,
		UpdateFunc: mt.updateResource,
		DeleteFunc: mt.deleteResource,
	})

	return mt
}

func (mt *MigrationTrigger) dequeue() <-chan interface{} {
	work := make(chan interface{})
	go func() {
		for {
			item, quit := mt.queue.Get()
			if quit {
				return
			}
			work <- item
		}
	}()
	return work
}

// queueItem is the object in the workqueue.
type queueItem struct {
	// the namespace of the storageVersionMigration object.
	namespace string
	// the name of the storageVersionMigration object.
	name string
	// the resource the storageVersionMigration object is about.
	resource migrationv1alpha1.GroupVersionResource
}

func (mt *MigrationTrigger) addResource(obj interface{}) {
	m, ok := obj.(*migrationv1alpha1.StorageVersionMigration)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("expected StorageVersionMigration, got %#v", reflect.TypeOf(obj)))
		return
	}
	mt.enqueueResource(m)
}

func (mt *MigrationTrigger) deleteResource(obj interface{}) {
	m, ok := obj.(*migrationv1alpha1.StorageVersionMigration)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %+v", obj))
			return
		}
		m, ok = tombstone.Obj.(*migrationv1alpha1.StorageVersionMigration)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a StorageVersionMigration %#v", obj))
			return
		}
	}
	mt.enqueueResource(m)
}

func (mt *MigrationTrigger) updateResource(oldObj interface{}, obj interface{}) {
	mt.addResource(obj)
}

func (mt *MigrationTrigger) enqueueResource(migration *migrationv1alpha1.StorageVersionMigration) {
	it := &queueItem{
		namespace: migration.Namespace,
		name:      migration.Name,
		resource:  migration.Spec.Resource,
	}
	mt.queue.Add(it)
}

func (mt *MigrationTrigger) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()
	go mt.migrationInformer.Run(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), mt.migrationInformer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Unable to sync caches"))
		return
	}
	work := mt.dequeue()

	// We need to run the discovery routine and the migration management
	// routine in serial. Otherwise, they can corrupt
	// storageState.status.persistedStorageVersions.
	//
	// The discovery routine does the following for each resource:
	// a. checks if storageState.status.currentStorageVersion == discovered.storageVersion
	// b. if not, cleans existing migrations
	// c. launches a new migration
	// d. updates the storageState.status.currentStorageVersion and .persistedStorageVersions.
	//
	// The migration management routine does the following:
	// 1. gets a migration object from the workqueue
	// 2. gets the latest migration object from the apiserver
	// 3. if the latest migration object shows the migration has completed
	// successfully, updates the storageState.status.persistedStorageVersions
	// to only contain storageState.status.currentStorageVersion.
	//
	// The PersistedStorageVersions will be corrupted if the above steps
	// interleave in this order: 2, b, c, d, 3
	//
	// TODO: if we let the migration note down the currentStorageVersion,
	// we can avoid the race.
	ticker := time.NewTicker(discoveryPeriod)
	// Do a discovery once started.
	mt.processDiscovery(ctx)
	for {
		select {
		case <-ticker.C:
			mt.processDiscovery(ctx)
		case w := <-work:
			defer mt.queue.Done(w)
			err := mt.processQueue(ctx, w)
			if err == nil {
				mt.queue.Forget(w)
				break
			}
			utilruntime.HandleError(fmt.Errorf("failed to process %v: %v", w, err))
			mt.queue.AddRateLimited(w)
		case <-ctx.Done():
			return
		}
	}
}
