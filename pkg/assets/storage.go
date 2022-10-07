package assets

import (
	"context"
	"fmt"

	embedded "github.com/openshift/microshift/assets"

	scv1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	scclientv1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
)

var (
	scScheme = runtime.NewScheme()
	scCodecs = serializer.NewCodecFactory(scScheme)
)

func init() {
	if err := scv1.AddToScheme(scScheme); err != nil {
		panic(err)
	}
}

func scClient(kubeconfigPath string) *scclientv1.StorageV1Client {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	return scclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "sc-agent"))
}

type scApplier struct {
	Client *scclientv1.StorageV1Client
	sc     *scv1.StorageClass
}

func (s *scApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(scCodecs.UniversalDecoder(scv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	s.sc = obj.(*scv1.StorageClass)
}
func (s *scApplier) Applier() error {
	_, _, err := resourceapply.ApplyStorageClass(context.TODO(), s.Client, assetsEventRecorder, s.sc)
	return err
}

func applySCs(scs []string, applier readerApplier, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, sc := range scs {
		klog.Infof("Applying sc %s", sc)
		objBytes, err := embedded.Asset(sc)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", sc, err)
		}
		applier.Reader(objBytes, render, params)
		if err := applier.Applier(); err != nil {
			klog.Warningf("Failed to apply sc api %s: %v", sc, err)
			return err
		}
	}

	return nil
}

func ApplyStorageClasses(scs []string, render RenderFunc, params RenderParams, kubeconfigPath string) error {
	sc := &scApplier{}
	sc.Client = scClient(kubeconfigPath)
	return applySCs(scs, sc, render, params)
}

type cdApplier struct {
	Client *scclientv1.StorageV1Client
	cd     *scv1.CSIDriver
}

func (c *cdApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(scCodecs.UniversalDecoder(scv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	c.cd = obj.(*scv1.CSIDriver)
}

func (c *cdApplier) Applier() error {
	_, _, err := resourceapply.ApplyCSIDriver(context.TODO(), c.Client, assetsEventRecorder, c.cd)
	return err
}

func ApplyCSIDrivers(drivers []string, render RenderFunc, params RenderParams, kubeconfigPath string) error {
	applier := &cdApplier{}
	applier.Client = scClient(kubeconfigPath)
	return applyCDs(drivers, applier, render, params)
}

func applyCDs(cds []string, applier readerApplier, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, cd := range cds {
		klog.Infof("Applying csiDriver %s", cd)
		objBytes, err := embedded.Asset(cd)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", cd, err)
		}
		applier.Reader(objBytes, render, params)
		if err := applier.Applier(); err != nil {
			klog.Warningf("Failed to apply CSIDriver api %s: %v", cd, err)
			return err
		}
	}
	return nil
}
