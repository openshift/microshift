package cmd

import (
	"fmt"
	"time"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/admin/prerun"
	"github.com/openshift/microshift/pkg/config"

	"github.com/spf13/cobra"
)

func backup(cmd *cobra.Command, args []string) error {
	storage := data.StoragePath(cmd.Flag("storage").Value.String())
	name := data.BackupName(cmd.Flag("name").Value.String())

	if name == "" {
		// If backup name is not given by user, construct a "default value".
		// We cannot do it when creating cobra's flag because reading
		// /var/lib/microshift/version requires elevated permissions
		// and it would be poor UX to expect sudo for --help.
		name = data.BackupName(fmt.Sprintf("%s__%s",
			prerun.GetVersionStringOfData(),
			time.Now().UTC().Format("20060102_150405")))
	}

	dataManager, err := data.NewManager(storage)
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

func newAdminDataCommand() *cobra.Command {
	backup := &cobra.Command{
		Use:   "backup",
		Short: "Backup MicroShift data",
		RunE:  backup,
	}

	data := &cobra.Command{
		Use:   "data",
		Short: "Commands for managing MicroShift data",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flag("storage").Value.String() == "" {
				return fmt.Errorf("--storage must not be empty")
			}

			if err := data.MicroShiftIsNotRunning(); err != nil {
				return fmt.Errorf("microshift must not be running: %w", err)
			}

			return nil
		},
	}

	data.PersistentFlags().String("name", "",
		"Backup name (if not provided, name is based on version of MicroShift data directory and current date and time)")
	data.PersistentFlags().String("storage", config.BackupsDir, "Directory with backups")

	data.AddCommand(backup)
	return data
}

func NewAdminCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Commands for managing MicroShift",
	}
	cmd.AddCommand(newAdminDataCommand())
	return cmd
}
