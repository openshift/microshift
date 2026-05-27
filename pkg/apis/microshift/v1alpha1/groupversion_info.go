package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	GroupName    = "microshift.io"
	GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

	SchemeGroupVersion = GroupVersion

	schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = schemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return schema.GroupResource{Group: GroupName, Resource: resource}
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&RemoteCluster{},
		&RemoteClusterList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
