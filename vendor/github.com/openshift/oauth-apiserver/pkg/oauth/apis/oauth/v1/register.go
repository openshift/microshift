package v1

import (
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"

	oauthv1 "github.com/openshift/api/oauth/v1"
	"github.com/openshift/oauth-apiserver/pkg/oauth/apis/oauth"
)

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		oauth.Install,
		oauthv1.Install,
		authenticationv1.AddToScheme,

		addFieldSelectorKeyConversions,
		RegisterDefaults,
	)
	Install = localSchemeBuilder.AddToScheme
)
