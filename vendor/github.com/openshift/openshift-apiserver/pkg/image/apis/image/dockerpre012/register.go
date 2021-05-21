package dockerpre012

import (
	"k8s.io/apimachinery/pkg/runtime"
	corev1conversions "k8s.io/kubernetes/pkg/apis/core/v1"

	"github.com/openshift/api/image/dockerpre012"
	"github.com/openshift/openshift-apiserver/pkg/image/apis/image"
)

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		image.Install,
		corev1conversions.AddToScheme,
		dockerpre012.AddToScheme,
	)
	Install = localSchemeBuilder.AddToScheme
)
