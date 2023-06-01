package cmd

import (
	"fmt"
	"time"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/version"

	"github.com/spf13/cobra"
)

func backup(cmd *cobra.Command, args []string) error {
	storage, err := cmd.Flags().GetString("storage")
	if err != nil {
		return err
	}
	nameArg, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	name := data.BackupName(nameArg)

	dataManager, err := data.NewManager(data.StoragePath(storage))
	if err != nil {
		return err
	}

	if exists, err := dataManager.BackupExists(name); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("backup %s already exists", dataManager.GetBackupPath(name))
	}

	return dataManager.Backup(name)
}

func defaultBackupName() string {
	v := version.Get()
	return fmt.Sprintf("%s.%s__%s", v.Major, v.Minor, time.Now().UTC().Format("20060102_150405"))
}

func NewDataCommand() *cobra.Command {
	data := &cobra.Command{
		Use:   "data",
		Short: "Manage MicroShift data and backups",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flag("storage").Value.String() == "" {
				return fmt.Errorf("--storage must not be empty")
			}

			if cmd.Flag("name").Value.String() == "" {
				return fmt.Errorf("--name must not be empty")
			}

			if err := data.MicroShiftIsNotRunning(); err != nil {
				return fmt.Errorf("microshift must not be running: %w", err)
			}

			return nil
		},
	}
	data.PersistentFlags().String("storage", config.BackupsDir, "Directory with backups")
	data.PersistentFlags().String("name", defaultBackupName(), "Backup name")

	backup := &cobra.Command{
		Use:   "backup",
		Short: "Make a backup of MicroShift data",
		RunE:  backup,
	}
	data.AddCommand(backup)

	return data
}
