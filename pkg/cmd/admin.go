package cmd

import (
	"fmt"
	"time"

	"github.com/openshift/microshift/pkg/admin"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/version"

	"github.com/spf13/cobra"
)

func NewAdminCommand() *cobra.Command {
	backup := &cobra.Command{
		Use:   "backup",
		Short: "Back up current data",
		Long:  fmt.Sprintf("Backs up current MicroShift data to %s", config.BackupsDir),
		RunE: func(cmd *cobra.Command, args []string) error {
			return admin.MakeBackup(cmd.Flag("name").Value.String())
		},
	}

	v := version.Get()
	backup.PersistentFlags().String("name",
		fmt.Sprintf("%s.%s__%s", v.Major, v.Minor, time.Now().UTC().Format("20060102_150405")),
		fmt.Sprintf("Backup dir name in %s", config.BackupsDir))

	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Commands for managing MicroShift",
	}
	cmd.AddCommand(backup)

	return cmd
}
