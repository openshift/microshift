package route

import (
	"context"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/controller-manager/app"
	"k8s.io/controller-manager/pkg/clientbuilder"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	operatorclient "github.com/openshift/client-go/operator/clientset/versioned"
	operatorinformer "github.com/openshift/client-go/operator/informers/externalversions"
	routeclient "github.com/openshift/client-go/route/clientset/versioned"
	routeinformer "github.com/openshift/client-go/route/informers/externalversions"
	"github.com/openshift/openshift-controller-manager/pkg/client/genericinformers"
)

type ControllerClientBuilder interface {
	clientbuilder.ControllerClientBuilder

	OpenshiftConfigClient(name string) (configclient.Interface, error)
	OpenshiftConfigClientOrDie(name string) configclient.Interface

	OpenshiftOperatorClient(name string) (operatorclient.Interface, error)
	OpenshiftOperatorClientOrDie(name string) operatorclient.Interface
}

type ControllerContext struct {
	// TODO: Make this minimal config instead of passing entire controller manager config
	OpenshiftControllerConfig openshiftcontrolplanev1.OpenShiftControllerManagerConfig

	// ClientBuilder will provide a client for this controller to use
	ClientBuilder ControllerClientBuilder
	// HighRateLimitClientBuilder will provide a client for this controller utilizing a higher rate limit.
	// This will have a rate limit of at least 100 QPS, with a burst up to 200 QPS.
	HighRateLimitClientBuilder ControllerClientBuilder

	KubernetesInformers                informers.SharedInformerFactory
	OpenshiftConfigKubernetesInformers informers.SharedInformerFactory

	RouteInformers    routeinformer.SharedInformerFactory
	OperatorInformers operatorinformer.SharedInformerFactory

	GenericResourceInformer genericinformers.GenericResourceInformer
	RestMapper              meta.RESTMapper

	// Stop is the stop channel
	Stop    <-chan struct{}
	Context context.Context

	informersStartedLock   sync.Mutex
	informersStartedClosed bool
	// InformersStarted is closed after all of the controllers have been initialized and are running.  After this point it is safe,
	// for an individual controller to start the shared informers. Before it is closed, they should not.
	InformersStarted chan struct{}
}

type RouteControllerClientBuilder struct {
	clientbuilder.ControllerClientBuilder
}

// OpenshiftConfigClient provides a REST client for the build  API.
// If the client cannot be created because of configuration error, this function
// will error.
func (b RouteControllerClientBuilder) OpenshiftConfigClient(name string) (configclient.Interface, error) {
	clientConfig, err := b.Config(name)
	if err != nil {
		return nil, err
	}
	return configclient.NewForConfig(nonProtobufConfig(clientConfig))
}

// RouteControllerClientBuilder provides a REST client for the build API.
// If the client cannot be created because of configuration error, this function
// will panic.
func (b RouteControllerClientBuilder) OpenshiftConfigClientOrDie(name string) configclient.Interface {
	client, err := b.OpenshiftConfigClient(name)
	if err != nil {
		klog.Fatal(err)
	}
	return client
}

func (b RouteControllerClientBuilder) OpenshiftOperatorClient(name string) (operatorclient.Interface, error) {
	clientConfig, err := b.Config(name)
	if err != nil {
		return nil, err
	}
	return operatorclient.NewForConfig(clientConfig)
}

func (b RouteControllerClientBuilder) OpenshiftOperatorClientOrDie(name string) operatorclient.Interface {
	client, err := b.OpenshiftOperatorClient(name)
	if err != nil {
		klog.Fatal(err)
	}
	return client
}

func (c *ControllerContext) IsControllerEnabled(name string) bool {
	return app.IsControllerEnabled(name, sets.String{}, c.OpenshiftControllerConfig.Controllers)
}

