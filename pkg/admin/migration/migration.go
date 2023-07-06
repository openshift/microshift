package migration

import (
	"context"
	"fmt"
	"strings"
	"time"

	crdclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	apiserviceclient "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/typed/apiregistration/v1"

	"k8s.io/klog/v2"
)

// Skipping resources which might cycle quickly or cause a lot of overhead to migrate
var excludeResources = sets.NewString(
	"events",
)

var metadataAccessor = meta.NewAccessor()

type migrator struct {
	client           dynamic.Interface
	discoveryClient  discovery.ServerResourcesInterface
	crdClient        v1.CustomResourceDefinitionInterface
	apiserviceClient apiregistrationv1.APIServiceInterface
}

func NewMigrator(kubeConfigPath string) (*migrator, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build rest config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes clientset config: %w", err)
	}
	crd, err := crdclient.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build crd clientset config: %w", err)
	}
	apiservice, err := apiserviceclient.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build apiservice client config: %w", err)
	}
	dynamic, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build dynamic client config: %w", err)
	}

	return &migrator{
		client:           dynamic,
		discoveryClient:  clientset.Discovery(),
		crdClient:        crd.ApiextensionsV1().CustomResourceDefinitions(),
		apiserviceClient: apiservice.ApiregistrationV1().APIServices(),
	}, nil
}

func (d *migrator) Start(ctx context.Context) (*MigrationResultList, error) {
	return d.start(ctx)
}

func (d *migrator) start(ctx context.Context) (*MigrationResultList, error) {
	schemas, err := d.findMigratableResources(ctx)
	if err != nil {
		return nil, err
	}
	results := &MigrationResultList{
		Status: MigrationSuccess,
	}
	errorOccured := false
	start := time.Now()
	klog.Info("schema migration started")

	// Currently we are sequentially migrating items, we will need to revisit this if performance becomes a problem
	for _, sch := range schemas {
		// A list of objects might be very large, they will be chunked results with a continue token
		// here we loop for as many times we have a continue token or an error occurred
		continueToken := ""
		for {
			objectList := &unstructured.UnstructuredList{}
			migrationErr := retry.OnError(retry.DefaultBackoff, canRetry, func() error {
				var err error
				objectList, err = d.client.Resource(sch).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
					Continue: continueToken,
				})
				if err != nil {
					return err
				}
				return nil
			})

			// If resource expired error, retry
			if migrationErr != nil && errors.IsResourceExpired(migrationErr) {
				token, err := inconsistentContinueToken(migrationErr)
				if err != nil {
					err = fmt.Errorf("failed to get continue token: %w", err)
					results.Items = append(results.Items, MigrationResult{
						Error:                err,
						GroupVersionResource: sch,
						Timestamp:            time.Now()})
					break
				}
				continueToken = token
				continue
			}

			if migrationErr != nil {
				errorOccured = true
				migrationErr = fmt.Errorf("could not list resources: %w", migrationErr)
				results.Items = append(results.Items, MigrationResult{
					Error:                migrationErr,
					GroupVersionResource: sch,
					Timestamp:            time.Now()})
				break
			}

			status := MigrationSuccess
			for _, object := range objectList.Items {
				ref := object
				migrationErr := d.migrateOneItem(ctx, sch, &ref)
				if migrationErr != nil {
					errorOccured = true
					status = MigrationFailure
				}
			}

			results.Items = append(results.Items, MigrationResult{
				Error:                migrationErr,
				Status:               status,
				GroupVersionResource: sch,
				Timestamp:            time.Now()})

			// Check if the list contains a continue token
			token, err := metadataAccessor.Continue(objectList)
			if err != nil {
				err = fmt.Errorf("failed to get continue token: %w", err)
				results.Items = append(results.Items, MigrationResult{
					Error:                err,
					GroupVersionResource: sch,
					Timestamp:            time.Now()})
				break
			}
			if len(token) == 0 {
				break
			}
			continueToken = token
		}
	}
	if errorOccured {
		results.Status = MigrationFailure
	}
	klog.InfoS("schema migration finished", "duration", time.Since(start).String())
	return results, nil
}

