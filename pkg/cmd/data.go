package cmd

import (
	"fmt"

	"github.com/openshift/microshift/pkg/backup"
	"github.com/openshift/microshift/pkg/config"
	"github.com/spf13/cobra"
)

func NewDataCommand() *cobra.Command {
	backup := &cobra.Command{
		Use:   "backup",
		Short: "Perform backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			return backup.MakeBackup(cmd.Flag("name").Value.String())
		},
	}
	backup.PersistentFlags().String("name", "backup-0",
		fmt.Sprintf("Name of the backup. It is a name of directory inside %s", config.AuxDataDir))

	cmd := &cobra.Command{
		Use:   "data",
		Short: "MicroShift data management",
	}

	cmd.AddCommand(backup)

	return cmd
}
