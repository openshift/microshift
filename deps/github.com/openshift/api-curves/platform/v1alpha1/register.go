package v1alpha1

// PlatformOperator was removed from the OpenShift API in 4.16.
// See
// * https://issues.redhat.com/browse/OPRUN-3336
// * https://github.com/openshift/platform-operators/pull/113
// This API specficiation is kept here for historical reference only.
// It is not used in the codebase and will not be used to generate code by code-gen.

// import (
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apimachinery/pkg/runtime/schema"
// )

// var (
// 	GroupName     = "platform.openshift.io"
// 	GroupVersion  = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}
// 	schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
// 	// Install is a function which adds this version to a scheme
// 	Install = schemeBuilder.AddToScheme
// )

// // Adds the list of known types to api.Scheme.
// func addKnownTypes(scheme *runtime.Scheme) error {
// 	scheme.AddKnownTypes(GroupVersion,
// 		&PlatformOperator{},
// 		&PlatformOperatorList{},
// 	)
// 	metav1.AddToGroupVersion(scheme, GroupVersion)
// 	return nil
// }
