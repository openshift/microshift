// Package v1alpha1 defines the versioned output documents for microshift certs.
// These Kubernetes API objects are registered in a local scheme for CLI
// serialization, but are not served or persisted by the API server.
// JSON and YAML use the same field names and value types.
// Kubebuilder markers describe schema constraints; unmarshalling alone does
// not enforce them.
//
// +kubebuilder:validation:Required
// +kubebuilder:object:generate=true
// +groupName=microshift.openshift.io
// +k8s:deepcopy-gen=package
package v1alpha1
