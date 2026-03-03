package v1alpha1

// PlatformOperator was removed from the OpenShift API in 4.16.
// See
// * https://issues.redhat.com/browse/OPRUN-3336
// * https://github.com/openshift/platform-operators/pull/113
// This API specficiation is kept here for historical reference only.
// It is not used in the codebase and will not be used to generate code by code-gen.

// import (
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// // Package contains fields to configure which OLM package this PlatformOperator will install
// type Package struct {
// 	// name contains the desired OLM-based Operator package name
// 	// that is defined in an existing CatalogSource resource in the cluster.
// 	//
// 	// This configured package will be managed with the cluster's lifecycle. In
// 	// the current implementation, it will be retrieving this name from a list of
// 	// supported operators out of the catalogs included with OpenShift.
// 	// +required
// 	//
// 	// +kubebuilder:validation:Pattern:=[a-z0-9]([-a-z0-9]*[a-z0-9])?
// 	// +kubebuilder:validation:MaxLength:=56
// 	// ---
// 	// + More restrictions to package names supported is an intentional design
// 	// + decision that, while limiting to user options, allows code built on these
// 	// + API's to make more confident assumptions on data structure.
// 	Name string `json:"name"`
// }

// // PlatformOperatorSpec defines the desired state of PlatformOperator.
// type PlatformOperatorSpec struct {
// 	// package contains the desired package and its configuration for this
// 	// PlatformOperator.
// 	// +required
// 	Package Package `json:"package"`
// }

// // ActiveBundleDeployment references a BundleDeployment resource.
// type ActiveBundleDeployment struct {
// 	// name is the metadata.name of the referenced BundleDeployment object.
// 	// +required
// 	Name string `json:"name"`
// }

// // PlatformOperatorStatus defines the observed state of PlatformOperator
// type PlatformOperatorStatus struct {
// 	// conditions represent the latest available observations of a platform operator's current state.
// 	// +optional
// 	// +listType=map
// 	// +listMapKey=type
// 	Conditions []metav1.Condition `json:"conditions,omitempty"`

// 	// activeBundleDeployment is the reference to the BundleDeployment resource that's
// 	// being managed by this PO resource. If this field is not populated in the status
// 	// then it means the PlatformOperator has either not been installed yet or is
// 	// failing to install.
// 	// +optional
// 	ActiveBundleDeployment ActiveBundleDeployment `json:"activeBundleDeployment,omitempty"`
// }

// // +genclient
// // +genclient:nonNamespaced
// //+kubebuilder:object:root=true
// //+kubebuilder:subresource:status
// // +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// // +kubebuilder:resource:path=platformoperators,scope=Cluster
// // +openshift:api-approved.openshift.io=https://github.com/openshift/api/pull/1234
// // +openshift:enable:FeatureGate=PlatformOperators
// // +kubebuilder:metadata:annotations=include.release.openshift.io/self-managed-high-availability=true
// // +kubebuilder:metadata:annotations="exclude.release.openshift.io/internal-openshift-hosted=true"

// // PlatformOperator is the Schema for the PlatformOperators API.
// //
// // Compatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.
// // +openshift:compatibility-gen:level=4
// type PlatformOperator struct {
// 	metav1.TypeMeta `json:",inline"`

// 	// metadata is the standard object's metadata.
// 	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
// 	metav1.ObjectMeta `json:"metadata,omitempty"`

// 	Spec   PlatformOperatorSpec   `json:"spec"`
// 	Status PlatformOperatorStatus `json:"status,omitempty"`
// }

// // +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// // PlatformOperatorList contains a list of PlatformOperators
// //
// // Compatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.
// // +openshift:compatibility-gen:level=4
// type PlatformOperatorList struct {
// 	metav1.TypeMeta `json:",inline"`

// 	// metadata is the standard list's metadata.
// 	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
// 	metav1.ListMeta `json:"metadata,omitempty"`

// 	Items []PlatformOperator `json:"items"`
// }
