package docker10

import (
	"k8s.io/apimachinery/pkg/runtime"
	corev1conversions "k8s.io/kubernetes/pkg/apis/core/v1"

	"github.com/openshift/api/image/docker10"
	"github.com/openshift/openshift-apiserver/pkg/image/apis/image"
)

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		image.Install,
		corev1conversions.AddToScheme,
		docker10.AddToScheme,
	)
	Install = localSchemeBuilder.AddToScheme
)
