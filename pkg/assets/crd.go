package assets

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	crd_assets "github.com/openshift/microshift/pkg/assets/crd"
	"github.com/openshift/microshift/pkg/config"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiext_clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	customResourceReadyInterval = 5 * time.Second
	customResourceReadyTimeout  = 10 * time.Minute
)

var (
	apiExtensionsScheme = apiruntime.NewScheme()
	apiExtensionsCodecs = serializer.NewCodecFactory(apiExtensionsScheme)
	crds                = []string{
		"assets/crd/0000_03_security-openshift_01_scc.crd.yaml",
		"assets/crd/0000_11_imageregistry-configs.crd.yaml",
		"assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml",
		"assets/crd/0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml",
		"assets/crd/0000_10_config-operator_01_build.crd.yaml",
		"assets/crd/0000_10_config-operator_01_image.crd.yaml",
		"assets/crd/0000_03_config-operator_01_proxy.crd.yaml",
		"assets/crd/0000_03_quota-openshift_01_clusterresourcequota.crd.yaml",
		/*
			"assets/crd/0000_03_quota-openshift_01_clusterresourcequota.crd.yaml",
			"assets/crd/0000_10_config-operator_01_apiserver.crd.yaml",
			"assets/crd/0000_10_config-operator_01_authentication.crd.yaml",

			"assets/crd/0000_10_config-operator_01_console.crd.yaml",
			"assets/crd/0000_10_config-operator_01_dns.crd.yaml",
			"assets/crd/0000_10_config-operator_01_featuregate.crd.yaml",


			"assets/crd/0000_10_config-operator_01_infrastructure.crd.yaml",
			"assets/crd/0000_10_config-operator_01_ingress.crd.yaml",
			"assets/crd/0000_10_config-operator_01_network.crd.yaml",
			"assets/crd/0000_10_config-operator_01_oauth.crd.yaml",
			"assets/crd/0000_10_config-operator_01_project.crd.yaml",
			"assets/crd/0000_03_config-operator_01_operatorhub.crd.yaml",


			"assets/crd/0000_00_cluster-version-operator_01_clusteroperator.crd.yaml",
			"assets/crd/cluster-ingress-00-custom-resource-definition.yaml",
			"assets/crd/0000_70_dns-operator_00-custom-resource-definition.yaml",
			"assets/crd/0000_50_service-ca-operator_02_crd.yaml",
			"assets/crd/0000_00_cluster-version-operator_01_clusterversion.crd.yaml",
		*/
	}
)

func getCRD(client apiextclientv1beta1.CustomResourceDefinitionsGetter, resource *apiextv1beta1.CustomResourceDefinition) error {
	_, err := client.CustomResourceDefinitions().Get(context.TODO(), resource.Name, metav1.GetOptions{})
	if err != nil {
		lastErr := fmt.Errorf("error getting CustomResourceDefinition %s: %v", resource.Name, err)
		logrus.Infof("getting openshift CRD status %v", lastErr)
		return lastErr
	}
	return nil
}

func waitForCRD(client apiextclientv1beta1.CustomResourceDefinitionsGetter, resource *apiextv1beta1.CustomResourceDefinition) error {
	var lastErr error
	if err := wait.Poll(customResourceReadyInterval, customResourceReadyTimeout, func() (bool, error) {
		crd, err := client.CustomResourceDefinitions().Get(context.TODO(), resource.Name, metav1.GetOptions{})
		if err != nil {
			lastErr = fmt.Errorf("error getting CustomResourceDefinition %s: %v", resource.Name, err)
			logrus.Infof("getting openshift CRD status %v", lastErr)
			return false, nil
		}

		for _, condition := range crd.Status.Conditions {
			if condition.Type == apiextv1beta1.Established && condition.Status == apiextv1beta1.ConditionTrue {
				return true, nil
			}
		}
		lastErr = fmt.Errorf("CustomResourceDefinition %s is not ready. conditions: %v", crd.Name, crd.Status.Conditions)
		logrus.Infof("getting openshift CRD status %v", lastErr)
		return false, nil
	}); err != nil {
		if err == wait.ErrWaitTimeout {
			return fmt.Errorf("%v during syncCustomResourceDefinitions: %v", err, lastErr)
		}
		return err
	}
	return nil
}
func readCRDOrDie(objBytes []byte) *apiextv1beta1.CustomResourceDefinition {
	requiredObj, err := apiruntime.Decode(apiExtensionsCodecs.UniversalDecoder(apiextv1beta1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	return requiredObj.(*apiextv1beta1.CustomResourceDefinition)
}

func applyCRD(client apiextclientv1beta1.CustomResourceDefinitionsGetter, obj *apiextv1beta1.CustomResourceDefinition) error {
	_, err := client.CustomResourceDefinitions().Get(context.TODO(), obj.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := client.CustomResourceDefinitions().Create(context.TODO(), obj, metav1.CreateOptions{})
		return err
	}
	return nil
}

func init() {
	if err := apiextv1beta1.AddToScheme(apiExtensionsScheme); err != nil {
		panic(err)
	}
}
func ApplyCRDs(cfg *config.MicroshiftConfig) error {
	lock.Lock()
	defer lock.Unlock()

	restConfig, err := clientcmd.BuildConfigFromFlags("", cfg.DataDir+"/resources/kubeadmin/kubeconfig")
	if err != nil {
		return err
	}

	apiExtClient := apiext_clientset.NewForConfigOrDie(rest.AddUserAgent(restConfig, "crd-agent"))

	for _, crd := range crds {
		logrus.Infof("applying openshift CRD %s", crd)
		crdBytes, err := crd_assets.Asset(crd)
		if err != nil {
			return fmt.Errorf("error getting asset %s: %v", crd, err)
		}
		c := readCRDOrDie(crdBytes)
		if err := wait.Poll(customResourceReadyInterval, customResourceReadyTimeout, func() (bool, error) {
			if err := applyCRD(apiExtClient.ApiextensionsV1beta1(), c); err != nil {
				logrus.Warningf("failed to apply openshift CRD %s: %v", crd, err)
				return false, nil
			}
			logrus.Infof("waiting openshift CRD %s", crd)
			if err := getCRD(apiExtClient.ApiextensionsV1beta1(), c); err != nil {
				logrus.Warningf("failed to wait for openshift CRD %s: %v", crd, err)
				return false, nil
			}
			return true, nil
		}); err != nil {
			if err == wait.ErrWaitTimeout {
				return fmt.Errorf("%v during syncCustomResourceDefinitions", err)
			}
			return err
		}
	}

	return nil
}
