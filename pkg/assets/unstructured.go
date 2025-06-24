package assets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	embedded "github.com/openshift/microshift/assets"
	"github.com/openshift/microshift/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

var defaultingScheme = scheme

// ModifyOnExists is a function that modifies the existing object based on the required object.
// The first argument is a pointer to a boolean that should be set to true if the object was modified.
// i.e. `*modified = true`
// The second argument is a pointer to the existing object on the cluster.
// The third argument is a pointer to the required object from the template.
// If Not set, the object will only have its metadata fields updated.
// If set, the object will be modified based on the function.
type ModifyOnExists func(modified *bool, existing, required *unstructured.Unstructured)

type configClientCacheEntry struct {
	base   dynamic.Interface
	mapper meta.ResettableRESTMapper
	disco  discovery.AggregatedDiscoveryInterface
}

// configClientCache is a cache of dynamic clients and REST mappers for each kubeconfig path.
// This is used to avoid creating a new client and REST mapper for each resource.
var configClientCache = make(map[string]configClientCacheEntry, 1)
var configClientCacheLock sync.RWMutex

type unstructuredClient struct {
	base   dynamic.Interface
	mapper meta.ResettableRESTMapper

	Client    dynamic.ResourceInterface
	Discovery discovery.AggregatedDiscoveryInterface

	unstructured *unstructured.Unstructured

	// modify is a function that modifies the existing object based on the required object.
	// The first argument is a pointer to a boolean that should be set to true if the object was modified.
	// i.e. `*modified = true`
	// The second argument is a pointer to the existing object on the cluster.
	// The third argument is a pointer to the required object from the template.
	// If Not set, the object will only have its metadata fields updated.
	// If set, the object will be modified based on the function.
	ModifyOnExists
}

func unstructuredConfigAndClient(kubeconfigPath string) (configClientCacheEntry, error) {
	configClientCacheLock.RLock()
	entry, ok := configClientCache[kubeconfigPath]
	configClientCacheLock.RUnlock()
	if ok {
		return entry, nil
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return configClientCacheEntry{}, fmt.Errorf("failed to build unstructured rest config: %w", err)
	}
	restConfig = rest.AddUserAgent(restConfig, "generic-microshift-agent")

	httpClient, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		return configClientCacheEntry{}, fmt.Errorf("failed to get HTTP client for unstructured rest config: %w", err)
	}

	base, err := dynamic.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return configClientCacheEntry{}, fmt.Errorf("failed to get dynamic client for unstructured rest config: %w", err)
	}

	disco, err := discovery.NewDiscoveryClientForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return configClientCacheEntry{}, fmt.Errorf("failed to get discovery client for unstructured rest config: %w", err)
	}

	entry = configClientCacheEntry{
		base:   base,
		mapper: restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disco)),
		disco:  disco,
	}

	configClientCacheLock.Lock()
	configClientCache[kubeconfigPath] = entry
	configClientCacheLock.Unlock()

	return entry, nil
}

func (d *unstructuredClient) Read(obj io.Reader, render RenderFuncV2, params RenderParams) error {
	var err error
	if render != nil {
		if obj, err = render(obj, params); err != nil {
			return fmt.Errorf("failed to render object: %w", err)
		}
	}

	unstruct, err := util.ConvertYAMLOrJSONToUnstructured(obj)
	if err != nil {
		return err
	}

	d.unstructured = unstruct

	gvk := d.unstructured.GroupVersionKind()
	mapping, err := d.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("failed to get RESTMapping for %s: %w", unstruct.GetObjectKind(), err)
	}
	d.unstructured.SetGroupVersionKind(mapping.GroupVersionKind)

	if d.unstructured.GetNamespace() != "" {
		d.Client = d.base.Resource(mapping.Resource).Namespace(d.unstructured.GetNamespace())
	} else {
		d.Client = d.base.Resource(mapping.Resource)
	}

	return nil
}

