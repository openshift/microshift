package data

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"

	"k8s.io/klog/v2"
)

var (
	cpArgs = []string{
		"--verbose",
		"--recursive",
		"--preserve",
		"--reflink=auto",
	}
)

// MakeBackup backs up MicroShift data (/var/lib/microshift) to
// target/name/ (e.g. /var/lib/microshift-backups/backup-00001).
func makeBackup(cfg BackupConfig) error {
	klog.InfoS("Backup started", "cfg", cfg)

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid BackupConfig: %w", err)
	}

	if err := microshiftIsNotRunning(); err != nil {
		return err
	}

	if err := ensureDirExists(cfg.Storage); err != nil {
		return err
	}

	dest := filepath.Join(cfg.Storage, cfg.Name)
	dest_tmp := dest + ".tmp"
	dest_old := dest + ".old"

	if err := removePath(dest_tmp); err != nil {
		return err
	}

	if err := copyDataDir(dest_tmp); err != nil {
		return err
	}

	backupAlreadyExists, err := pathExists(dest)
	if err != nil {
		return err
	} else if backupAlreadyExists {
		klog.V(2).InfoS("Backup already exists - renaming temporarily", "path", dest)
		if err := renamePath(dest, dest_old); err != nil {
			return err
		}
	}

	if err := renamePath(dest_tmp, dest); err != nil {
		klog.Errorf("Renaming path failed - renaming %s back and deleting %s: %v", dest_old, dest_tmp, err)
		renameErr := renamePath(dest_old, dest)
		rmErr := removePath(dest_tmp)
		return errors.Join(err, renameErr, rmErr)
	}

	if backupAlreadyExists {
		if err := removePath(dest_old); err != nil {
			return err
		}
	}

	klog.InfoS("Backup finished", "backup", dest, "data", config.DataDir)
	return nil
}

func ensureDirExists(path string) error {
	klog.V(2).InfoS("Making sure directory exists", "path", path)

	found, err := pathExists(path)
	if err != nil {
		return err
	}
	if found {
		klog.V(2).InfoS("Directory already exists", "path", path)
		return nil
	}

	err = util.MakeDir(path)
	if err != nil {
		return fmt.Errorf("failed creating %s: %w", path, err)
	}
	klog.InfoS("Directory created", "path", path)
	return nil
}

func removePath(path string) error {
	klog.V(2).InfoS("Path removal attempt", "path", path)

	exists, err := pathExists(path)
	if err != nil {
		return err
	}

	if exists {
		klog.V(2).InfoS("Path exists - removing", "path", path)
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
		klog.InfoS("Path removed", "path", path)
	} else {
		klog.V(2).InfoS("Path not found", "path", path)
	}
	return nil
}

func copyDataDir(dest string) error {
	cmd := exec.Command("cp", append(cpArgs, config.DataDir, dest)...) //nolint:gosec
	klog.V(2).InfoS("Executing command", "cmd", cmd)

	out, err := cmd.CombinedOutput()
	klog.V(2).InfoS("Command finished running", "cmd", cmd, "output", out)

	if err != nil {
		klog.ErrorS(err, "Command failed", "cmd", cmd, "output", out)
		return fmt.Errorf("failed to copy data: %w", err)
	}

	klog.InfoS("Command successful", "cmd", cmd)
	return nil
}

func renamePath(from, to string) error {
	klog.V(2).InfoS("Renaming path", "from", from, "to", to)

	if err := os.Rename(from, to); err != nil {
		klog.ErrorS(err, "Failed to rename path", "from", from, "to", to)
		return fmt.Errorf("renaming %s to %s failed: %w", from, to, err)
	}

	klog.InfoS("Path renamed", "from", from, "to", to)
	return nil
}

func pathExists(path string) (bool, error) {
	exists, err := util.PathExists(path)
	if err != nil {
		klog.ErrorS(err, "Failed to check if path exists", "path", path)
		return false, fmt.Errorf("checking if %s exists failed: %w", path, err)
	}
	return exists, nil
}