// findMigratableResources finds all the resources that potentially need
// migration. Although all migratable resources are accessible via multiple
// versions, the returned list only include one version.
//
// It builds the list in these steps:
// 1. build a map from resource name to the groupVersions, excluding subresources, custom resources, or aggregated resources.
// 2. exclude all the resource that is only available from one groupVersions.
// 3. exclude the resource that does not support "list" and "update" (thus not migratable).
//
// More information can be found here:
// https://github.com/kubernetes-sigs/kube-storage-version-migrator/blob/acdee30ced218b79e39c6a701985e8cd8bd33824/pkg/initializer/discover.go#L55-L125
func (d *migrator) findMigratableResources(ctx context.Context) ([]schema.GroupVersionResource, error) {
	aggregatedGroups, err := d.findAggregatedGroups(ctx)
	if err != nil {
		return nil, err
	}
	ret := []schema.GroupVersionResource{}
	resourceLists, err := d.discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			klog.ErrorS(err, "cannot parse group version, ignored", "version", resourceList.GroupVersion)
			continue
		}

		if aggregatedGroups.Has(gv.Group) {
			klog.InfoS("ignored because it's an aggregated group", "group", gv.Group)
			continue
		}
		for _, r := range resourceList.APIResources {
			// ignore subresources
			if strings.Contains(r.Name, "/") {
				klog.InfoS("ignored subresource", "group", gv.Group, "name", r.Name, "version", gv.Version)
				continue
			}
			// ignore excluded resources
			if excludeResources.Has(r.Name) {
				klog.InfoS("ignored excluded resource", "group", gv.Group, "name", r.Name, "version", gv.Version)
				continue
			}
			// ignore resources that cannot be listed and updated
			if !sets.NewString(r.Verbs...).HasAll("list", "update") {
				klog.InfoS("ignored because verb does not contain list or update", "group", gv.Group, "name", r.Name, "version", gv.Version)
				continue
			}
			ret = append(ret, gv.WithResource(r.Name))
		}
	}

	return ret, nil
}

func (m *migrator) migrateOneItem(ctx context.Context, resource schema.GroupVersionResource, item *unstructured.Unstructured) error {
	namespace, err := metadataAccessor.Namespace(item)
	if err != nil {
		return err
	}
	name, err := metadataAccessor.Name(item)
	if err != nil {
		return err
	}

	for {
		err = m.try(ctx, resource, namespace, item)
		if err == nil || errors.IsNotFound(err) {
			klog.InfoS("successfully migrated object", "name", name, "namespace", namespace, "resource", resource.String())
			return nil
		}
		if canRetry(err) {
			seconds, delay := errors.SuggestsClientDelay(err)
			klog.ErrorS(err, "migration of an object will be retried", "name", name, "namespace", namespace, "delay", seconds)
			if delay {
				time.Sleep(time.Duration(seconds) * time.Second)
			}
			continue
		}
		// error is not retriable
		return fmt.Errorf("can not retry: %+v", err)
	}
}

func (m *migrator) try(ctx context.Context, resource schema.GroupVersionResource, namespace string, item *unstructured.Unstructured) error {
	_, err := m.client.
		Resource(resource).
		Namespace(namespace).
		Update(ctx, item, metav1.UpdateOptions{})
	return err
}

func (d *migrator) findAggregatedGroups(ctx context.Context) (sets.Set[string], error) {
	ret := sets.New[string]()
	l, err := d.apiserviceClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return ret, err
	}
	for _, apiservice := range l.Items {
		if apiservice.Spec.Service != nil {
			ret.Insert(apiservice.Spec.Group)
		}
	}
	return ret, nil
}
