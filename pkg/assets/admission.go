package assets

import (
	"fmt"
	"context"

	arv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	arclientv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
)

var (
	admissionScheme = runtime.NewScheme()
	admissionCodecs = serializer.NewCodecFactory(admissionScheme)
)

func init() {
	if err := arv1.AddToScheme(appsScheme); err != nil {
		panic(err)
	}
}

func admissionRegistrationClient(kubeconfigPath string) *arclientv1.AdmissionregistrationV1Client {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	return arclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "apps-agent"))
}

type vwcApplier struct {
	Client *arclientv1.AdmissionregistrationV1Client
	vwc    *arv1.ValidatingWebhookConfiguration
}

func (d *vwcApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(appsCodecs.UniversalDecoder(arv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	d.vwc = obj.(*arv1.ValidatingWebhookConfiguration)
}

func (d *vwcApplier) Applier() error {
	cache := resourceapply.NewResourceCache()
	_, _, err := resourceapply.ApplyValidatingWebhookConfigurationImproved(context.TODO(), d.Client, assetsEventRecorder, d.vwc, cache)
	return err
}

func ApplyValidatingWebhookConfiguration(filepaths []string, kubeconfigPath string) error {
	applier := &vwcApplier{}
	applier.Client = admissionRegistrationClient(kubeconfigPath)

	lock.Lock()
	defer lock.Unlock()

	for _, file := range filepaths {
		klog.Infof("Applying admission registration asset: %s", file)
		objBytes, err := Asset(file)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", file, err)
		}
		applier.Reader(objBytes, nil, nil)
		if err := applier.Applier(); err != nil {
			klog.Warningf("Failed to apply admission registration asset %s: %v", file, err)
			return err
		}
	}

	return nil
}
