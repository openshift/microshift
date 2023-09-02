package cmd

import (
	"fmt"
	"os"
	"os/exec"
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

func servicesShouldBeInactive(backingUp bool) error {
	var services = []string{"microshift.service", "microshift-etcd.scope"}

	for _, service := range services {
		cmd := exec.Command("systemctl", "show", "-p", "ActiveState", "--value", service)
		out, err := cmd.CombinedOutput()
		state := strings.TrimSpace(string(out))
		if err != nil {
			return fmt.Errorf("error when checking if %q is active: %w", service, err)
		}

		if state == "failed" && backingUp {
			return fmt.Errorf("service %q is %q which suggest that MicroShift data might be unhealthy - "+
				"address any issues before attempting to create a backup again", service, state)
		}

		if state != "inactive" && state != "failed" {
			return fmt.Errorf("MicroShift must be stopped before creating or restoring backup (%q is %q, should be %q)",
				service, state, "inactive")
		}
	}

	return nil
}

func checkPathExistence(path string, shouldExist bool) error {
	exists, err := util.PathExists(path)
	if err != nil {
		return err
	}

	if shouldExist && !exists {
		return fmt.Errorf("expected %q to exist", path)
	}

	if !shouldExist && exists {
		return fmt.Errorf("expected %q to not exist", path)
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

// backupRestorePreRun contains necessary checks before attempting to perform
// backup or restore. It is meant to be used as PersistentPreRunE.
// It is not a part of RunE because k8s.io/component-base/cli.Run() wrapper
// sets up klog  which considerably decreases readability of errors because
// they're hidden in logging information such as pid, datetime,
// source file and line.
func backupRestorePreRun(backingUp bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := shouldRunPrivileged(); err != nil {
			return err
		}

		if err := servicesShouldBeInactive(backingUp); err != nil {
			return err
		}

		path := args[0]
		_, _, err := backupPathToStorageAndName(path)
		if err != nil {
			return err
		}

		pathShouldExist := !backingUp
		if err := checkPathExistence(path, pathShouldExist); err != nil {
			return err
		}

		return nil
	}
}

func NewBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "backup PATH",
		Short:             "Create a backup of MicroShift data",
		Long:              "Create a backup of MicroShift data. PATH should not exist.",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: backupRestorePreRun(true),

		RunE: func(cmd *cobra.Command, args []string) error {
			// err is checked in PersistentPreRunE
			storage, name, _ := backupPathToStorageAndName(args[0])
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
		Use:               "restore PATH",
		Short:             "Restore MicroShift data from a backup",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: backupRestorePreRun(false),

		RunE: func(cmd *cobra.Command, args []string) error {
			// err is checked in PersistentPreRunE
			storage, name, _ := backupPathToStorageAndName(args[0])
			dataManager, err := data.NewManager(storage)
			if err != nil {
				return err
			}

			return dataManager.Restore(name)
		},
	}

	return cmd
}
