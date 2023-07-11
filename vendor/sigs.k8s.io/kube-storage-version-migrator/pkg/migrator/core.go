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

package migrator

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kube-storage-version-migrator/pkg/migrator/metrics"
)

var metadataAccessor = meta.NewAccessor()

const (
	defaultChunkLimit  = 500
	defaultConcurrency = 1
)

type migrator struct {
	resource    schema.GroupVersionResource
	client      dynamic.Interface
	progress    progressInterface
	concurrency int
}

// NewMigrator creates a migrator that can migrate a single resource type.
func NewMigrator(resource schema.GroupVersionResource, client dynamic.Interface, progress progressInterface) *migrator {
	return &migrator{
		resource:    resource,
		client:      client,
		progress:    progress,
		concurrency: defaultConcurrency,
	}
}

func (m *migrator) get(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	// if namespace is empty, .Namespace(namespace) is ineffective.
	return m.client.
		Resource(m.resource).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
}

func (m *migrator) put(ctx context.Context, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	// if namespace is empty, .Namespace(namespace) is ineffective.
	return m.client.
		Resource(m.resource).
		Namespace(namespace).
		Update(ctx, obj, metav1.UpdateOptions{})
}

func (m *migrator) list(ctx context.Context, options metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return m.client.
		Resource(m.resource).
		Namespace(metav1.NamespaceAll).
		List(ctx, options)
}

// Run migrates all the instances of the resource type managed by the migrator.
func (m *migrator) Run(ctx context.Context) error {
	continueToken, err := m.progress.load(ctx)
	if err != nil {
		return err
	}
	for {
		list, listError := m.list(ctx,
			metav1.ListOptions{
				Limit:    defaultChunkLimit,
				Continue: continueToken,
			},
		)
		if errors.IsNotFound(listError) {
			// Fail this migration, we don't want to get stuck on a migration for a resource that does not exist.
			return fmt.Errorf("failed to list resources: %v", listError)
		}
		if listError != nil && !errors.IsResourceExpired(listError) {
			if canRetry(listError) {
				if seconds, delay := errors.SuggestsClientDelay(listError); delay {
					time.Sleep(time.Duration(seconds) * time.Second)
				}
				continue
			}
			return listError
		}
		if listError != nil && errors.IsResourceExpired(listError) {
			token, err := inconsistentContinueToken(listError)
			if err != nil {
				return err
			}
			continueToken = token
			err = m.progress.save(ctx, continueToken)
			if err != nil {
				utilruntime.HandleError(err)
			}
			continue
		}
		if err := m.migrateList(list); err != nil {
			return err
		}
		token, err := metadataAccessor.Continue(list)
		if err != nil {
			return err
		}
		metrics.Metrics.ObserveObjectsMigrated(len(list.Items), m.resource.String())
		// TODO: call ObserveObjectsRemaining as well, once https://github.com/kubernetes/kubernetes/pull/75993 is in.
		if len(token) == 0 {
			return nil
		}
		continueToken = token
		err = m.progress.save(ctx, continueToken)
		if err != nil {
			utilruntime.HandleError(err)
		}
	}
}

func (m *migrator) migrateList(l *unstructured.UnstructuredList) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workc := make(chan *unstructured.Unstructured)
	go func() {
		defer close(workc)
		for i := range l.Items {
			select {
			case workc <- &l.Items[i]:
			case <-ctx.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(m.concurrency)
	errc := make(chan error)
	for i := 0; i < m.concurrency; i++ {
		go func() {
			defer wg.Done()
			m.worker(ctx, workc, errc)
		}()
	}

	go func() {
		wg.Wait()
		close(errc)
	}()

	var errors []error
	for err := range errc {
		errors = append(errors, err)
	}
	return utilerrors.NewAggregate(errors)
}

func (m *migrator) worker(ctx context.Context, workc <-chan *unstructured.Unstructured, errc chan<- error) {
	for item := range workc {
		err := m.migrateOneItem(ctx, item)
		if err != nil {
			select {
			case errc <- err:
				continue
			case <-ctx.Done():
				return
			}
		}
	}
}

func (m *migrator) migrateOneItem(ctx context.Context, item *unstructured.Unstructured) error {
	namespace, err := metadataAccessor.Namespace(item)
	if err != nil {
		return err
	}
	name, err := metadataAccessor.Name(item)
	if err != nil {
		return err
	}
	getBeforePut := false
	for {
		getBeforePut, err = m.try(ctx, namespace, name, item, getBeforePut)
		if err == nil || errors.IsNotFound(err) {
			return nil
		}
		if canRetry(err) {
			seconds, delay := errors.SuggestsClientDelay(err)
			switch {
			case delay && len(namespace) > 0:
				klog.Warningf("migration of %s, in the %s namespace, will be retried after a %ds delay: %v", name, namespace, seconds, err)
				time.Sleep(time.Duration(seconds) * time.Second)
			case delay:
				klog.Warningf("migration of %s will be retried after a %ds delay: %v", name, seconds, err)
				time.Sleep(time.Duration(seconds) * time.Second)
			case !delay && len(namespace) > 0:
				klog.Warningf("migration of %s, in the %s namespace, will be retried: %v", name, namespace, err)
			default:
				klog.Warningf("migration of %s will be retried: %v", name, err)
			}
			continue
		}
		// error is not retriable
		return err
	}
}

// try tries to migrate the single object by PUT. It refreshes the object via
// GET if "get" is true. If the PUT fails due to conflicts, or the GET fails,
// the function requests the next try to GET the new object.
func (m *migrator) try(ctx context.Context, namespace, name string, item *unstructured.Unstructured, get bool) (bool, error) {
	var err error
	if get {
		item, err = m.get(ctx, namespace, name)
		if err != nil {
			return true, err
		}
	}
	_, err = m.put(ctx, namespace, item)
	if err == nil {
		return false, nil
	}
	return errors.IsConflict(err), err

	// TODO: The oc admin uses a defer function to do bandwidth limiting
	// after doing all operations. The rate limiter is marked as an alpha
	// feature.  Is it better than the built-in qps limit in the REST
	// client? Maybe it's necessary because not all resource types are of
	// the same size?
}

// TODO: move this helper to "k8s.io/apimachinery/pkg/api/errors"
func inconsistentContinueToken(err error) (string, error) {
	status, ok := err.(errors.APIStatus)
	if !ok {
		return "", fmt.Errorf("expected error to implement the APIStatus interface, got %v", reflect.TypeOf(err))
	}
	token := status.Status().ListMeta.Continue
	if len(token) == 0 {
		return "", fmt.Errorf("expected non empty continue token")
	}
	return token, nil
}
