package controller

import (
	"context"

	"fmt"
	"net"
	"time"

	"github.com/openshift/route-controller-manager/pkg/route/ingressip"
)

func RunIngressIPController(ctx context.Context, controllerContext *EnhancedControllerContext) (bool, error) {
	// TODO configurable?
	resyncPeriod := 10 * time.Minute

	if len(controllerContext.OpenshiftControllerConfig.Ingress.IngressIPNetworkCIDR) == 0 {
		return true, nil
	}

	_, ipNet, err := net.ParseCIDR(controllerContext.OpenshiftControllerConfig.Ingress.IngressIPNetworkCIDR)
	if err != nil {
		return false, fmt.Errorf("unable to start ingress IP controller: %v", err)
	}

	if ipNet.IP.IsUnspecified() {
		// TODO: Is this an error?
		return true, nil
	}

	ingressIPController := ingressip.NewIngressIPController(
		controllerContext.KubernetesInformers.Core().V1().Services().Informer(),
		controllerContext.ClientBuilder.ClientOrDie(infraServiceIngressIPControllerServiceAccountName),
		ipNet,
		resyncPeriod,
	)
	go ingressIPController.Run(ctx.Done())

	return true, nil
}
