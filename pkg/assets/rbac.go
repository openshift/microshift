package assets

import (
	"context"
	"fmt"

	embedded "github.com/openshift/microshift/assets"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
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

func (crb *clusterRoleBindingApplier) New(kubeconfigPath string) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	crb.client = kubernetes.NewForConfigOrDie(rest.AddUserAgent(restConfig, "rbac-agent"))
}

func (crb *clusterRoleBindingApplier) Reader(objBytes []byte, _ RenderFunc, _ RenderParams) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	crb.crb = obj.(*rbacv1.ClusterRoleBinding)
}

func (crb *clusterRoleBindingApplier) Applier() error {
	_, _, err := resourceapply.ApplyClusterRoleBinding(context.TODO(), crb.client.RbacV1(), assetsEventRecorder, crb.crb)
	return err
}

type clusterRoleApplier struct {
	client *kubernetes.Clientset
	cr     *rbacv1.ClusterRole
}

func (cr *clusterRoleApplier) New(kubeconfigPath string) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	cr.client = kubernetes.NewForConfigOrDie(rest.AddUserAgent(restConfig, "rbac-agent"))
}

func (cr *clusterRoleApplier) Reader(objBytes []byte, _ RenderFunc, _ RenderParams) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	cr.cr = obj.(*rbacv1.ClusterRole)
}

func (cr *clusterRoleApplier) Applier() error {
	_, _, err := resourceapply.ApplyClusterRole(context.TODO(), cr.client.RbacV1(), assetsEventRecorder, cr.cr)
	return err
}

type roleBindingApplier struct {
	client *kubernetes.Clientset
	rb     *rbacv1.RoleBinding
}

func (rb *roleBindingApplier) New(kubeconfigPath string) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	rb.client = kubernetes.NewForConfigOrDie(rest.AddUserAgent(restConfig, "rbac-agent"))
}

func (rb *roleBindingApplier) Reader(objBytes []byte, _ RenderFunc, _ RenderParams) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	rb.rb = obj.(*rbacv1.RoleBinding)
}

func (rb *roleBindingApplier) Applier() error {
	_, _, err := resourceapply.ApplyRoleBinding(context.TODO(), rb.client.RbacV1(), assetsEventRecorder, rb.rb)
	return err
}

type roleApplier struct {
	client *kubernetes.Clientset
	r      *rbacv1.Role
}

func (r *roleApplier) New(kubeconfigPath string) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	r.client = kubernetes.NewForConfigOrDie(rest.AddUserAgent(restConfig, "rbac-agent"))
}

func (r *roleApplier) Reader(objBytes []byte, _ RenderFunc, _ RenderParams) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	r.r = obj.(*rbacv1.Role)
}

func (r *roleApplier) Applier() error {
	_, _, err := resourceapply.ApplyRole(context.TODO(), r.client.RbacV1(), assetsEventRecorder, r.r)
	return err
}

func applyRbac(rbacs []string, applier readerApplier) error {
	lock.Lock()
	defer lock.Unlock()

	for _, rbac := range rbacs {
		klog.Infof("Applying rbac %s", rbac)
		objBytes, err := embedded.Asset(rbac)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", rbac, err)
		}
		applier.Reader(objBytes, nil, nil)
		if err := applier.Applier(); err != nil {
			klog.Warningf("Failed to apply rbac %s: %v", rbac, err)
			return err
		}
	}

	return nil
}

func ApplyClusterRoleBindings(rbacs []string, kubeconfigPath string) error {
	crb := &clusterRoleBindingApplier{}
	crb.New(kubeconfigPath)
	return applyRbac(rbacs, crb)
}

func ApplyClusterRoles(rbacs []string, kubeconfigPath string) error {
	cr := &clusterRoleApplier{}
	cr.New(kubeconfigPath)
	return applyRbac(rbacs, cr)
}
func ApplyRoleBindings(rbacs []string, kubeconfigPath string) error {
	rb := &roleBindingApplier{}
	rb.New(kubeconfigPath)
	return applyRbac(rbacs, rb)
}

func ApplyRoles(rbacs []string, kubeconfigPath string) error {
	r := &roleApplier{}
	r.New(kubeconfigPath)
	return applyRbac(rbacs, r)
}
