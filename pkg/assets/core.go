package assets

import (
	"context"
	"fmt"

	embedded "github.com/openshift/microshift/assets"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
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

func (ns *nsApplier) Read(objBytes []byte, render RenderFunc, params RenderParams) {
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

func (ns *nsApplier) Handle(ctx context.Context) error {
	_, _, err := resourceapply.ApplyNamespace(ctx, ns.Client, assetsEventRecorder, ns.ns)
	return err
}

type secretApplier struct {
	Client *coreclientv1.CoreV1Client
	secret *corev1.Secret
}

func (secret *secretApplier) Read(objBytes []byte, render RenderFunc, params RenderParams) {
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

func (secret *secretApplier) Handle(ctx context.Context) error {
	_, _, err := resourceapply.ApplySecret(ctx, secret.Client, assetsEventRecorder, secret.secret)
	return err
}

type svcApplier struct {
	Client *coreclientv1.CoreV1Client
	svc    *corev1.Service
}

func (svc *svcApplier) Read(objBytes []byte, render RenderFunc, params RenderParams) {
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

func (svc *svcApplier) Handle(ctx context.Context) error {
	_, _, err := resourceapply.ApplyService(ctx, svc.Client, assetsEventRecorder, svc.svc)
	return err
}

type saApplier struct {
	Client *coreclientv1.CoreV1Client
	sa     *corev1.ServiceAccount
}

func (sa *saApplier) Read(objBytes []byte, render RenderFunc, params RenderParams) {
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

func (sa *saApplier) Handle(ctx context.Context) error {
	_, _, err := resourceapply.ApplyServiceAccount(ctx, sa.Client, assetsEventRecorder, sa.sa)
	return err
}

type cmApplier struct {
	Client *coreclientv1.CoreV1Client
	cm     *corev1.ConfigMap
}

func (cm *cmApplier) Read(objBytes []byte, render RenderFunc, params RenderParams) {
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

func (cm *cmApplier) Handle(ctx context.Context) error {
	_, _, err := resourceapply.ApplyConfigMap(ctx, cm.Client, assetsEventRecorder, cm.cm)
	return err
}

func applyCore(ctx context.Context, cores []string, handler resourceHandler, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, core := range cores {
		klog.Infof("Applying corev1 api %s", core)
		objBytes, err := embedded.Asset(core)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", core, err)
		}
		handler.Read(objBytes, render, params)
		if err := handler.Handle(ctx); err != nil {
			klog.Warningf("Failed to apply corev1 api %s: %v", core, err)
			return err
		}
	}

	return nil
}

func ApplyNamespaces(ctx context.Context, cores []string, kubeconfigPath string) error {
	ns := &nsApplier{}
	ns.Client = coreClient(kubeconfigPath)
	return applyCore(ctx, cores, ns, nil, nil)
}

func ApplyServices(ctx context.Context, cores []string, render RenderFunc, params RenderParams, kubeconfigPath string) error {
	svc := &svcApplier{}
	svc.Client = coreClient(kubeconfigPath)
	return applyCore(ctx, cores, svc, render, params)
}

func ApplyServiceAccounts(ctx context.Context, cores []string, kubeconfigPath string) error {
	sa := &saApplier{}
	sa.Client = coreClient(kubeconfigPath)
	return applyCore(ctx, cores, sa, nil, nil)
}

func ApplyConfigMaps(ctx context.Context, cores []string, render RenderFunc, params RenderParams, kubeconfigPath string) error {
	cm := &cmApplier{}
	cm.Client = coreClient(kubeconfigPath)
	return applyCore(ctx, cores, cm, render, params)
}

func ApplyConfigMapWithData(ctx context.Context, cmPath string, data map[string]string, kubeconfigPath string) error {
	cm := &cmApplier{}
	cm.Client = coreClient(kubeconfigPath)
	cmBytes, err := embedded.Asset(cmPath)
	if err != nil {
		return err
	}
	cm.Read(cmBytes, nil, nil)
	cm.cm.Data = data
	_, _, err = resourceapply.ApplyConfigMap(ctx, cm.Client, assetsEventRecorder, cm.cm)
	return err
}

func ApplySecretWithData(ctx context.Context, secretPath string, data map[string][]byte, kubeconfigPath string) error {
	secret := &secretApplier{}
	secret.Client = coreClient(kubeconfigPath)
	secretBytes, err := embedded.Asset(secretPath)
	if err != nil {
		return err
	}
	secret.Read(secretBytes, nil, nil)
	secret.secret.Data = data
	_, _, err = resourceapply.ApplySecret(ctx, secret.Client, assetsEventRecorder, secret.secret)
	return err
}
