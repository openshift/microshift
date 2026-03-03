package manifestmerge

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// noopConverter is a converter that only sets the apiVersion fields, but does not real conversion.
type noopConverter struct {
}

func (c noopConverter) Convert(in, out, context interface{}) error {
	realOut, err := c.convert(in.(runtime.Object), context.(schema.GroupVersion))
	if err != nil {
		return err
	}
	out = realOut
	return nil
}

func (c noopConverter) ConvertToVersion(in runtime.Object, gv runtime.GroupVersioner) (out runtime.Object, err error) {
	return in, nil
}

func (c noopConverter) ConvertFieldLabel(gvk schema.GroupVersionKind, label, value string) (string, string, error) {
	return label, value, nil
}

// ConvertToVersion converts in object to the given gv in place and returns the same `in` object.
func (c noopConverter) convert(in runtime.Object, targetGV schema.GroupVersion) (runtime.Object, error) {
	// Run the converter on the list items instead of list itself
	if list, ok := in.(*unstructured.UnstructuredList); ok {
		for i := range list.Items {
			list.Items[i].SetGroupVersionKind(targetGV.WithKind(list.Items[i].GroupVersionKind().Kind))
		}
	}
	in.GetObjectKind().SetGroupVersionKind(targetGV.WithKind(in.GetObjectKind().GroupVersionKind().Kind))
	return in, nil
}
