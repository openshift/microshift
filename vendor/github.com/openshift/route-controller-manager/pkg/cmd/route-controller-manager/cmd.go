package openshift_controller_manager

import (
	"context"
	"github.com/spf13/cobra"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/clock"

	"github.com/openshift/library-go/pkg/controller/controllercmd"

	rcmversion "github.com/openshift/route-controller-manager/pkg/version"
)

const (
	podNameEnv      = "POD_NAME"
	podNamespaceEnv = "POD_NAMESPACE"
)

func NewRouteControllerManagerCommand(name string) *cobra.Command {
	cmd := controllercmd.NewControllerCommandConfig("route-controller-manager", rcmversion.Get(), RunRouteControllerManager, clock.RealClock{}).
		WithComponentOwnerReference(&corev1.ObjectReference{
			Kind:      "Pod",
			Name:      os.Getenv(podNameEnv),
			Namespace: getNamespace(),
		}).
		NewCommandWithContext(context.Background())
	cmd.Use = name
	cmd.Short = "Start the additional Ingress and Route controllers"

	return cmd
}

// getNamespace returns in-cluster namespace
func getNamespace() string {
	if nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		return string(nsBytes)
	}
	if podNamespace := os.Getenv(podNamespaceEnv); len(podNamespace) > 0 {
		return podNamespace
	}
	return "openshift-route-controller-manager"
}
