package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openshift/microshift/pkg/admin/autorecovery"
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
			fmt.Fprintf(os.Stderr, "WARNING: Service %q is %q - backup can potentially contain unhealthy data\n", service, state)
		}

		if state != "inactive" && state != "failed" {
			return fmt.Errorf("MicroShift must be stopped before creating or restoring backup (%q is %q, should be %q or %q)",
				service, state, "inactive", "failed")
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

		if autorec, err := cmd.Flags().GetBool("auto-recovery"); err != nil {
			return fmt.Errorf("failed to get `auto-recovery` flag: %w", err)
		} else if autorec {
			// For auto-recovery, the existence of the storage is not important.
			// If it doesn't exist before backup, MicroShift will create it.
			return nil
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

func validateArgs(cmd *cobra.Command, args []string) error {
	var err error
	if len(args) == 0 {
		err = fmt.Errorf("command requires an argument")
	} else if len(args) > 1 {
		err = fmt.Errorf("command accepts only 1 argument")
	} else if args[0] == "" {
		err = fmt.Errorf("argument cannot be empty")
	}

	if err != nil {
		// Remove 'Global Flags' and everything after because
		// it contains some hidden flags.
		usage, _, _ := strings.Cut(cmd.UsageString(), "Global Flags")
		fmt.Printf("Error: %v\n\n%s", err, usage)
		os.Exit(1)
	}
	return nil
}

func NewBackupCommand() *cobra.Command {
	autorec := false

	cmd := &cobra.Command{
		Use:               "backup PATH",
		Short:             "Create a backup of MicroShift data",
		Long:              "Create a backup of MicroShift data. PATH should not exist.",
		Args:              validateArgs,
		PersistentPreRunE: backupRestorePreRun(true),

		RunE: func(cmd *cobra.Command, args []string) error {
			// err is checked in PersistentPreRunE
			storage, name, _ := backupPathToStorageAndName(args[0])

			if autorec {
				// For auto-recovery mode we treat given path as a directory where the backup subdirectory will be created.
				// Normally it's interpreted as final destination.
				storage = data.StoragePath(args[0])
				if err := autorecovery.CreateStorageIfAbsent(storage); err != nil {
					return err
				}
				var err error
				name, err = autorecovery.GetBackupName()
				if err != nil {
					return err
				}
			}

			dataManager, err := data.NewManager(storage)
			if err != nil {
				return err
			}

			backupPath, err := dataManager.Backup(name)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", backupPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&autorec, "auto-recovery", false,
		`Changes the behavior of the backup command.
The PATH argument will be treated as a directory where backups are
created using the naming scheme compatible with "restore --auto-recovery"`)

	return cmd
}

func NewRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "restore PATH",
		Short:             "Restore MicroShift data from a backup",
		Args:              validateArgs,
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
