package assets

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

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
	_, err := cr.client.RbacV1().ClusterRoles().Get(context.TODO(), cr.cr.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := cr.client.RbacV1().ClusterRoles().Create(context.TODO(), cr.cr, metav1.CreateOptions{})
		return err
	}

	return nil
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
	_, err := rb.client.RbacV1().RoleBindings(rb.rb.Namespace).Get(context.TODO(), rb.rb.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := rb.client.RbacV1().RoleBindings(rb.rb.Namespace).Create(context.TODO(), rb.rb, metav1.CreateOptions{})
		return err
	}

	return nil
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
	_, err := r.client.RbacV1().Roles(r.r.Namespace).Get(context.TODO(), r.r.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := r.client.RbacV1().Roles(r.r.Namespace).Create(context.TODO(), r.r, metav1.CreateOptions{})
		return err
	}

	return nil
}

func applyRbac(rbacs []string, applier readerApplier) error {
	lock.Lock()
	defer lock.Unlock()

	for _, rbac := range rbacs {
		klog.Infof("Applying rbac %s", rbac)
		objBytes, err := Asset(rbac)
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