func (d *unstructuredClient) Handle(ctx context.Context) error {
	existing, err := d.Client.Get(ctx, d.unstructured.GetName(), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if _, err := d.Client.Create(ctx, d.unstructured, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create %s: %w", d.unstructured.GroupVersionKind(), err)
		}
		klog.Infof("Created %s: %s", d.unstructured.GroupVersionKind(), d.unstructured.GetName())
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to verify existence of %s: %w", d.unstructured.GroupVersionKind(), err)
	}

	var modified bool

	if d.ModifyOnExists != nil {
		d.ModifyOnExists(&modified, existing, d.unstructured)
	}

	// replicate resourcemerge.EnsureObjectMeta
	// this is a bit of a hack, but it's the only way to get the same behavior as resourceMerge
	resourcemerge.SetStringIfSet(&modified, ptr.To(existing.GetNamespace()), d.unstructured.GetNamespace())
	resourcemerge.SetStringIfSet(&modified, ptr.To(existing.GetName()), d.unstructured.GetName())
	resourcemerge.MergeMap(&modified, ptr.To(existing.GetLabels()), d.unstructured.GetLabels())
	resourcemerge.MergeMap(&modified, ptr.To(existing.GetAnnotations()), d.unstructured.GetAnnotations())
	resourcemerge.MergeOwnerRefs(&modified, ptr.To(existing.GetOwnerReferences()), d.unstructured.GetOwnerReferences())

	if !modified {
		return nil
	}

	if _, err = d.Client.Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update %s: %w", d.unstructured.GroupVersionKind(), err)
	}

	klog.Infof("Updated %s: %s", d.unstructured.GroupVersionKind(), d.unstructured.GetName())

	return nil
}

func applyGeneric(ctx context.Context, resources []string, handler resourceHandlerV2, render RenderFuncV2, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, resource := range resources {
		klog.Infof("Applying resource %s", resource)
		asset, err := embedded.AssetStreamed(resource)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", resource, err)
		}
		// call within IIFE to ensure asset is closed without leak with defer
		if err := func() error {
			defer func() { _ = asset.Close() }()
			if err := handler.Read(asset, render, params); err != nil {
				return fmt.Errorf("failed to read resource %s: %w", resource, err)
			}
			return nil
		}(); err != nil {
			return err
		}
		if err := handler.Handle(ctx); err != nil {
			klog.Warningf("failed to apply resource %s: %v", resource, err)
			return err
		}
	}

	return nil
}

func ApplyGeneric(
	ctx context.Context,
	resources []string,
	render RenderFuncV2,
	params RenderParams,
	modify ModifyOnExists,
	kubeconfigPath string,
) error {
	configAndClient, err := unstructuredConfigAndClient(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to get config and client for generic apply: %w", err)
	}
	clnt := &unstructuredClient{
		base:           configAndClient.base,
		mapper:         configAndClient.mapper,
		ModifyOnExists: modify,
	}
	return applyGeneric(ctx, resources, clnt, render, params)
}

func DeleteGeneric(
	ctx context.Context,
	objects []runtime.Object,
	kubeconfigPath string,
) error {
	configAndClient, err := unstructuredConfigAndClient(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to get config and client for generic delete: %w", err)
	}
	clnt := &unstructuredClient{
		base:      configAndClient.base,
		mapper:    configAndClient.mapper,
		Discovery: configAndClient.disco,
	}
	return clnt.concurrentDeleteGeneric(ctx, objects)
}

