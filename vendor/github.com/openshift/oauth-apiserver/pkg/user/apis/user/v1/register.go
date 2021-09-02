package v1

import (
	"k8s.io/apimachinery/pkg/runtime"

	userv1 "github.com/openshift/api/user/v1"
	"github.com/openshift/oauth-apiserver/pkg/user/apis/user"
)

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		user.Install,
		userv1.Install,

		addFieldSelectorKeyConversions,
		RegisterDefaults,
	)
	Install = localSchemeBuilder.AddToScheme
)
