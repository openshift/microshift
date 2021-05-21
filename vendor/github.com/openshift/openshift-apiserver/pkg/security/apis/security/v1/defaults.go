package v1

import (
	v1 "github.com/openshift/api/security/v1"
	"github.com/openshift/apiserver-library-go/pkg/securitycontextconstraints/sccdefaults"
	"k8s.io/apimachinery/pkg/runtime"
)

func AddDefaultingFuncs(scheme *runtime.Scheme) error {
	RegisterDefaults(scheme)
	scheme.AddTypeDefaultingFunc(&v1.SecurityContextConstraints{}, func(obj interface{}) {
		sccdefaults.SetDefaults_SCC(obj.(*v1.SecurityContextConstraints))
	})

	return nil
}
