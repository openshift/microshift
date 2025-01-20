package assets

import (
	sccv1 "github.com/openshift/api/security/v1"
	arv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	scv1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var scheme = runtime.NewScheme()

func init() {
	if err := corev1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := scv1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := rbacv1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := sccv1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := arv1.AddToScheme(scheme); err != nil {
		panic(err)
	}
}
