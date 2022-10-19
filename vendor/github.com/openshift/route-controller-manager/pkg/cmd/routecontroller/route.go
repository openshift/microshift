package routecontroller

import (
	"context"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"

	routecontrollers "github.com/openshift/route-controller-manager/pkg/cmd/controller/route"
)

func RunRouteControllerManager(config *openshiftcontrolplanev1.OpenShiftControllerManagerConfig, kubeClient kubernetes.Interface, clientConfig *rest.Config, ctx context.Context) error {
	routeControllerManager := func(cntx context.Context) {
		// Start Route Controllers
		// TODO: This can be split further
		routeControllerContext, err := routecontrollers.NewControllerContext(cntx, *config, clientConfig)
		if err != nil {
			klog.Fatal(err)
		}
		if err := startControllers(routeControllerContext); err != nil {
			klog.Fatal(err)
		}
		routeControllerContext.StartInformers(cntx.Done())
	}
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	eventRecorder := eventBroadcaster.NewRecorder(legacyscheme.Scheme, v1.EventSource{Component: "route-controller-manager"})
	id, err := os.Hostname()
	if err != nil {
		return err
	}
	// Create a new lease for the route controller manager
	rl, err := resourcelock.New(
		"leases",
		"openshift-route-controller-manager", // TODO: This namespace needs to be created by ocm for now.
		"openshift-route-controllers",
		kubeClient.CoreV1(),
		kubeClient.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: eventRecorder,
		})
	if err != nil {
		return err
	}
	leaderelection.RunOrDie(ctx,
		leaderelection.LeaderElectionConfig{
			Lock:            rl,
			ReleaseOnCancel: true,
			LeaseDuration:   config.LeaderElection.LeaseDuration.Duration,
			RenewDeadline:   config.LeaderElection.RenewDeadline.Duration,
			RetryPeriod:     config.LeaderElection.RetryPeriod.Duration,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: routeControllerManager,
				OnStoppedLeading: func() {
					defer os.Exit(0)
					klog.Warningf("Route Controller Manager received stop signal: leaderelection lost")
				},
			},
		})
	return nil
}

func startControllers(controllerContext *routecontrollers.ControllerContext) error {
	for controllerName, initFn := range routecontrollers.ControllerManagerInitialization {
		if !controllerContext.IsControllerEnabled(controllerName) {
			klog.Warningf("%q is disabled", controllerName)
			continue
		}
		klog.V(1).Infof("Starting %q", controllerName)
		started, err := initFn(controllerContext)
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
