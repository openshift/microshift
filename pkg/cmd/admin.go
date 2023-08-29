package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/admin/prerun"
	"github.com/openshift/microshift/pkg/config"

	"github.com/spf13/cobra"
)

func NewBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup MicroShift data",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flag("storage").Value.String() == "" {
				return fmt.Errorf("--storage must not be empty")
			}

			if err := data.MicroShiftIsNotRunning(); err != nil {
				return fmt.Errorf("microshift must not be running: %w", err)
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
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
				return fmt.Errorf("backup %q already exists", dataManager.GetBackupPath(name))
			}

			return dataManager.Backup(name)
		},
	}

	cmd.PersistentFlags().String("name", "",
		"Backup name (if not provided, name is based on version of MicroShift data directory and current date and time)")
	cmd.PersistentFlags().String("storage", config.BackupsDir, "Directory with backups")

	return cmd
}

func NewRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore backup-path",
		Short: "Restore MicroShift data from a backup",
		Args:  cobra.ExactArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := data.MicroShiftIsNotRunning(); err != nil {
				return fmt.Errorf("microshift must not be running: %w", err)
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			// Remove trailing slash (/) so filepath.Dir gives us parent dir path.
			// Otherwise, it would return the same path we got from a user.
			path, _ := strings.CutSuffix(args[0], string(os.PathSeparator))
			storage := data.StoragePath(filepath.Dir(path))
			name := data.BackupName(filepath.Base(path))

			if storage == "" || name == "" {
				return fmt.Errorf("unexpected problem when parsing given path (%q)"+
					" after transformation (%q)"+
					" - storage (%q) or name (%q) is empty",
					args[0], path, storage, name)
			}

			dataManager, err := data.NewManager(storage)
			if err != nil {
				return err
			}

			return dataManager.Restore(name)
		},
	}

	return cmd
}