func (c *ControllerContext) StartInformers(stopCh <-chan struct{}) {
	c.KubernetesInformers.Start(stopCh)
	c.OpenshiftConfigKubernetesInformers.Start(stopCh)

	c.RouteInformers.Start(stopCh)
	c.OperatorInformers.Start(stopCh)

	c.informersStartedLock.Lock()
	defer c.informersStartedLock.Unlock()
	if !c.informersStartedClosed {
		close(c.InformersStarted)
		c.informersStartedClosed = true
	}
}

func (c *ControllerContext) ToGenericInformer() genericinformers.GenericResourceInformer {
	return genericinformers.NewGenericInformers(
		c.StartInformers,
		c.KubernetesInformers,
		genericinformers.GenericResourceInformerFunc(func(resource schema.GroupVersionResource) (informers.GenericInformer, error) {
			return c.RouteInformers.ForResource(resource)
		}),
	)
}

// InitFunc is used to launch a particular controller.  It may run additional "should I activate checks".
// Any error returned will cause the controller process to `Fatal`
// The bool indicates whether the controller was enabled.
type InitFunc func(ctx *ControllerContext) (bool, error)

func NewControllerContext(
	ctx context.Context,
	config openshiftcontrolplanev1.OpenShiftControllerManagerConfig,
	inClientConfig *rest.Config,
) (*ControllerContext, error) {

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

	discoveryClient := cacheddiscovery.NewMemCacheClient(kubeClient.Discovery())
	dynamicRestMapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	dynamicRestMapper.Reset()
	go wait.Until(dynamicRestMapper.Reset, 30*time.Second, ctx.Done())

	routerClient, err := routeclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	operatorClient, err := operatorclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	// Create a new clientConfig for high rate limit workloads.
	// Increase kube QPS to at least 100 QPS, burst to at least 200 QPS.
	highRateLimitClientConfig := rest.CopyConfig(inClientConfig)
	if highRateLimitClientConfig.QPS < 100 {
		highRateLimitClientConfig.QPS = 100
	}
	if highRateLimitClientConfig.Burst < 200 {
		highRateLimitClientConfig.Burst = 200
	}

	routeControllerContext := &ControllerContext{
		OpenshiftControllerConfig: config,

		// k8s 1.21 rebase - SAControllerClientBuilder replaced with NewDynamicClientBuilder
		// See https://github.com/kubernetes/kubernetes/pull/99291
		ClientBuilder: RouteControllerClientBuilder{
			ControllerClientBuilder: clientbuilder.NewDynamicClientBuilder(
				rest.AnonymousClientConfig(clientConfig),
				kubeClient.CoreV1(),
				defaultOpenShiftInfraNamespace),
		},
		HighRateLimitClientBuilder: RouteControllerClientBuilder{
			ControllerClientBuilder: clientbuilder.NewDynamicClientBuilder(
				rest.AnonymousClientConfig(highRateLimitClientConfig),
				kubeClient.CoreV1(),
				defaultOpenShiftInfraNamespace),
		},
		KubernetesInformers:                informers.NewSharedInformerFactory(kubeClient, defaultInformerResyncPeriod),
		OpenshiftConfigKubernetesInformers: informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultInformerResyncPeriod, informers.WithNamespace("openshift-config")),
		OperatorInformers:                  operatorinformer.NewSharedInformerFactory(operatorClient, defaultInformerResyncPeriod),
		RouteInformers:                     routeinformer.NewSharedInformerFactory(routerClient, defaultInformerResyncPeriod),
		Stop:                               ctx.Done(),
		Context:                            ctx,
		InformersStarted:                   make(chan struct{}),
		RestMapper:                         dynamicRestMapper,
	}
	routeControllerContext.GenericResourceInformer = routeControllerContext.ToGenericInformer()

	return routeControllerContext, nil
}

// nonProtobufConfig returns a copy of inConfig that doesn't force the use of protobufs,
// for working with CRD-based APIs.
func nonProtobufConfig(inConfig *rest.Config) *rest.Config {
	npConfig := rest.CopyConfig(inConfig)
	npConfig.ContentConfig.AcceptContentTypes = "application/json"
	npConfig.ContentConfig.ContentType = "application/json"
	return npConfig
}