// concurrentDeleteGeneric deletes the given objects concurrently.
// It first deletes the objects in parallel, then waits for each object to be deleted.
// If an object is not found, it is skipped.
// If an object is not deleted before the context is cancelled, an error is returned.
// The delete polling period is 1 second due to absence of a watcher.
// The function is thread-safe and can be called concurrently from multiple goroutines.
func (clnt *unstructuredClient) concurrentDeleteGeneric(ctx context.Context, objects []runtime.Object) error {
	lock.Lock()
	defer lock.Unlock()

	type verification struct {
		client   dynamic.ResourceInterface
		accessor metav1.Object
		mapping  *meta.RESTMapping
	}

	skipped := make([]string, 0, len(objects))
	toVerify := make([]verification, 0, len(objects))

	for i, obj := range objects {
		accessor, err := meta.Accessor(obj)
		if err != nil {
			return fmt.Errorf("failed to get accessor for object at index %v for generic delete: %w", i, err)
		}

		if accessor.GetName() == "" {
			return fmt.Errorf("object at index %v has no name and cannot be deleted generically", i)
		}

		gvk := obj.GetObjectKind().GroupVersionKind()
		if gvk, err = defaultGVK(gvk, obj, i); err != nil {
			return err
		} else {
			obj.GetObjectKind().SetGroupVersionKind(gvk)
		}

		mapping, err := clnt.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return fmt.Errorf("failed to get RESTMapping for %s to delete generically: %w", obj.GetObjectKind(), err)
		}

		var client dynamic.ResourceInterface
		if accessor.GetNamespace() != "" {
			client = clnt.base.Resource(mapping.Resource).Namespace(accessor.GetNamespace())
		} else {
			client = clnt.base.Resource(mapping.Resource)
		}

		err = client.Delete(ctx, accessor.GetName(), metav1.DeleteOptions{
			PropagationPolicy: ptr.To(metav1.DeletePropagationBackground),
		})
		if apierrors.IsNotFound(err) {
			skipped = append(skipped, nameFromAccessorAndMapping(accessor, mapping))
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to delete %s %s: %w", obj.GetObjectKind(), accessor.GetName(), err)
		}

		toVerify = append(toVerify, verification{
			client:   client,
			accessor: accessor,
			mapping:  mapping,
		})
	}

	errs := make(chan error, len(toVerify))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	verify := func(ctx context.Context, v verification) {
		exists := func(ctx context.Context) (bool, error) {
			_, err := v.client.Get(ctx, v.accessor.GetName(), metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			if err != nil {
				return false, err
			}
			return false, nil
		}
		err := wait.PollUntilContextCancel(ctx, 1*time.Second, true, exists)
		if err != nil {
			errs <- fmt.Errorf("failed to wait for deletion of %s: %w", v.accessor.GetName(), err)
		} else {
			errs <- nil
			klog.Infof("Deleted %s", nameFromAccessorAndMapping(v.accessor, v.mapping))
		}
	}

	for _, v := range toVerify {
		go verify(ctx, v)
	}

	var err error
	for range toVerify {
		err = errors.Join(err, <-errs)
	}
	if err != nil {
		return err
	}

	if len(skipped) > 0 {
		klog.Infof("Skipped deleting %v resources: %s", len(skipped), strings.Join(skipped, ", "))
	}

	return nil
}

// defaultGVK is a partial clone from https://github.com/kubernetes-sigs/controller-runtime/blob/main/pkg/client/apiutil/apimachinery.go#L95
// This is done to ensure proper scheme based defaulting of GVKs for conversion from objects into REST Mappings.
// This effectively allows someone to specify a core type, e.g. v1.Pod, without having to set the GVK, as its defaulting
// from defaultingScheme will be used.
func defaultGVK(gvk schema.GroupVersionKind, obj runtime.Object, i int) (schema.GroupVersionKind, error) {
	if gvk != schema.EmptyObjectKind.GroupVersionKind() {
		return gvk, nil
	}

	gvks, unversioned, err := defaultingScheme.ObjectKinds(obj)
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("failed to get GroupVersionKind for object at index %v: %w", i, err)
	}

	if unversioned {
		return schema.GroupVersionKind{}, fmt.Errorf("cannot create group-version-kind for unversioned type %T", obj)
	}

	if len(gvks) < 1 {
		return schema.GroupVersionKind{}, fmt.Errorf("no GroupVersionKind associated with Go type %T, was the type registered with the Scheme", obj)
	}

	if len(gvks) > 1 {
		if gvk.Empty() {
			return schema.GroupVersionKind{}, fmt.Errorf("multiple GroupVersionKinds associated with Go type %T within the Scheme, this can happen when a type is registered for multiple GVKs at the same time", obj)
		}
		for _, fromList := range gvks {
			if fromList == gvk {
				break
			}
		}
	} else {
		gvk = gvks[0]
	}

	return gvk, nil
}

func nameFromAccessorAndMapping(accessor metav1.Object, mapping *meta.RESTMapping) string {
	name := accessor.GetName()
	if accessor.GetNamespace() != "" {
		name = fmt.Sprintf("%s/%s", accessor.GetNamespace(), name)
	}
	gvr := fmt.Sprintf("%s/%s", mapping.Resource.GroupVersion().String(), mapping.Resource.Resource)
	return fmt.Sprintf("%s/%s", gvr, name)
}
