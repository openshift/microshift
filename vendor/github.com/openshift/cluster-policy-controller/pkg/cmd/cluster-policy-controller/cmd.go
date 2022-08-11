package cluster_policy_controller

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/library-go/pkg/controller/controllercmd"

	clusterpolicyversion "github.com/openshift/cluster-policy-controller/pkg/version"
)

const (
	podNameEnv      = "POD_NAME"
	podNamespaceEnv = "POD_NAMESPACE"
)

func NewClusterPolicyControllerCommand(name string) *cobra.Command {
	cmd := controllercmd.NewControllerCommandConfig("cluster-policy-controller", clusterpolicyversion.Get(), RunClusterPolicyController).
		WithComponentOwnerReference(&corev1.ObjectReference{
			Kind:      "Pod",
			Name:      os.Getenv(podNameEnv),
			Namespace: os.Getenv(podNamespaceEnv),
		}).
		NewCommandWithContext(context.Background())
	cmd.Use = name
	cmd.Short = "Start the cluster policy controller"

	return cmd
}
