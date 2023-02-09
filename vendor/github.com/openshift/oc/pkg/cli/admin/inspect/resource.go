package inspect

import (
	"fmt"
	routev1 "github.com/openshift/api/route/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"os"
	"path"

	configv1 "github.com/openshift/api/config/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog/v2"
)

const (
	clusterScopedResourcesDirname = "cluster-scoped-resources"
	namespaceResourcesDirname     = "namespaces"

	configResourceDataKey   = "/cluster-scoped-resources/config.openshift.io"
	operatorResourceDataKey = "/cluster-scoped-resources/operator.openshift.io"
)

// InspectResource receives an object to gather debugging data for, and a context to keep track of
// already-seen objects when following related-object reference chains.
func InspectResource(info *resource.Info, context *resourceContext, o *InspectOptions) error {
	if context.visited.Has(infoToContextKey(info)) {
		klog.V(1).Infof("Skipping previously-inspected resource: %q ...", infoToContextKey(info))
		return nil
	}
	context.visited.Insert(infoToContextKey(info))

	switch info.ResourceMapping().Resource.GroupResource() {
	case configv1.GroupVersion.WithResource("clusteroperators").GroupResource():
		unstr, ok := info.Object.(*unstructured.Unstructured)
		if !ok {
			return fmt.Errorf("unexpected type. Expecting %q but got %T", "*unstructured.Unstructured", info.Object)
		}

		// first, gather config.openshift.io resource data
		errs := []error{}
		if err := o.gatherConfigResourceData(path.Join(o.DestDir, "/cluster-scoped-resources/config.openshift.io"), context); err != nil {
			errs = append(errs, err)
		}

		// then, gather operator.openshift.io resource data
		if err := o.gatherOperatorResourceData(path.Join(o.DestDir, "/cluster-scoped-resources/operator.openshift.io"), context); err != nil {
			errs = append(errs, err)
		}

		// save clusteroperator resources to disk
		if err := gatherClusterOperatorResource(o.DestDir, unstr, o.fileWriter); err != nil {
			errs = append(errs, err)
		}

		// obtain associated objects for the current resource
		if err := gatherRelatedObjects(context, unstr, o); err != nil {
			errs = append(errs, err)
		}

		return errors.NewAggregate(errs)

	case corev1.SchemeGroupVersion.WithResource("namespaces").GroupResource():
		errs := []error{}
		if err := o.gatherNamespaceData(o.DestDir, info.Name); err != nil {
			errs = append(errs, err)
		}
		resourcesToCollect := namespaceResourcesToCollect()
		for _, resource := range resourcesToCollect {
			if context.visited.Has(resourceToContextKey(resource, info.Name)) {
				continue
			}
			resourceInfos, err := groupResourceToInfos(o.configFlags, resource, info.Name, context.serverResources)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			for _, resourceInfo := range resourceInfos {
				if err := InspectResource(resourceInfo, context, o); err != nil {
					errs = append(errs, err)
					continue
				}
			}
		}

		return errors.NewAggregate(errs)

	case corev1.SchemeGroupVersion.WithResource("secrets").GroupResource():
		if err := inspectSecretInfo(info, o); err != nil {
			return err
		}
		return nil

	case routev1.GroupVersion.WithResource("routes").GroupResource():
		if err := inspectRouteInfo(info, o); err != nil {
			return err
		}
		return nil

	case admissionregistrationv1.SchemeGroupVersion.WithResource("mutatingwebhookconfigurations").GroupResource():
		if err := gatherMutatingAdmissionWebhook(context, info, o); err != nil {
			return err
		}
		return nil

	case admissionregistrationv1.SchemeGroupVersion.WithResource("validatingwebhookconfigurations").GroupResource():
		if err := gatherValidatingAdmissionWebhook(context, info, o); err != nil {
			return err
		}
		return nil

	case apiextensionsv1.SchemeGroupVersion.WithResource("customresourcedefinitions").GroupResource():
		if err := gatherCustomResourceDefinition(context, info, o); err != nil {
			return err
		}
		return nil

	default:
		return gatherGenericObject(context, info, o)
	}
}

type listAccessor interface {
	addItem(obj interface{}) error
}

func newListAccessor(structuredList interface{}) (listAccessor, error) {
	switch castObj := structuredList.(type) {
	case *admissionregistrationv1.MutatingWebhookConfigurationList:
		return &mutatingWebhookConfigList{castObj}, nil

	case *admissionregistrationv1.ValidatingWebhookConfigurationList:
		return &validatingWebhookConfigList{castObj}, nil

	case *corev1.SecretList:
		return &secretList{castObj}, nil

	case *routev1.RouteList:
		return &routeList{castObj}, nil

	default:
		return nil, fmt.Errorf("unhandledStructuredListType: %T", castObj)

	}
}

