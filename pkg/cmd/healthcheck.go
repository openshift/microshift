package cmd

import (
	"context"
	"time"

	"github.com/openshift/microshift/pkg/healthcheck"
	"github.com/spf13/cobra"
)

func NewHealthcheckCommand() *cobra.Command {
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "Verify health of the MicroShift",

		RunE: func(cmd *cobra.Command, args []string) error {
			return healthcheck.MicroShiftHealthcheck(context.Background(), timeout)
		},
	}

	cmd.Flags().DurationVar(&timeout, "timeout", 300*time.Second, "The maximum duration of each stage of the healthcheck.")

	return cmd
}
