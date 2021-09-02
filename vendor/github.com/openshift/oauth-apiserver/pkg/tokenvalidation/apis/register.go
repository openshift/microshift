package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	kauthenticationv1 "k8s.io/api/authentication/v1"
	kauthentication "k8s.io/kubernetes/pkg/apis/authentication"
	kauthv1internal "k8s.io/kubernetes/pkg/apis/authentication/v1"
)

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		kauthentication.AddToScheme,
		kauthenticationv1.AddToScheme,
	)
	Install = localSchemeBuilder.AddToScheme
)

func init() {
	localSchemeBuilder.Register(kauthv1internal.RegisterConversions)
}
