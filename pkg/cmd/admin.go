package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/util"

	"github.com/spf13/cobra"
)

func shouldRunPrivileged() error {
	if os.Geteuid() > 0 {
		return fmt.Errorf("command requires root privileges")
	}
	return nil
}

func backupPathToStorageAndName(path string) (data.StoragePath, data.BackupName, error) {
	if path == "" {
		return "", "", fmt.Errorf("provided path is empty")
	}

	// filepath.Clean() also removes trailing slash
	// (e.g. `/var/lib/microshift-backups/my-backup/` -> `/var/lib/microshift-backups/my-backup`)
	// so filepath.Dir will give us parent dir (`/var/lib/microshift-backups`)
	// and filepath.Base will give us name of backup dir (`my-backup`).
	path = filepath.Clean(path)
	storage := data.StoragePath(filepath.Dir(path))
	name := data.BackupName(filepath.Base(path))

	if storage == "" {
		return "", "", fmt.Errorf("parsing %q resulted in empty backup location: %q", path, storage)
	}

	if name == "" {
		return "", "", fmt.Errorf("parsing %q resulted in empty backup name: %q", path, name)
	}

	if name == "/" {
		return "", "", fmt.Errorf("%q contains invalid backup name: %q", path, name)
	}

	return storage, name, nil
}

func NewBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup PATH",
		Short: "Create a backup of MicroShift data",
		Long:  "Create a backup of MicroShift data. PATH should not exist.",
		Args:  cobra.ExactArgs(1),

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := shouldRunPrivileged(); err != nil {
				return err
			}

			if err := util.PathShouldNotExist(args[0]); err != nil {
				return err
			}

			if err := data.MicroShiftIsNotRunning(); err != nil {
				return fmt.Errorf("microshift must not be running: %w", err)
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			storage, name, err := backupPathToStorageAndName(args[0])
			if err != nil {
				return err
			}

			dataManager, err := data.NewManager(storage)
			if err != nil {
				return err
			}

			return dataManager.Backup(name)
		},
	}

	return cmd
}

func NewRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore PATH",
		Short: "Restore MicroShift data from a backup",
		Args:  cobra.ExactArgs(1),

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := shouldRunPrivileged(); err != nil {
				return err
			}

			if err := util.PathShouldExist(args[0]); err != nil {
				return err
			}

			if err := data.MicroShiftIsNotRunning(); err != nil {
				return fmt.Errorf("microshift must not be running: %w", err)
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			storage, name, err := backupPathToStorageAndName(args[0])
			if err != nil {
				return err
			}

			dataManager, err := data.NewManager(storage)
			if err != nil {
				return err
			}

			// TODO: Verify content of provided path
			// Check if provided path points to a directory that is really
			// a backup of MicroShift's data.

			return dataManager.Restore(name)
		},
	}

	return cmd
}
