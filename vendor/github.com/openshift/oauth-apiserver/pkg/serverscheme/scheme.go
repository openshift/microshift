package serverscheme

import (
	oauthinstall "github.com/openshift/oauth-apiserver/pkg/oauth/apis/oauth/install"
	kauthinstall "github.com/openshift/oauth-apiserver/pkg/tokenvalidation/apis"
	userinstall "github.com/openshift/oauth-apiserver/pkg/user/apis/user/install"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	oauthinstall.Install(Scheme)
	userinstall.Install(Scheme)
	kauthinstall.Install(Scheme)
}
