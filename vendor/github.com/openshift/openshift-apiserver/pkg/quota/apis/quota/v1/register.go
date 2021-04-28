package v1

import (
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/openshift/api/quota/v1"
	"github.com/openshift/openshift-apiserver/pkg/quota/apis/quota"
)

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		quota.Install,
		v1.Install,
		RegisterDefaults,
	)
	Install = localSchemeBuilder.AddToScheme
)
