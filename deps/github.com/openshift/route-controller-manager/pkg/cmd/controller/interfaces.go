package controller

import (
	"context"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/controller-manager/app"
	"k8s.io/controller-manager/pkg/clientbuilder"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned"
	routeinformer "github.com/openshift/client-go/route/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
)

type ControllerClientBuilder interface {
	clientbuilder.ControllerClientBuilder
}

type EnhancedControllerContext struct {
	*controllercmd.ControllerContext
	OpenshiftControllerConfig openshiftcontrolplanev1.OpenShiftControllerManagerConfig

	// ClientBuilder will provide a client for this controller to use
	ClientBuilder ControllerClientBuilder

	KubernetesInformers informers.SharedInformerFactory
	RouteInformers      routeinformer.SharedInformerFactory

	informersStartedLock   sync.Mutex
	informersStartedClosed bool
	// InformersStarted is closed after all of the controllers have been initialized and are running.  After this point it is safe,
	// for an individual controller to start the shared informers. Before it is closed, they should not.
	InformersStarted chan struct{}
}

type RouteControllerClientBuilder struct {
	clientbuilder.ControllerClientBuilder
}

func (c *EnhancedControllerContext) IsControllerEnabled(name string) bool {
	return app.IsControllerEnabled(name, sets.String{}, c.OpenshiftControllerConfig.Controllers)
}

func (c *EnhancedControllerContext) StartInformers(stopCh <-chan struct{}) {
	c.KubernetesInformers.Start(stopCh)

	c.RouteInformers.Start(stopCh)

	c.informersStartedLock.Lock()
	defer c.informersStartedLock.Unlock()
	if !c.informersStartedClosed {
		close(c.InformersStarted)
		c.informersStartedClosed = true
	}
}

// InitFunc is used to launch a particular controller.  It may run additional "should I activate checks".
// Any error returned will cause the controller process to `Fatal`
// The bool indicates whether the controller was enabled.
type InitFunc func(ctx context.Context, controllerCtx *EnhancedControllerContext) (bool, error)

func NewControllerContext(ctx context.Context, controllerContext *controllercmd.ControllerContext, config openshiftcontrolplanev1.OpenShiftControllerManagerConfig) (*EnhancedControllerContext, error) {
	inClientConfig := controllerContext.KubeConfig

	const defaultInformerResyncPeriod = 10 * time.Minute
	kubeClient, err := kubernetes.NewForConfig(inClientConfig)
	if err != nil {
		return nil, err
	}

	// copy to avoid messing with original
	clientConfig := rest.CopyConfig(inClientConfig)
	// divide up the QPS since it re-used separately for every client
	numOfControllers := len(ControllerManagerInitialization)
	if clientConfig.QPS > 0 {
		clientConfig.QPS = clientConfig.QPS/float32(numOfControllers) + 1
	}
	if clientConfig.Burst > 0 {
		clientConfig.Burst = clientConfig.Burst/numOfControllers + 1
	}

	routerClient, err := routeclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	routeControllerContext := &EnhancedControllerContext{
		OpenshiftControllerConfig: config,

		ClientBuilder: RouteControllerClientBuilder{
			ControllerClientBuilder: clientbuilder.NewDynamicClientBuilder(
				rest.AnonymousClientConfig(clientConfig),
				kubeClient.CoreV1(),
				defaultOpenShiftInfraNamespace),
		},
		KubernetesInformers: informers.NewSharedInformerFactory(kubeClient, defaultInformerResyncPeriod),
		RouteInformers:      routeinformer.NewSharedInformerFactory(routerClient, defaultInformerResyncPeriod),
		InformersStarted:    make(chan struct{}),
	}

	return routeControllerContext, nil
}
