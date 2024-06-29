package util

import (
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// ConvertYAMLOrJSONToUnstructured converts a YAML or JSON stream to an unstructured object.
// It returns an error if the stream cannot be decoded or the object cannot be converted to unstructured.
// The function is used to parse Kubernetes resources from files that are unknown at compile time.
func ConvertYAMLOrJSONToUnstructured(reader io.Reader) (*unstructured.Unstructured, error) {
	raw := map[string]interface{}{}

	if err := yaml.NewYAMLOrJSONDecoder(reader, 0 /* no support for JSON streams */).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode object: %w", err)
	}

	unstruct := unstructured.Unstructured{}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw, &unstruct); err != nil {
		return nil, fmt.Errorf("failed to convert object to unstructured: %w", err)
	}

	return &unstruct, nil
}

func ClientForDynamicObjectKind(config *rest.Config, object schema.ObjectKind) (dynamic.NamespaceableResourceInterface, error) {
	gvk := object.GroupVersionKind()

	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}

	clnt, err := dynamic.NewForConfigAndClient(config, httpClient)
	if err != nil {
		return nil, err
	}

	disco, err := discovery.NewDiscoveryClientForConfigAndClient(config, httpClient)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disco))

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	object.SetGroupVersionKind(mapping.GroupVersionKind)

	return clnt.Resource(mapping.Resource), nil
}
