package assets

import (
	"context"
	"fmt"

	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	arV1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	arClientV1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	embedded "github.com/openshift/microshift/assets"
)

type validationWebhookCfg struct {
	Client                  *arClientV1.AdmissionregistrationV1Client
	validationWebhookConfig *arV1.ValidatingWebhookConfiguration
	codec                   serializer.CodecFactory
}

func (v *validationWebhookCfg) Read(objBytes []byte, renderFunc RenderFunc, params RenderParams) {
	var err error
	if renderFunc != nil {
		objBytes, err = renderFunc(objBytes, RenderParams{})
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(v.codec.UniversalDecoder(arV1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	v.validationWebhookConfig = obj.(*arV1.ValidatingWebhookConfiguration)
}

func (v *validationWebhookCfg) Handle(ctx context.Context) error {
	_, _, err := resourceapply.ApplyValidatingWebhookConfigurationImproved(ctx, v.Client, assetsEventRecorder, v.validationWebhookConfig, resourceapply.NewResourceCache())
	return err
}

func admissionRegistrationClient(kubeconfigPath string) *arClientV1.AdmissionregistrationV1Client {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		klog.Fatalf("failed to create admission-registration Client: %v", err)
	}
	return arClientV1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "admission-registration"))
}

func applyAdmissionRegistration(ctx context.Context, admissionRegistrations []string, handler resourceHandler, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, ar := range admissionRegistrations {
		klog.Infof("applying admissionRegistration: %s", ar)
		objBytes, err := embedded.Asset(ar)
		if err != nil {
			return fmt.Errorf("error getting embedded asset %s: %w", ar, err)
		}
		handler.Read(objBytes, render, params)
		if err := handler.Handle(ctx); err != nil {
			klog.Warningf("failed to apply admissionRegistration object: %s, %v", ar, err)
			return err
		}
	}
	return nil
}

func ApplyValidatingWebhookConfiguration(ctx context.Context, admissionRegistrations []string, kubeconfigPath string) error {
	scheme := runtime.NewScheme()
	if err := arV1.AddToScheme(scheme); err != nil {
		return err
	}

	a := &validationWebhookCfg{}
	a.Client = admissionRegistrationClient(kubeconfigPath)
	a.codec = serializer.NewCodecFactory(scheme)
	return applyAdmissionRegistration(ctx, admissionRegistrations, a, nil, nil)
}
