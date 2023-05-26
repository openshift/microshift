package data

import (
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
	if cfg.Storage == "" {
		return fmt.Errorf("backup storage must not be empty")
	}
	if cfg.Name == "" {
		return fmt.Errorf("backup name must not be empty")
	}

	if err := microshiftShouldNotRun(); err != nil {
		return err
	}

	if err := ensureDirExists(cfg.Storage); err != nil {
		return err
	}

	dest := filepath.Join(cfg.Storage, cfg.Name)
	dest_tmp := dest + ".tmp"

	if err := dirShouldNotExist(dest_tmp); err != nil {
		return err
	}

	if err := copyDataDir(dest_tmp); err != nil {
		return err
	}

	if err := removeDirBeforeRenaming(dest); err != nil {
		return err
	}

	if err := renameDir(dest_tmp, dest); err != nil {
		return err
	}

	klog.InfoS("Data backed up", "data", config.DataDir, "backup", dest)
	return nil
}

func ensureDirExists(path string) error {
	klog.Infof("Making sure %s exists", path)

	found, err := util.FileExists(path)
	if err != nil {
		return fmt.Errorf("failed checking if %s exists: %w", path, err)
	}
	if found {
		klog.Infof("Directory %s already exists", path)
		return nil
	}

	err = util.MakeDir(path)
	if err != nil {
		return fmt.Errorf("failed creating %s: %w", path, err)
	}
	klog.Infof("Directory %s created", path)
	return nil
}

func dirShouldNotExist(path string) error {
	klog.Infof("Making sure %s does not exist", path)

	found, err := util.FileExists(path)
	if err != nil {
		return fmt.Errorf("failed to check if %s exists: %w", path, err)
	}
	if found {
		return fmt.Errorf("directory %s already exists", path)
	}
	return nil
}

func removeDirBeforeRenaming(path string) error {
	found, err := util.FileExists(path)
	if err != nil {
		return fmt.Errorf("failed to check if %s exists: %w", path, err)
	}
	if found {
		klog.Infof("Removing %s before renaming tmp directory", path)
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}
	return nil
}

func copyDataDir(dest string) error {
	cmd := exec.Command("cp", append(cpArgs, config.DataDir, dest)...) //nolint:gosec
	klog.InfoS("Executing command", "cmd", cmd)
	out, err := cmd.CombinedOutput()
	klog.InfoS("Command finished running", "cmd", cmd, "output", out)

	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}
	return nil
}

func renameDir(from, to string) error {
	klog.InfoS("Renaming directory", "from", from, "to", to)
	err := os.Rename(from, to)
	if err == nil {
		return nil
	}

	klog.ErrorS(err, "Renaming directory failed - deleting 'from' dir", "from", from, "to", to)
	if rmErr := os.RemoveAll(from); rmErr != nil {
		klog.ErrorS(rmErr, "Failed to remove directory", "dir", from)
		return fmt.Errorf("failed to remove %s: %w: %w", from, err, rmErr)
	}
	return nil
}
