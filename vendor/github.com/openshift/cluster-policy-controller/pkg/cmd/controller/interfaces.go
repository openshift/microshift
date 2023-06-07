package controller

import (
	"context"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/rest"
	"k8s.io/controller-manager/app"
	"k8s.io/controller-manager/pkg/clientbuilder"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	imageclient "github.com/openshift/client-go/image/clientset/versioned"
	imageinformer "github.com/openshift/client-go/image/informers/externalversions"
	quotaclient "github.com/openshift/client-go/quota/clientset/versioned"
	quotainformer "github.com/openshift/client-go/quota/informers/externalversions"
	securityclient "github.com/openshift/client-go/security/clientset/versioned"
	securityinformer "github.com/openshift/client-go/security/informers/externalversions"
	securityinternalclient "github.com/openshift/client-go/securityinternal/clientset/versioned"
	"github.com/openshift/library-go/pkg/controller/controllercmd"

	"github.com/openshift/cluster-policy-controller/pkg/client/genericinformers"
)

func NewControllerContext(
	ctx context.Context,
	controllerContext *controllercmd.ControllerContext,
	config openshiftcontrolplanev1.OpenShiftControllerManagerConfig,
) (*EnhancedControllerContext, error) {
	inClientConfig := controllerContext.KubeConfig

	const defaultInformerResyncPeriod = 10 * time.Minute
	kubeClient, err := kubernetes.NewForConfig(inClientConfig)
	if err != nil {
		return nil, err
	}

	// copy to avoid messing with original
	clientConfig := rest.CopyConfig(inClientConfig)
	// divide up the QPS since it re-used separately for every client
	// TODO, eventually make this configurable individually in some way.
	if clientConfig.QPS > 0 {
		clientConfig.QPS = clientConfig.QPS/10 + 1
	}
	if clientConfig.Burst > 0 {
		clientConfig.Burst = clientConfig.Burst/10 + 1
	}

	metadataClient, err := metadata.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	imageClient, err := imageclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	quotaClient, err := quotaclient.NewForConfig(nonProtobufConfig(clientConfig))
	if err != nil {
		return nil, err
	}
	securityClient, err := securityclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	openshiftControllerContext := &EnhancedControllerContext{
		ControllerContext:         controllerContext,
		OpenshiftControllerConfig: config,

		ClientBuilder: OpenshiftControllerClientBuilder{
			ControllerClientBuilder: clientbuilder.NewDynamicClientBuilder(rest.AnonymousClientConfig(clientConfig), kubeClient.CoreV1(), defaultOpenShiftInfraNamespace),
		},
		KubernetesInformers: informers.NewSharedInformerFactory(kubeClient, defaultInformerResyncPeriod),
		MetadataInformers:   metadatainformer.NewSharedInformerFactory(metadataClient, defaultInformerResyncPeriod),
		ImageInformers:      imageinformer.NewSharedInformerFactory(imageClient, defaultInformerResyncPeriod),
		QuotaInformers:      quotainformer.NewSharedInformerFactory(quotaClient, defaultInformerResyncPeriod),
		SecurityInformers:   securityinformer.NewSharedInformerFactory(securityClient, defaultInformerResyncPeriod),
		InformersStarted:    make(chan struct{}),
	}
	openshiftControllerContext.GenericResourceInformer = openshiftControllerContext.ToGenericInformer()

	return openshiftControllerContext, nil
}

func (c *EnhancedControllerContext) ToGenericInformer() genericinformers.GenericResourceInformer {
	return genericinformers.NewGenericInformers(
		c.StartInformers,
		// first shared informers used by the controllers
		c.KubernetesInformers,
		genericinformers.GenericResourceInformerFunc(func(resource schema.GroupVersionResource) (informers.GenericInformer, error) {
			return c.ImageInformers.ForResource(resource)
		}),
		genericinformers.GenericResourceInformerFunc(func(resource schema.GroupVersionResource) (informers.GenericInformer, error) {
			return c.QuotaInformers.ForResource(resource)
		}),
		// fallback to metadata shared informers
		genericinformers.GenericResourceInformerFunc(func(resource schema.GroupVersionResource) (informers.GenericInformer, error) {
			return c.MetadataInformers.ForResource(resource), nil
		}),
	)
}

