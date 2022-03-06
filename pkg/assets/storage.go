package assets

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	scv1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	scclientv1 "k8s.io/client-go/kubernetes/typed/storage/v1"
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
	_, err := s.Client.StorageClasses().Get(context.TODO(), s.sc.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := s.Client.StorageClasses().Create(context.TODO(), s.sc, metav1.CreateOptions{})
		return err
	}
	return nil
}

func applySCs(scs []string, applier readerApplier, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, sc := range scs {
		klog.Infof("Applying sc %s", sc)
		objBytes, err := Asset(sc)
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
