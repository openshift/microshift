package assets

import (
	"context"
	"fmt"

	embedded "github.com/openshift/microshift/assets"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	sccv1 "github.com/openshift/api/security/v1"
	sccclientv1 "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
)

var (
	sccScheme = runtime.NewScheme()
	sccCodecs = serializer.NewCodecFactory(sccScheme)
)

func init() {
	if err := sccv1.AddToScheme(sccScheme); err != nil {
		panic(err)
	}
}

type sccApplier struct {
	Client *sccclientv1.SecurityV1Client
	scc    *sccv1.SecurityContextConstraints
}

func sccClient(kubeconfigPath string) *sccclientv1.SecurityV1Client {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}
	return sccclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "scc-agent"))
}

func (s *sccApplier) Reader(objBytes []byte, render RenderFunc, params RenderParams) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes, params)
		if err != nil {
			panic(err)
		}
	}
	obj, err := runtime.Decode(sccCodecs.UniversalDecoder(sccv1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	s.scc = obj.(*sccv1.SecurityContextConstraints)
}

func (s *sccApplier) Applier() error {
	// adapted from cvo
	existing, err := s.Client.SecurityContextConstraints().Get(context.TODO(), s.scc.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := s.Client.SecurityContextConstraints().Create(context.TODO(), s.scc, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}

	var modified bool
	resourcemerge.EnsureObjectMeta(&modified, &existing.ObjectMeta, s.scc.ObjectMeta)
	if !modified {
		return nil
	}

	_, err = s.Client.SecurityContextConstraints().Update(context.TODO(), existing, metav1.UpdateOptions{})
	return err
}

func applySCCs(sccs []string, applier readerApplier, render RenderFunc, params RenderParams) error {
	lock.Lock()
	defer lock.Unlock()

	for _, scc := range sccs {
		klog.Infof("Applying scc api %s", scc)
		objBytes, err := embedded.Asset(scc)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", scc, err)
		}
		applier.Reader(objBytes, render, params)
		if err := applier.Applier(); err != nil {
			klog.Warningf("Failed to apply scc api %s: %v", scc, err)
			return err
		}
	}
	return nil
}

func ApplySCCs(sccs []string, render RenderFunc, params RenderParams, kubeconfigPath string) error {
	scc := &sccApplier{}
	scc.Client = sccClient(kubeconfigPath)
	return applySCCs(sccs, scc, render, params)
}
