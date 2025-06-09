package controller

import (
	"context"

	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	networkingv1 "k8s.io/client-go/kubernetes/typed/networking/v1"

	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"

	"github.com/openshift/route-controller-manager/pkg/route/ingress"
)

func RunIngressToRouteController(ctx context.Context, controllerContext *EnhancedControllerContext) (bool, error) {
	clientConfig := controllerContext.ClientBuilder.ConfigOrDie(InfraIngressToRouteControllerServiceAccountName)
	coreClient, err := coreclient.NewForConfig(clientConfig)
	if err != nil {
		return false, err
	}
	routeClient, err := routeclient.NewForConfig(clientConfig)
	if err != nil {
		return false, err
	}
	networkingClient, err := networkingv1.NewForConfig(clientConfig)
	if err != nil {
		return false, err
	}

	controller := ingress.NewController(
		coreClient,
		routeClient,
		networkingClient,
		controllerContext.KubernetesInformers.Networking().V1().Ingresses(),
		controllerContext.KubernetesInformers.Networking().V1().IngressClasses(),
		controllerContext.KubernetesInformers.Core().V1().Secrets(),
		controllerContext.KubernetesInformers.Core().V1().Services(),
		controllerContext.RouteInformers.Route().V1().Routes(),
	)

	go controller.Run(5, ctx.Done())

	return true, nil
}
