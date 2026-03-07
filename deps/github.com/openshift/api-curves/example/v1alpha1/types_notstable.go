package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +openshift:compatibility-gen:level=4
// +openshift:api-approved.openshift.io=https://github.com/openshift/api/pull/xxx
// +openshift:file-pattern=cvoRunLevel=0000_50,operatorName=my-operator,operatorOrdering=01

// NotStableConfigType is a stable config type that is TechPreviewNoUpgrade only.
//
// Compatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=notstableconfigtypes,scope=Cluster
// +openshift:enable:FeatureGate=Example
// +kubebuilder:subresource:status
type NotStableConfigType struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec is the specification of the desired behavior of the NotStableConfigType.
	Spec NotStableConfigTypeSpec `json:"spec,omitempty"`
	// status is the most recently observed status of the NotStableConfigType.
	Status NotStableConfigTypeStatus `json:"status,omitempty"`
}

// NotStableConfigTypeSpec is the desired state
type NotStableConfigTypeSpec struct {
	// newField is a field that is tech preview, but because the entire type is gated, there is no marker on the field.
	//
	// +required
	NewField string `json:"newField"`
}

// NotStableConfigTypeStatus defines the observed status of the NotStableConfigType.
type NotStableConfigTypeStatus struct {
	// Represents the observations of a foo's current state.
	// Known .status.conditions.type are: "Available", "Progressing", and "Degraded"
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +openshift:compatibility-gen:level=4

// NotStableConfigTypeList contains a list of NotStableConfigTypes.
//
// Compatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.
type NotStableConfigTypeList struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard list's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []NotStableConfigType `json:"items"`
}
