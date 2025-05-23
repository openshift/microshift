package openshift_controller_manager

import (
	"context"
	"k8s.io/klog/v2"

	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/route-controller-manager/pkg/cmd/controller"
)

func RunRouteControllerManager(ctx context.Context, controllerContext *controllercmd.ControllerContext) error {
	config, err := asOpenshiftControllerManagerConfig(controllerContext.ComponentConfig)
	if err != nil {
		return err
	}

	routeControllerManagerContext, err := controller.NewControllerContext(ctx, controllerContext, *config)
	if err != nil {
		klog.Fatal(err)
	}
	if err := startControllers(ctx, routeControllerManagerContext); err != nil {
		klog.Fatal(err)
	}
	routeControllerManagerContext.StartInformers(ctx.Done())

	<-ctx.Done()
	return nil
}

func startControllers(ctx context.Context, controllerContext *controller.EnhancedControllerContext) error {
	for controllerName, initFn := range controller.ControllerManagerInitialization {
		if !controllerContext.IsControllerEnabled(controllerName) {
			klog.Warningf("%q is disabled", controllerName)
			continue
		}
		klog.V(1).Infof("Starting %q", controllerName)
		started, err := initFn(ctx, controllerContext)
		if err != nil {
			klog.Fatalf("Error starting %q (%v)", controllerName, err)
			return err
		}
		if !started {
			klog.Warningf("Skipping %q", controllerName)
			continue
		}
		klog.Infof("Started %q", controllerName)
	}
	klog.Infof("Started Route Controllers")
	return nil
}
