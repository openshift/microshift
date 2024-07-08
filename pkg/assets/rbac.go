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
	rbacScheme = scheme
	rbacCodecs = serializer.NewCodecFactory(rbacScheme)
)

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

func (crb *clusterRoleBindingApplier) Read(objBytes []byte, _ RenderFunc, _ RenderParams) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	crb.crb = obj.(*rbacv1.ClusterRoleBinding)
}

func (crb *clusterRoleBindingApplier) Handle(ctx context.Context) error {
	_, _, err := resourceapply.ApplyClusterRoleBinding(ctx, crb.client.RbacV1(), assetsEventRecorder, crb.crb)
	return err
}

type clusterRoleBindingDeleter struct {
	client *kubernetes.Clientset
	crb    *rbacv1.ClusterRoleBinding
}

func (crb *clusterRoleBindingDeleter) New(kubeconfigPath string) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	crb.client = kubernetes.NewForConfigOrDie(rest.AddUserAgent(restConfig, "rbac-agent"))
}

func (crb *clusterRoleBindingDeleter) Read(objBytes []byte, _ RenderFunc, _ RenderParams) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	crb.crb = obj.(*rbacv1.ClusterRoleBinding)
}

func (crb *clusterRoleBindingDeleter) Handle(ctx context.Context) error {
	_, _, err := resourceapply.DeleteClusterRoleBinding(ctx, crb.client.RbacV1(), assetsEventRecorder, crb.crb)
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

func (cr *clusterRoleApplier) Read(objBytes []byte, _ RenderFunc, _ RenderParams) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	cr.cr = obj.(*rbacv1.ClusterRole)
}

func (cr *clusterRoleApplier) Handle(ctx context.Context) error {
	_, _, err := resourceapply.ApplyClusterRole(ctx, cr.client.RbacV1(), assetsEventRecorder, cr.cr)
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

func (rb *roleBindingApplier) Read(objBytes []byte, _ RenderFunc, _ RenderParams) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	rb.rb = obj.(*rbacv1.RoleBinding)
}

func (rb *roleBindingApplier) Handle(ctx context.Context) error {
	_, _, err := resourceapply.ApplyRoleBinding(ctx, rb.client.RbacV1(), assetsEventRecorder, rb.rb)
	return err
}

type clusterRoleDeleter struct {
	client *kubernetes.Clientset
	cr     *rbacv1.ClusterRole
}

func (cr *clusterRoleDeleter) New(kubeconfigPath string) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	cr.client = kubernetes.NewForConfigOrDie(rest.AddUserAgent(restConfig, "rbac-agent"))
}

func (cr *clusterRoleDeleter) Read(objBytes []byte, _ RenderFunc, _ RenderParams) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	cr.cr = obj.(*rbacv1.ClusterRole)
}

func (cr *clusterRoleDeleter) Handle(ctx context.Context) error {
	_, _, err := resourceapply.DeleteClusterRole(ctx, cr.client.RbacV1(), assetsEventRecorder, cr.cr)
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

func (r *roleApplier) Read(objBytes []byte, _ RenderFunc, _ RenderParams) {
	obj, err := runtime.Decode(rbacCodecs.UniversalDecoder(rbacv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	r.r = obj.(*rbacv1.Role)
}

func (r *roleApplier) Handle(ctx context.Context) error {
	_, _, err := resourceapply.ApplyRole(ctx, r.client.RbacV1(), assetsEventRecorder, r.r)
	return err
}

func handleRbac(ctx context.Context, rbacs []string, handler resourceHandler) error {
	lock.Lock()
	defer lock.Unlock()

	for _, rbac := range rbacs {
		klog.Infof("Handling rbac %s", rbac)
		objBytes, err := embedded.Asset(rbac)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", rbac, err)
		}
		handler.Read(objBytes, nil, nil)
		if err := handler.Handle(ctx); err != nil {
			klog.Warningf("Failed to handle rbac %s: %v", rbac, err)
			return err
		}
	}

	return nil
}

func ApplyClusterRoleBindings(ctx context.Context, rbacs []string, kubeconfigPath string) error {
	crb := &clusterRoleBindingApplier{}
	crb.New(kubeconfigPath)
	return handleRbac(ctx, rbacs, crb)
}

func DeleteClusterRoleBindings(ctx context.Context, rbacs []string, kubeconfigPath string) error {
	crb := &clusterRoleBindingDeleter{}
	crb.New(kubeconfigPath)
	return handleRbac(ctx, rbacs, crb)
}

func ApplyClusterRoles(ctx context.Context, rbacs []string, kubeconfigPath string) error {
	cr := &clusterRoleApplier{}
	cr.New(kubeconfigPath)
	return handleRbac(ctx, rbacs, cr)
}

func DeleteClusterRoles(ctx context.Context, rbacs []string, kubeconfigPath string) error {
	cr := &clusterRoleDeleter{}
	cr.New(kubeconfigPath)
	return handleRbac(ctx, rbacs, cr)
}

func ApplyRoleBindings(ctx context.Context, rbacs []string, kubeconfigPath string) error {
	rb := &roleBindingApplier{}
	rb.New(kubeconfigPath)
	return handleRbac(ctx, rbacs, rb)
}

func ApplyRoles(ctx context.Context, rbacs []string, kubeconfigPath string) error {
	r := &roleApplier{}
	r.New(kubeconfigPath)
	return handleRbac(ctx, rbacs, r)
}
