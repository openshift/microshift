package assets

import (
	"context"
	"fmt"

	rbacassets "github.com/openshift/microshift/pkg/assets/rbac"
	"github.com/openshift/microshift/pkg/constant"

	"github.com/sirupsen/logrus"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
)

var (
	rbacScheme = runtime.NewScheme()
	rbacCodecs = serializer.NewCodecFactory(rbacScheme)
)

func init() {
	if err := rbacv1.AddToScheme(rbacScheme); err != nil {
		panic(err)
	}
}

type clusterRoleBindingApplier struct {
	client *kubernetes.Clientset
	crb    *rbacv1.ClusterRoleBinding
}

func (crb *clusterRoleBindingApplier) New() {
	restConfig, err := clientcmd.BuildConfigFromFlags("", constant.AdminKubeconfigPath)
	if err != nil {
		panic(err)
	}

	crb.client = kubernetes.NewForConfigOrDie(rest.AddUserAgent(restConfig, "rbac-agent"))
}

func (crb *clusterRoleBindingApplier) Reader(objBytes []byte, _ RenderFunc) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	crb.crb = obj.(*rbacv1.ClusterRoleBinding)
}

func (crb *clusterRoleBindingApplier) Applier() error {
	_, err := crb.client.RbacV1().ClusterRoleBindings().Get(context.TODO(), crb.crb.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := crb.client.RbacV1().ClusterRoleBindings().Create(context.TODO(), crb.crb, metav1.CreateOptions{})
		return err
	}

	return nil
}

type clusterRoleApplier struct {
	client *kubernetes.Clientset
	cr     *rbacv1.ClusterRole
}

func (cr *clusterRoleApplier) New() {
	restConfig, err := clientcmd.BuildConfigFromFlags("", constant.AdminKubeconfigPath)
	if err != nil {
		panic(err)
	}

	cr.client = kubernetes.NewForConfigOrDie(rest.AddUserAgent(restConfig, "rbac-agent"))
}

func (cr *clusterRoleApplier) Reader(objBytes []byte, _ RenderFunc) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	cr.cr = obj.(*rbacv1.ClusterRole)
}

func (cr *clusterRoleApplier) Applier() error {
	_, err := cr.client.RbacV1().ClusterRoles().Get(context.TODO(), cr.cr.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := cr.client.RbacV1().ClusterRoles().Create(context.TODO(), cr.cr, metav1.CreateOptions{})
		return err
	}

	return nil
}

func applyRbac(rbacs []string, applier readerApplier) error {
	lock.Lock()
	defer lock.Unlock()

	for _, rbac := range rbacs {
		logrus.Infof("applying rbac %s", rbac)
		objBytes, err := rbacassets.Asset(rbac)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", rbac, err)
		}
		applier.Reader(objBytes, nil)
		if err := applier.Applier(); err != nil {
			logrus.Warningf("failed to apply rbac %s: %v", rbac, err)
			return err
		}
	}

	return nil
}

func ApplyClusterRoleBindings(rbacs []string) error {
	crb := &clusterRoleBindingApplier{}
	crb.New()
	return applyRbac(rbacs, crb)
}

func ApplyClusterRoles(rbacs []string) error {
	cr := &clusterRoleApplier{}
	cr.New()
	return applyRbac(rbacs, cr)

}
