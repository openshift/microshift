package cmd

import (
	"github.com/openshift/microshift/pkg/ostree"
	"github.com/spf13/cobra"
)

func NewOstreeCommand() *cobra.Command {
	preRun := &cobra.Command{
		Use:   "pre-run",
		Short: "Execute ostree-specific pre-run procedures",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ostree.PreRun()
		},
	}
	scheduleBackup := &cobra.Command{
		Use:   "schedule-backup",
		Short: "Schedule backup for next boot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ostree.ScheduleBackup()
		},
	}

	cmd := &cobra.Command{
		Use:   "ostree",
		Short: "Functionality for ostree-based systems",
	}

	cmd.AddCommand(scheduleBackup)
	cmd.AddCommand(preRun)

	return cmd
}
