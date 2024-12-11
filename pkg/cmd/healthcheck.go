package cmd

import (
	"context"
	"time"

	"github.com/openshift/microshift/pkg/healthcheck"
	"github.com/spf13/cobra"
)

func NewHealthcheckCommand() *cobra.Command {
	var timeout time.Duration
	var namespaces []string

	cmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "Verify health of the MicroShift",

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(namespaces) != 0 {
				return healthcheck.NamespacesHealthcheck(context.Background(), timeout, namespaces)
			}
			return healthcheck.MicroShiftHealthcheck(context.Background(), timeout)
		},
	}

	cmd.Flags().DurationVar(&timeout, "timeout", 300*time.Second, "The maximum duration of each stage of the healthcheck.")
	cmd.Flags().StringSliceVar(&namespaces, "namespaces", []string{}, "Custom namespaces to wait for workload readiness (does not perform core MicroShift checks).")

	return cmd
}
