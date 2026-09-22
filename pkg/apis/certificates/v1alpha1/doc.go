// Package v1alpha1 defines the versioned output documents for microshift certs.
// These are CLI documents, not Kubernetes resources, and are not registered
// with the API server. JSON and YAML use the same field names and value types.
// Kubebuilder markers describe schema constraints; unmarshalling alone does
// not enforce them.
//
// +kubebuilder:validation:Required
package v1alpha1
