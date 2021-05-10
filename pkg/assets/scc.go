package assets

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	sccassets "github.com/openshift/microshift/pkg/assets/scc"
	"github.com/openshift/microshift/pkg/constant"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	sccv1 "github.com/openshift/api/security/v1"
	sccclientv1 "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
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

func sccClient() *sccclientv1.SecurityV1Client {
	restConfig, err := clientcmd.BuildConfigFromFlags("", constant.AdminKubeconfigPath)
	if err != nil {
		panic(err)
	}
	return sccclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "scc-agent"))
}

func (s *sccApplier) Reader(objBytes []byte, render RenderFunc) {
	var err error
	if render != nil {
		objBytes, err = render(objBytes)
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
	_, err := s.Client.SecurityContextConstraints().Get(context.TODO(), s.scc.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := s.Client.SecurityContextConstraints().Create(context.TODO(), s.scc, metav1.CreateOptions{})
		return err
	}
	return nil
}

func applySCCs(sccs []string, applier readerApplier, render RenderFunc) error {
	lock.Lock()
	defer lock.Unlock()

	for _, scc := range sccs {
		logrus.Infof("applying scc api %s", scc)
		objBytes, err := sccassets.Asset(scc)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", scc, err)
		}
		applier.Reader(objBytes, render)
		if err := applier.Applier(); err != nil {
			logrus.Warningf("failed to apply scc api %s: %v", scc, err)
			return err
		}
	}
	return nil
}

func ApplySCCs(sccs []string, render RenderFunc) error {
	scc := &sccApplier{}
	scc.Client = sccClient()
	return applySCCs(sccs, scc, render)
}