func toStructuredObject[T any, TList any](obj runtime.Object) (runtime.Object, error) {
	if unstructureObj, ok := obj.(*unstructured.Unstructured); ok {
		var t T
		structuredItem := &t
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructureObj.Object, structuredItem)
		if err != nil {
			return nil, err
		}
		return interface{}(structuredItem).(runtime.Object), nil
	}
	if unstructureObjList, ok := obj.(*unstructured.UnstructuredList); ok {
		var t TList
		structuredList := &t
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructureObjList.Object, structuredList)
		if err != nil {
			return nil, err
		}
		for _, unstructureObj := range unstructureObjList.Items {
			var t T
			structuredItem := &t
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructureObj.Object, structuredItem)
			if err != nil {
				return nil, err
			}
			listMeta, err := newListAccessor(structuredList)
			if err != nil {
				return nil, err
			}
			listMeta.addItem(structuredItem)
		}

		return interface{}(structuredList).(runtime.Object), nil
	}

	return nil, fmt.Errorf("unhandled object type")
}

func gatherGenericObject(context *resourceContext, info *resource.Info, o *InspectOptions) error {
	unstr, ok := info.Object.(*unstructured.Unstructured)
	if ok {
		// obtain associated objects for the current resource
		if err := gatherRelatedObjects(context, unstr, o); err != nil {
			return err
		}
	}

	// save the current object to disk
	dirPath := dirPathForInfo(o.DestDir, info)
	filename := filenameForInfo(info)
	// ensure destination path exists
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}

	return o.fileWriter.WriteFromResource(path.Join(dirPath, filename), info.Object)
}

func gatherRelatedObjects(context *resourceContext, unstr *unstructured.Unstructured, o *InspectOptions) error {
	relatedObjReferences, err := obtainRelatedObjects(unstr)
	if err != nil {
		return err
	}

	return gatherMoreObjects(context, o, relatedObjReferences...)
}

func gatherMoreObjects(context *resourceContext, o *InspectOptions, relatedObjReferences ...*configv1.ObjectReference) error {
	errs := []error{}
	for _, relatedRef := range relatedObjReferences {
		if context.visited.Has(objectRefToContextKey(relatedRef)) {
			continue
		}

		relatedInfos, err := objectReferenceToResourceInfos(o.configFlags, relatedRef)
		if err != nil {
			errs = append(errs, fmt.Errorf("skipping gathering %s due to error: %v", objectReferenceToString(relatedRef), err))
			continue
		}

		for _, relatedInfo := range relatedInfos {
			if err := InspectResource(relatedInfo, context, o); err != nil {
				errs = append(errs, fmt.Errorf("skipping gathering %s due to error: %v", objectReferenceToString(relatedRef), err))
				continue
			}
		}
	}

	return errors.NewAggregate(errs)
}

func gatherNamespaces(context *resourceContext, o *InspectOptions, namespaces ...string) error {
	relatedObjReferences := []*configv1.ObjectReference{}

	for _, namespace := range namespaces {
		relatedObjReferences = append(relatedObjReferences, &configv1.ObjectReference{
			Resource: "namespaces",
			Name:     namespace,
		})
	}

	return gatherMoreObjects(context, o, relatedObjReferences...)
}

func gatherClusterOperatorResource(baseDir string, obj *unstructured.Unstructured, fileWriter *MultiSourceFileWriter) error {
	klog.V(1).Infof("Gathering cluster operator resource data...\n")

	// ensure destination path exists
	destDir := path.Join(baseDir, "/"+clusterScopedResourcesDirname, "/"+obj.GroupVersionKind().Group, "/clusteroperators")
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.yaml", obj.GetName())
	return fileWriter.WriteFromResource(path.Join(destDir, "/"+filename), obj)
}

func obtainRelatedObjects(obj *unstructured.Unstructured) ([]*configv1.ObjectReference, error) {
	// obtain related namespace info for the current resource
	klog.V(1).Infof("Gathering related object reference information for %q...\n", unstructuredToString(obj))

	val, found, err := unstructured.NestedSlice(obj.Object, "status", "relatedObjects")
	if !found || err != nil {
		klog.V(1).Infof("%q does not contain .status.relatedObjects", unstructuredToString(obj))
		return nil, nil
	}

	relatedObjs := []*configv1.ObjectReference{}
	for _, relatedObj := range val {
		ref := &configv1.ObjectReference{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(relatedObj.(map[string]interface{}), ref); err != nil {
			return nil, err
		}
		relatedObjs = append(relatedObjs, ref)
		klog.V(1).Infof("    Found related object %q...\n", objectReferenceToString(ref))
	}

	return relatedObjs, nil
}
