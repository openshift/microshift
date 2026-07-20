package cmd

import (
	"github.com/openshift/microshift/pkg/controllers/c2cc"
	"github.com/spf13/cobra"
)

func NewC2CCProbeCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "c2cc-probe",
		Short:  "Run C2CC remote cluster probe (designed to run as a pod)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c2cc.RunProbe(cmd.Context())
		},
	}
}
