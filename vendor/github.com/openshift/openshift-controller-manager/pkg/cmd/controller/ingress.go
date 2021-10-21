package controller

import (
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	networkingv1 "k8s.io/client-go/kubernetes/typed/networking/v1"

	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"github.com/openshift/openshift-controller-manager/pkg/route/ingress"
)

func RunIngressToRouteController(ctx *ControllerContext) (bool, error) {
	clientConfig := ctx.ClientBuilder.ConfigOrDie(infraIngressToRouteControllerServiceAccountName)
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
		ctx.KubernetesInformers.Networking().V1().Ingresses(),
		ctx.KubernetesInformers.Networking().V1().IngressClasses(),
		ctx.KubernetesInformers.Core().V1().Secrets(),
		ctx.KubernetesInformers.Core().V1().Services(),
		ctx.RouteInformers.Route().V1().Routes(),
	)

	go controller.Run(5, ctx.Stop)

	return true, nil
}
