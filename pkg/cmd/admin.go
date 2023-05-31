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
	storage := data.StoragePath(cmd.Flag("storage").Value.String())
	name := data.BackupName(cmd.Flag("name").Value.String())

	dataManager, err := data.NewManager(storage)
	if err != nil {
		return err
	}

	if exists, err := dataManager.BackupExists(name); err != nil {
		return err
	} else if exists {
		if force, err := cmd.Flags().GetBool("force"); err != nil {
			return err
		} else if !force {
			return fmt.Errorf("backup %s already exists, use --force to overwrite",
				dataManager.GetBackupPath(name))
		}
	}

	return dataManager.Backup(name)
}

func newAdminDataCommand() *cobra.Command {
	backup := &cobra.Command{
		Use:   "backup",
		Short: "Backup MicroShift data",
		RunE:  backup,
	}
	backup.PersistentFlags().Bool("force", false, "Overwrite existing backup")

	data := &cobra.Command{
		Use:   "data",
		Short: "Commands for managing MicroShift data",
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
	v := version.Get()
	data.PersistentFlags().String("storage", config.BackupsDir, "Directory with backups")
	data.PersistentFlags().String("name",
		fmt.Sprintf("%s.%s__%s", v.Major, v.Minor, time.Now().UTC().Format("20060102_150405")),
		"Backup name")

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
