package machineconfiguration

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	machineconfigurationv1 "github.com/openshift/api/machineconfiguration/v1"
)

// GroupName defines the API group for machineconfiguration.
const GroupName = "machineconfiguration.openshift.io"

var (
	SchemeBuilder = runtime.NewSchemeBuilder(machineconfigurationv1.Install)
	// Install is a function which adds every version of this group to a scheme
	Install = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return schema.GroupResource{Group: GroupName, Resource: resource}
}

func Kind(kind string) schema.GroupKind {
	return schema.GroupKind{Group: GroupName, Kind: kind}
}
