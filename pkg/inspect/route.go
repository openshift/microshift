package inspect

import (
	"os"
	"path"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

func inspectRouteInfo(info *resource.Info, o *InspectOptions) error {
	obj := info.Object

	if unstructureObj, ok := obj.(*unstructured.Unstructured); ok {
		structuredRoute := &routev1.Route{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructureObj.Object, structuredRoute)
		if err != nil {
			return err
		}
		obj = structuredRoute
	}
	if unstructureObjList, ok := obj.(*unstructured.UnstructuredList); ok {
		structuredRouteList := &routev1.RouteList{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructureObjList.Object, structuredRouteList)
		if err != nil {
			return err
		}
		for _, unstructureObj := range unstructureObjList.Items {
			structuredRoute := &routev1.Route{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructureObj.Object, structuredRoute)
			if err != nil {
				return err
			}
			structuredRouteList.Items = append(structuredRouteList.Items, *structuredRoute)
		}

		obj = structuredRouteList
	}

	switch castObj := obj.(type) {
	case *routev1.Route:
		elideRoute(castObj)

	case *routev1.RouteList:
		for i := range castObj.Items {
			elideRoute(&castObj.Items[i])
		}

	case *unstructured.UnstructuredList:

	}

	// save the current object to disk
	dirPath := dirPathForInfo(o.destDir, info)
	filename := filenameForInfo(info)
	// ensure destination path exists
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}
	return o.fileWriter.WriteFromResource(path.Join(dirPath, filename), obj)
}

func elideRoute(route *routev1.Route) {
	if route.Spec.TLS == nil {
		return
	}
	route.Spec.TLS.Key = ""
}
