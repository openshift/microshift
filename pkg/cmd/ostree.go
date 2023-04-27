package cmd

import (
	"fmt"

	"github.com/openshift/microshift/pkg/ostree"
	"github.com/openshift/microshift/pkg/util"
	"github.com/spf13/cobra"
)

func failIfNotOstree(*cobra.Command, []string) error {
	ostreeBootedFile := "/run/ostree-booted"
	exists, err := util.CheckIfFileExists(ostreeBootedFile)
	if exists {
		return nil
	}
	if !exists {
		return fmt.Errorf("command is intended for ostree-based systems only")
	}
	return fmt.Errorf("failed checking if %s exists: %w", ostreeBootedFile, err)
}

func NewOstreeCommand() *cobra.Command {
	preRun := &cobra.Command{
		Use:   "pre-run",
		Short: "Execute ostree-specific pre-run procedures",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ostree.PreRun()
		},
		PreRunE: failIfNotOstree,
	}
	scheduleBackup := &cobra.Command{
		Use:   "schedule-backup",
		Short: "Schedule backup for next boot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ostree.ScheduleBackup()
		},
		PreRunE: failIfNotOstree,
	}

	cmd := &cobra.Command{
		Use:   "ostree",
		Short: "Functionality for ostree-based systems",
	}

	cmd.AddCommand(scheduleBackup)
	cmd.AddCommand(preRun)

	return cmd
}
