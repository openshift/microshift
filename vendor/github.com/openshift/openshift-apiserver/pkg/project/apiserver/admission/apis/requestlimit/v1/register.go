package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/openshift/openshift-apiserver/pkg/project/apiserver/admission/apis/requestlimit"
)

const (
	GroupName = "project.openshift.io"
)

var (
	GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

	localSchemeBuilder = runtime.NewSchemeBuilder(
		addKnownTypes,
		requestlimit.Install,
	)
	Install = localSchemeBuilder.AddToScheme
)

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&ProjectRequestLimitConfig{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
