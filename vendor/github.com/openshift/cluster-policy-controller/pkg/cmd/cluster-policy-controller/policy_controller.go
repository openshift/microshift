package cluster_policy_controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	origincontrollers "github.com/openshift/cluster-policy-controller/pkg/cmd/controller"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
)

func RunClusterPolicyController(ctx context.Context, controllerContext *controllercmd.ControllerContext) error {
	config, err := asOpenshiftControllerManagerConfig(controllerContext.ComponentConfig)
	if err != nil {
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(controllerContext.KubeConfig)
	if err != nil {
		return err
	}

	if err := WaitForHealthyAPIServer(kubeClient.Discovery().RESTClient()); err != nil {
		klog.Fatal(err)
	}

	openshiftControllerManagerContext, err := origincontrollers.NewControllerContext(ctx, controllerContext, *config)
	if err != nil {
		klog.Fatal(err)
	}
	if err := startControllers(ctx, openshiftControllerManagerContext); err != nil {
		klog.Fatal(err)
	}
	openshiftControllerManagerContext.StartInformers(ctx.Done())

	<-ctx.Done()
	return nil
}

func WaitForHealthyAPIServer(client rest.Interface) error {
	var healthzContent string
	// If apiserver is not running we should wait for some time and fail only then. This is particularly
	// important when we start apiserver and controller manager at the same time.
	err := wait.PollImmediate(time.Second, 5*time.Minute, func() (bool, error) {
		healthStatus := 0
		resp := client.Get().AbsPath("/healthz").Do(context.TODO()).StatusCode(&healthStatus)
		if healthStatus != http.StatusOK {
			klog.Errorf("Server isn't healthy yet. Waiting a little while.")
			return false, nil
		}
		content, _ := resp.Raw()
		healthzContent = string(content)

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("server unhealthy: %v: %v", healthzContent, err)
	}

	return nil
}

// startControllers launches the controllers
// allocation controller is passed in because it wants direct etcd access.  Naughty.
func startControllers(ctx context.Context, controllerCtx *origincontrollers.EnhancedControllerContext) error {
	for controllerName, initFn := range origincontrollers.ControllerInitializers {
		if !controllerCtx.IsControllerEnabled(controllerName) {
			klog.Warningf("%q is disabled", controllerName)
			continue
		}

		klog.V(1).Infof("Starting %q", controllerName)
		started, err := initFn(ctx, controllerCtx)
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

	klog.Infof("Started Origin Controllers")

	return nil
}
