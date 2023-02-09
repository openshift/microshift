package inspect

import (
	"fmt"
	"os"
	"path"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/cli-runtime/pkg/resource"
)

type routeList struct {
	*routev1.RouteList
}

func (c *routeList) addItem(obj interface{}) error {
	structuredItem, ok := obj.(*routev1.Route)
	if !ok {
		return fmt.Errorf("unhandledStructuredItemType: %T", obj)
	}
	c.Items = append(c.Items, *structuredItem)
	return nil
}

func inspectRouteInfo(info *resource.Info, o *InspectOptions) error {
	structuredObj, err := toStructuredObject[routev1.Route, routev1.RouteList](info.Object)
	if err != nil {
		return err
	}

	switch castObj := structuredObj.(type) {
	case *routev1.Route:
		elideRoute(castObj)

	case *routev1.RouteList:
		for i := range castObj.Items {
			elideRoute(&castObj.Items[i])
		}
	}

	// save the current object to disk
	dirPath := dirPathForInfo(o.DestDir, info)
	filename := filenameForInfo(info)
	// ensure destination path exists
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}
	return o.fileWriter.WriteFromResource(path.Join(dirPath, filename), structuredObj)
}

func elideRoute(route *routev1.Route) {
	if route.Spec.TLS == nil {
		return
	}
	route.Spec.TLS.Key = ""
}
