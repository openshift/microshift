package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	corev1conversions "k8s.io/kubernetes/pkg/apis/core/v1"

	"github.com/openshift/api/image/docker10"
	"github.com/openshift/api/image/dockerpre012"
	v1 "github.com/openshift/api/image/v1"

	"github.com/openshift/openshift-apiserver/pkg/image/apis/image"
	imagedocker10helpers "github.com/openshift/openshift-apiserver/pkg/image/apis/image/docker10"
	imagedockerpre012helpers "github.com/openshift/openshift-apiserver/pkg/image/apis/image/dockerpre012"
)

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		image.Install,
		v1.Install,
		corev1conversions.AddToScheme,
		imagedocker10helpers.Install,
		imagedockerpre012helpers.Install,
		docker10.AddToSchemeInCoreGroup,
		dockerpre012.AddToSchemeInCoreGroup,

		addFieldSelectorKeyConversions,
		AddConversionFuncs,
		RegisterDefaults,
	)
	Install = localSchemeBuilder.AddToScheme
)
