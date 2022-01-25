package assets

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

	appsassets "github.com/openshift/microshift/pkg/assets/apps"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

var (
	appsScheme = runtime.NewScheme()
	appsCodecs = serializer.NewCodecFactory(appsScheme)
)

func init() {
	if err := appsv1.AddToScheme(appsScheme); err != nil {
		panic(err)
	}
}

func appsClient(kubeconfigPath string) *appsclientv1.AppsV1Client {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	return appsclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "apps-agent"))
}

type dpApplier struct {
	Client *appsclientv1.AppsV1Client
	dp     *appsv1.Deployment
}

func (d *dpApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(appsCodecs.UniversalDecoder(appsv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	d.dp = obj.(*appsv1.Deployment)
}
func (d *dpApplier) Applier() error {
	_, err := d.Client.Deployments(d.dp.Namespace).Get(context.TODO(), d.dp.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := d.Client.Deployments(d.dp.Namespace).Create(context.TODO(), d.dp, metav1.CreateOptions{})
		return err
	}
	return nil
}

type dsApplier struct {
	Client *appsclientv1.AppsV1Client
	ds     *appsv1.DaemonSet
}

func (d *dsApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(appsCodecs.UniversalDecoder(appsv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	d.ds = obj.(*appsv1.DaemonSet)
}
func (d *dsApplier) Applier() error {
	_, err := d.Client.DaemonSets(d.ds.Namespace).Get(context.TODO(), d.ds.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := d.Client.DaemonSets(d.ds.Namespace).Create(context.TODO(), d.ds, metav1.CreateOptions{})
		return err
	}
	return nil
}

func applyApps(apps []string, applier readerApplier, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, app := range apps {
		klog.Infof("Applying apps api %s", app)
		objBytes, err := appsassets.Asset(app)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", app, err)
		}
		applier.Reader(objBytes, render, params)
		if err := applier.Applier(); err != nil {
			klog.Warningf("Failed to apply apps api %s: %v", app, err)
			return err
		}
	}

	return nil
}

func ApplyDeployments(dps []string, render RenderFunc, params RenderParams, kubeconfigPath string) error {
	dp := &dpApplier{}
	dp.Client = appsClient(kubeconfigPath)
	return applyApps(dps, dp, render, params)
}

func ApplyDaemonSets(apps []string, render RenderFunc, params RenderParams, kubeconfigPath string) error {
	ds := &dsApplier{}
	ds.Client = appsClient(kubeconfigPath)
	return applyApps(apps, ds, render, params)
}
