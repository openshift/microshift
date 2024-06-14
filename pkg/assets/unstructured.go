package assets

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	embedded "github.com/openshift/microshift/assets"
	"github.com/openshift/microshift/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

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
}

// configClientCache is a cache of dynamic clients and REST mappers for each kubeconfig path.
// This is used to avoid creating a new client and REST mapper for each resource.
var configClientCache = make(map[string]configClientCacheEntry, 1)
var configClientCacheLock sync.RWMutex

type unstructuredApplier struct {
	base   dynamic.Interface
	mapper meta.ResettableRESTMapper

	Client       dynamic.ResourceInterface
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

func unstructuredConfigAndClient(kubeconfigPath string) configClientCacheEntry {
	configClientCacheLock.RLock()
	entry, ok := configClientCache[kubeconfigPath]
	configClientCacheLock.RUnlock()
	if ok {
		return entry
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}
	restConfig = rest.AddUserAgent(restConfig, "generic-microshift-agent")

	httpClient, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		panic(err)
	}

	base, err := dynamic.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		panic(err)
	}

	disco, err := discovery.NewDiscoveryClientForConfigAndClient(restConfig, httpClient)
	if err != nil {
		panic(err)
	}

	entry = configClientCacheEntry{
		base:   base,
		mapper: restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disco)),
	}

	configClientCacheLock.Lock()
	configClientCache[kubeconfigPath] = entry
	configClientCacheLock.Unlock()

	return entry
}

func (d *unstructuredApplier) Read(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}

	unstruct, err := util.ConvertYAMLOrJSONToUnstructured(bytes.NewReader(objBytes))
	if err != nil {
		panic(fmt.Errorf("failed to parse LVMCluster: %w", err))
	}

	d.unstructured = unstruct

	gvk := d.unstructured.GroupVersionKind()
	mapping, err := d.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		panic(fmt.Errorf("failed to get RESTMapping for %s: %w", unstruct.GetObjectKind(), err))
	}
	d.unstructured.SetGroupVersionKind(mapping.GroupVersionKind)

	if d.unstructured.GetNamespace() != "" {
		d.Client = d.base.Resource(mapping.Resource).Namespace(d.unstructured.GetNamespace())
	} else {
		d.Client = d.base.Resource(mapping.Resource)
	}
}

func (d *unstructuredApplier) Handle(ctx context.Context) error {
	existing, err := d.Client.Get(ctx, d.unstructured.GetName(), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if _, err := d.Client.Create(ctx, d.unstructured, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create %s: %w", d.unstructured.GroupVersionKind(), err)
		}
		klog.Infof("Created %s: %s", d.unstructured.GroupVersionKind(), d.unstructured.GetName())
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to verify existance of %s: %w", d.unstructured.GroupVersionKind(), err)
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

func applyGeneric(ctx context.Context, resources []string, handler resourceHandler, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, resource := range resources {
		klog.Infof("Applying resource %s", resource)
		objBytes, err := embedded.Asset(resource)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", resource, err)
		}
		handler.Read(objBytes, render, params)
		if err := handler.Handle(ctx); err != nil {
			klog.Warningf("Failed to apply resource %s: %v", resource, err)
			return err
		}
	}

	return nil
}

func ApplyGeneric(
	ctx context.Context,
	resources []string,
	render RenderFunc,
	params RenderParams,
	modify ModifyOnExists,
	kubeconfigPath string,
) error {
	configAndClient := unstructuredConfigAndClient(kubeconfigPath)
	applier := &unstructuredApplier{
		base:           configAndClient.base,
		mapper:         configAndClient.mapper,
		ModifyOnExists: modify,
	}
	return applyGeneric(ctx, resources, applier, render, params)
}