type EnhancedControllerContext struct {
	*controllercmd.ControllerContext
	OpenshiftControllerConfig openshiftcontrolplanev1.OpenShiftControllerManagerConfig

	// ClientBuilder will provide a client for this controller to use
	ClientBuilder ControllerClientBuilder

	KubernetesInformers informers.SharedInformerFactory
	MetadataInformers   metadatainformer.SharedInformerFactory

	QuotaInformers    quotainformer.SharedInformerFactory
	ImageInformers    imageinformer.SharedInformerFactory
	SecurityInformers securityinformer.SharedInformerFactory

	GenericResourceInformer genericinformers.GenericResourceInformer

	informersStartedLock   sync.Mutex
	informersStartedClosed bool
	// InformersStarted is closed after all of the controllers have been initialized and are running.  After this point it is safe,
	// for an individual controller to start the shared informers. Before it is closed, they should not.
	InformersStarted chan struct{}
}

func (c *EnhancedControllerContext) StartInformers(stopCh <-chan struct{}) {
	c.KubernetesInformers.Start(stopCh)

	c.ImageInformers.Start(stopCh)
	c.SecurityInformers.Start(stopCh)
	c.QuotaInformers.Start(stopCh)

	c.MetadataInformers.Start(stopCh)

	c.informersStartedLock.Lock()
	defer c.informersStartedLock.Unlock()
	if !c.informersStartedClosed {
		close(c.InformersStarted)
		c.informersStartedClosed = true
	}
}

func (c *EnhancedControllerContext) IsControllerEnabled(name string) bool {
	return app.IsControllerEnabled(name, sets.String{}, c.OpenshiftControllerConfig.Controllers)
}

type ControllerClientBuilder interface {
	clientbuilder.ControllerClientBuilder

	OpenshiftSecurityClient(name string) (securityinternalclient.Interface, error)
	OpenshiftSecurityClientOrDie(name string) securityinternalclient.Interface

	OpenshiftImageClient(name string) (imageclient.Interface, error)
	OpenshiftImageClientOrDie(name string) imageclient.Interface

	OpenshiftQuotaClient(name string) (quotaclient.Interface, error)
	OpenshiftQuotaClientOrDie(name string) quotaclient.Interface
}

// InitFunc is used to launch a particular controller.  It may run additional "should I activate checks".
// Any error returned will cause the controller process to `Fatal`
// The bool indicates whether the controller was enabled.
type InitFunc func(ctx context.Context, controllerCtx *EnhancedControllerContext) (bool, error)

type OpenshiftControllerClientBuilder struct {
	clientbuilder.ControllerClientBuilder
}

// OpenshiftImageClient provides a REST client for the image API.
// If the client cannot be created because of configuration error, this function
// will error.
func (b OpenshiftControllerClientBuilder) OpenshiftImageClient(name string) (imageclient.Interface, error) {
	clientConfig, err := b.Config(name)
	if err != nil {
		return nil, err
	}
	return imageclient.NewForConfig(clientConfig)
}

// OpenshiftImageClientOrDie provides a REST client for the image API.
// If the client cannot be created because of configuration error, this function
// will panic.
func (b OpenshiftControllerClientBuilder) OpenshiftImageClientOrDie(name string) imageclient.Interface {
	client, err := b.OpenshiftImageClient(name)
	if err != nil {
		klog.Fatal(err)
	}
	return client
}

func (b OpenshiftControllerClientBuilder) OpenshiftQuotaClient(name string) (quotaclient.Interface, error) {
	clientConfig, err := b.Config(name)
	if err != nil {
		return nil, err
	}
	return quotaclient.NewForConfig(nonProtobufConfig(clientConfig))
}

// OpenshiftInternalBuildClientOrDie provides a REST client for the build API.
// If the client cannot be created because of configuration error, this function
// will panic.
func (b OpenshiftControllerClientBuilder) OpenshiftQuotaClientOrDie(name string) quotaclient.Interface {
	client, err := b.OpenshiftQuotaClient(name)
	if err != nil {
		klog.Fatal(err)
	}
	return client
}

func (b OpenshiftControllerClientBuilder) OpenshiftSecurityClient(name string) (securityinternalclient.Interface, error) {
	clientConfig, err := b.Config(name)
	if err != nil {
		return nil, err
	}
	return securityinternalclient.NewForConfig(nonProtobufConfig(clientConfig))
}

func (b OpenshiftControllerClientBuilder) OpenshiftSecurityClientOrDie(name string) securityinternalclient.Interface {
	client, err := b.OpenshiftSecurityClient(name)
	if err != nil {
		klog.Fatal(err)
	}
	return client
}

// nonProtobufConfig returns a copy of inConfig that doesn't force the use of protobufs,
// for working with CRD-based APIs.
func nonProtobufConfig(inConfig *rest.Config) *rest.Config {
	npConfig := rest.CopyConfig(inConfig)
	npConfig.ContentConfig.AcceptContentTypes = "application/json"
	npConfig.ContentConfig.ContentType = "application/json"
	return npConfig
}
