package cmd

import (
	"fmt"
	"time"

	"github.com/openshift/microshift/pkg/admin"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/version"

	"github.com/spf13/cobra"
)

func newAdminBackupCommand() *cobra.Command {
	backup := &cobra.Command{
		Use:   "backup",
		Short: "Makes a backup of MicroShift data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return admin.MakeBackup(
				cmd.Flag("dest").Value.String(),
				cmd.Flag("name").Value.String(),
			)
		},
	}
	v := version.Get()
	backup.PersistentFlags().String(
		"dest",
		config.BackupsDir,
		"Directory with backups",
	)
	backup.PersistentFlags().String(
		"name",
		fmt.Sprintf("%s.%s__%s", v.Major, v.Minor, time.Now().UTC().Format("20060102_150405")),
		"Backup name",
	)
	return backup
}

func NewAdminCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Commands for managing MicroShift",
	}
	cmd.AddCommand(newAdminBackupCommand())
	return cmd
}
