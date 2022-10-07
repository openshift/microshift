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
	_, _, err := resourceapply.ApplyNamespace(context.TODO(), ns.Client, assetsEventRecorder, ns.ns)
	return err
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
	_, _, err := resourceapply.ApplySecret(context.TODO(), secret.Client, assetsEventRecorder, secret.secret)
	return err
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
	_, _, err := resourceapply.ApplyService(context.TODO(), svc.Client, assetsEventRecorder, svc.svc)
	return err
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
	_, _, err := resourceapply.ApplyServiceAccount(context.TODO(), sa.Client, assetsEventRecorder, sa.sa)
	return err
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
	_, _, err := resourceapply.ApplyConfigMap(context.TODO(), cm.Client, assetsEventRecorder, cm.cm)
	return err
}

func applyCore(cores []string, applier readerApplier, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, core := range cores {
		klog.Infof("Applying corev1 api %s", core)
		objBytes, err := embedded.Asset(core)
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

func ApplyConfigMaps(cores []string, render RenderFunc, params RenderParams, kubeconfigPath string) error {
	cm := &cmApplier{}
	cm.Client = coreClient(kubeconfigPath)
	return applyCore(cores, cm, render, params)
}

func ApplyConfigMapWithData(cmPath string, data map[string]string, kubeconfigPath string) error {
	cm := &cmApplier{}
	cm.Client = coreClient(kubeconfigPath)
	cmBytes, err := embedded.Asset(cmPath)
	if err != nil {
		return err
	}
	cm.Reader(cmBytes, nil, nil)
	_, _, err = resourceapply.ApplyConfigMap(context.TODO(), cm.Client, assetsEventRecorder, cm.cm)
	return err
}

func ApplySecretWithData(secretPath string, data map[string][]byte, kubeconfigPath string) error {
	secret := &secretApplier{}
	secret.Client = coreClient(kubeconfigPath)
	secretBytes, err := embedded.Asset(secretPath)
	if err != nil {
		return err
	}
	secret.Reader(secretBytes, nil, nil)
	secret.secret.Data = data
	_, _, err = resourceapply.ApplySecret(context.TODO(), secret.Client, assetsEventRecorder, secret.secret)
	return err
}
