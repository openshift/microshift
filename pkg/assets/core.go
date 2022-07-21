package assets

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	coreScheme = runtime.NewScheme()
	coreCodecs = serializer.NewCodecFactory(coreScheme)
)

func init() {
	if err := corev1.AddToScheme(coreScheme); err != nil {
		panic(err)
	}
}

type nsApplier struct {
	Client *coreclientv1.CoreV1Client
	ns     *corev1.Namespace
}

func coreClient(kubeconfigPath string) *coreclientv1.CoreV1Client {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	return coreclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "core-agent"))
}

func (ns *nsApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(coreCodecs.UniversalDecoder(corev1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	ns.ns = obj.(*corev1.Namespace)
}

func (ns *nsApplier) Applier() error {
	_, err := ns.Client.Namespaces().Get(context.TODO(), ns.ns.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := ns.Client.Namespaces().Create(context.TODO(), ns.ns, metav1.CreateOptions{})
		return err
	}
	return nil
}

type secretApplier struct {
	Client *coreclientv1.CoreV1Client
	secret *corev1.Secret
}

func (secret *secretApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(coreCodecs.UniversalDecoder(corev1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	secret.secret = obj.(*corev1.Secret)
}

func (secret *secretApplier) Applier() error {
	_, err := secret.Client.Secrets(secret.secret.Namespace).Get(context.TODO(), secret.secret.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := secret.Client.Secrets(secret.secret.Namespace).Create(context.TODO(), secret.secret, metav1.CreateOptions{})
		return err
	}
	return nil
}

type svcApplier struct {
	Client *coreclientv1.CoreV1Client
	svc    *corev1.Service
}

func (svc *svcApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(coreCodecs.UniversalDecoder(corev1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	svc.svc = obj.(*corev1.Service)
}

func (svc *svcApplier) Applier() error {
	_, err := svc.Client.Services(svc.svc.Namespace).Get(context.TODO(), svc.svc.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := svc.Client.Services(svc.svc.Namespace).Create(context.TODO(), svc.svc, metav1.CreateOptions{})
		return err
	}
	return nil
}

type saApplier struct {
	Client *coreclientv1.CoreV1Client
	sa     *corev1.ServiceAccount
}

func (sa *saApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(coreCodecs.UniversalDecoder(corev1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	sa.sa = obj.(*corev1.ServiceAccount)
}

func (sa *saApplier) Applier() error {
	_, err := sa.Client.ServiceAccounts(sa.sa.Namespace).Get(context.TODO(), sa.sa.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := sa.Client.ServiceAccounts(sa.sa.Namespace).Create(context.TODO(), sa.sa, metav1.CreateOptions{})
		return err
	}
	return nil
}

type cmApplier struct {
	Client *coreclientv1.CoreV1Client
	cm     *corev1.ConfigMap
}

func (cm *cmApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(coreCodecs.UniversalDecoder(corev1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	cm.cm = obj.(*corev1.ConfigMap)
}

func (cm *cmApplier) Applier() error {
	_, err := cm.Client.ConfigMaps(cm.cm.Namespace).Get(context.TODO(), cm.cm.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := cm.Client.ConfigMaps(cm.cm.Namespace).Create(context.TODO(), cm.cm, metav1.CreateOptions{})
		return err
	}
	return nil
}

func applyCore(cores []string, applier readerApplier, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, core := range cores {
		klog.Infof("Applying corev1 api %s", core)
		objBytes, err := Asset(core)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", core, err)
		}
		applier.Reader(objBytes, render, params)
		if err := applier.Applier(); err != nil {
			klog.Warningf("Failed to apply corev1 api %s: %v", core, err)
			return err
		}
	}

	return nil
}

func ApplyNamespaces(cores []string, kubeconfigPath string) error {
	ns := &nsApplier{}
	ns.Client = coreClient(kubeconfigPath)
	return applyCore(cores, ns, nil, nil)
}

func ApplyServices(cores []string, render RenderFunc, params RenderParams, kubeconfigPath string) error {
	svc := &svcApplier{}
	svc.Client = coreClient(kubeconfigPath)
	return applyCore(cores, svc, render, params)
}

func ApplyServiceAccounts(cores []string, kubeconfigPath string) error {
	sa := &saApplier{}
	sa.Client = coreClient(kubeconfigPath)
	return applyCore(cores, sa, nil, nil)
}

func ApplyConfigMaps(cores []string, kubeconfigPath string) error {
	cm := &cmApplier{}
	cm.Client = coreClient(kubeconfigPath)
	return applyCore(cores, cm, nil, nil)
}

func ApplyConfigMapWithData(cmPath string, data map[string]string, kubeconfigPath string) error {
	ctx := context.TODO()
	cm := &cmApplier{}
	cm.Client = coreClient(kubeconfigPath)
	if err := applyCore([]string{cmPath}, cm, nil, nil); err != nil {
		return err
	}
	c, err := cm.Client.ConfigMaps(cm.cm.Namespace).Get(ctx, cm.cm.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		c, err = cm.Client.ConfigMaps(cm.cm.Namespace).Create(ctx, cm.cm, metav1.CreateOptions{})
		return err
	}
	c.Data = data
	_, err = cm.Client.ConfigMaps(c.Namespace).Update(ctx, c, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func ApplySecretWithData(secretPath string, data map[string][]byte, kubeconfigPath string) error {
	ctx := context.TODO()
	secret := &secretApplier{}
	secret.Client = coreClient(kubeconfigPath)
	if err := applyCore([]string{secretPath}, secret, nil, nil); err != nil {
		return err
	}
	s, err := secret.Client.Secrets(secret.secret.Namespace).Get(ctx, secret.secret.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		s, err = secret.Client.Secrets(secret.secret.Namespace).Create(ctx, secret.secret, metav1.CreateOptions{})
		return err
	}
	s.Data = data
	_, err = secret.Client.Secrets(s.Namespace).Update(ctx, s, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
