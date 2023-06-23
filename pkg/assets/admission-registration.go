package assets

import (
	"context"
	"fmt"

	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	admissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	admclientV1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	embedded "github.com/openshift/microshift/assets"
)

var (
	admScheme = runtime.NewScheme()
	admCodecs = serializer.NewCodecFactory(admScheme)
)

type validationWebhookCfg struct {
	Client                  *admclientV1.AdmissionregistrationV1Client
	validationWebhookConfig *admissionregistrationV1.ValidatingWebhookConfiguration
}

func (v validationWebhookCfg) Reader(objBytes []byte, renderFunc RenderFunc, params RenderParams) {
	var err error
	if renderFunc != nil {
		objBytes, err = renderFunc(objBytes, RenderParams{})
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(admCodecs.UniversalDecoder(admissionregistrationV1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	v.validationWebhookConfig = obj.(*admissionregistrationV1.ValidatingWebhookConfiguration)
}

func (v validationWebhookCfg) Applier(ctx context.Context) error {
	_, _, err := resourceapply.ApplyValidatingWebhookConfigurationImproved(ctx, v.Client, assetsEventRecorder, v.validationWebhookConfig, nil)
	return err
}

func admissionRegistrationClient(ctx context.Context, kubeconfigPath string) *admclientV1.AdmissionregistrationV1Client {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		klog.Fatalf("failed to create admission-registration client: %w", err)
	}
	return admclientV1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "admission-registration"))
}

func applyAdmissionRegistration(ctx context.Context, admissionRegistrations []string, applier readerApplier, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, ar := range admissionRegistrations {
		klog.Infof("applying admissionRegistration: %s", ar)
		objBytes, err := embedded.Asset(ar)
		if err != nil {
			return fmt.Errorf("error getting embedded asset %s: %v", ar, err)
		}
		applier.Reader(objBytes, render, params)
		if err := applier.Applier(ctx); err != nil {
			klog.Warningf("failed to apply admissionRegistration object: %s, %v", ar, err)
			return err
		}
	}
	return nil
}

func ApplyValidatingWebhookConfiguration(ctx context.Context, admissionRegistrations []string, kubeconfigPath string) error {
	a := &validationWebhookCfg{}
	a.Client = admissionRegistrationClient(ctx, kubeconfigPath)
	return applyAdmissionRegistration(ctx, admissionRegistrations, a, nil, nil)
}
