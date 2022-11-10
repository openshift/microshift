package assets

import (
	"context"
	"fmt"

	embedded "github.com/openshift/microshift/assets"

	sv1 "k8s.io/api/scheduling/v1"
	scv1 "k8s.io/client-go/kubernetes/typed/scheduling/v1"

	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type pcApplier struct {
	Client *scv1.SchedulingV1Client
	pc     *sv1.PriorityClass
	codecs serializer.CodecFactory
}

func pcClient(kubeconfigPath string) *scv1.SchedulingV1Client {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}
	return scv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "pc-agent"))
}

func (s *pcApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(s.codecs.UniversalDecoder(sv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	s.pc = obj.(*sv1.PriorityClass)
}

func (s *pcApplier) Applier() error {
	// adapted from cvo
	existing, err := s.Client.PriorityClasses().Get(context.TODO(), s.pc.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := s.Client.PriorityClasses().Create(context.TODO(), s.pc, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}

	var modified bool
	resourcemerge.EnsureObjectMeta(&modified, &existing.ObjectMeta, s.pc.ObjectMeta)
	if !modified {
		return nil
	}

	_, err = s.Client.PriorityClasses().Update(context.TODO(), existing, metav1.UpdateOptions{})
	return err
}

func applyPriorityClasses(pcs []string, applier readerApplier) error {
	lock.Lock()
	defer lock.Unlock()

	for _, pc := range pcs {
		klog.Infof("Applying PriorityClass CR %s", pc)
		objBytes, err := embedded.Asset(pc)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", pc, err)
		}
		applier.Reader(objBytes, nil, nil)
		if err := applier.Applier(); err != nil {
			klog.Warningf("Failed to apply PriorityClass CR %s: %v", pc, err)
			return err
		}
	}
	return nil
}

func ApplyPriorityClasses(pcs []string, kubeconfigPath string) error {
	schedulingScheme := runtime.NewScheme()
	if err := sv1.AddToScheme(schedulingScheme); err != nil {
		return err
	}

	pcApplier := &pcApplier{
		Client: pcClient(kubeconfigPath),
		codecs: serializer.NewCodecFactory(schedulingScheme),
	}
	return applyPriorityClasses(pcs, pcApplier)
}
