package openshift_controller_manager

import (
	"context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	"github.com/openshift/library-go/pkg/serviceability"

	routecontrollers "github.com/openshift/route-controller-manager/pkg/cmd/controller/route"
	origincontrollers "github.com/openshift/route-controller-manager/pkg/cmd/routecontroller"
	"github.com/openshift/route-controller-manager/pkg/version"
)

func RunRouteControllerManager(config *openshiftcontrolplanev1.OpenShiftControllerManagerConfig, clientConfig *rest.Config, ctx context.Context) error {
	serviceability.InitLogrusFromKlog()
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	// only serve if we have serving information.
	if config.ServingInfo != nil {
		klog.Infof("Starting controllers on %s (%s)", config.ServingInfo.BindAddress, version.Get().String())

		if err := routecontrollers.RunControllerServer(*config.ServingInfo, kubeClient); err != nil {
			return err
		}
	}
	return origincontrollers.RunRouteControllerManager(config, kubeClient, clientConfig, ctx)
}
