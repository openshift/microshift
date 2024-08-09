package cmd

import (
	"context"

	"github.com/openshift/microshift/pkg/healthcheck"
	"github.com/spf13/cobra"
)

func NewHealthcheckCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "Verify health of the MicroShift",

		RunE: func(cmd *cobra.Command, args []string) error {
			return healthcheck.MicroShiftHealthcheck(context.Background())
		},
	}

	return cmd
}
