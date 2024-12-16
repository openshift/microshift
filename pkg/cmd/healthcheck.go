package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/openshift/microshift/pkg/healthcheck"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewHealthcheckCommand() *cobra.Command {
	var timeout time.Duration
	var custom string

	var namespace string
	var deployments []string
	var daemonsets []string
	var statefulsets []string

	cmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "Verify health of the MicroShift",
		Long: `Verify health of the MicroShift or custom workloads.

Running command without any options performs a healthcheck of MicroShift service and its core workloads.

Checking health of a custom workloads can be achieved in two ways:
1. Use --namespace option together with --deployments, --daemonsets, and --statefulsets
   For example:
      $ microshift healthcheck --namespace openshift-storage --deployments lvms-operator --daemonsets vg-manager
      $ microshift healthcheck --namespace openshift-ovn-kubernetes --daemonsets ovnkube-master,ovnkube-node

2. Provide --custom option with a JSON argument. It can wait for several namespaces in single execution.
   The schema is: '{ "NAMESPACE": {"deployments: ["NAMES"...], "daemonsets: ["NAMES"...], "statefulsets: ["NAMES"...]}}'
   For example:
      $ microshift healthcheck --custom '{"openshift-storage":{"deployments": ["lvms-operator"], "daemonsets": ["vg-manager"]}, "openshift-ovn-kubernetes":{"daemonsets": ["ovnkube-master", "ovnkube-node"]}}'
`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if namespace != "" && custom != "" {
				return fmt.Errorf("only --namespace or --custom can be provided")
			}
			if custom != "" && (len(deployments) != 0 || len(daemonsets) != 0 || len(statefulsets) != 0) {
				klog.Warning("WARNING: --custom was provided, ignoring --deployments, --daemonsets, and --statefulsets")
			}

			if len(custom) != 0 {
				return healthcheck.CustomWorkloadHealthcheck(context.Background(), timeout, custom)
			}

			if namespace != "" {
				return healthcheck.EasyCustomWorkloadHealthcheck(context.Background(), timeout, namespace, deployments, daemonsets, statefulsets)
			}

			return healthcheck.MicroShiftHealthcheck(context.Background(), timeout)
		},
	}

	cmd.Flags().StringVar(&custom, "custom", "", "Custom healthcheck definition.")
	cmd.Flags().DurationVar(&timeout, "timeout", 300*time.Second, "The maximum duration of each stage of the healthcheck.")

	// More user friendly, but only one namespace at the time
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace of the workloads to perform readiness check")
	cmd.Flags().StringSliceVar(&deployments, "deployments", []string{}, "Comma separated list of Deployments to check readiness")
	cmd.Flags().StringSliceVar(&daemonsets, "daemonsets", []string{}, "Comma separated list of DaemonSets to check readiness")
	cmd.Flags().StringSliceVar(&statefulsets, "statefulsets", []string{}, "Comma separated list of StatefulSets to check readiness")

	return cmd
}
