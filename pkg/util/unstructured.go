package util

import (
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
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
