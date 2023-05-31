package data

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func NewManager(storage StoragePath) (*manager, error) {
	if storage == "" {
		return nil, &EmptyArgErr{argName: "storage"}
	}
	return &manager{storage: storage}, nil
}

var _ Manager = (*manager)(nil)

type manager struct {
	storage StoragePath
}

func (dm *manager) GetBackupPath(name BackupName) string {
	return filepath.Join(string(dm.storage), string(name))
}

func (dm *manager) BackupExists(name BackupName) (bool, error) {
	return pathExists(dm.GetBackupPath(name))
}

func (dm *manager) Backup(name BackupName) error {
	klog.InfoS("Backing up the data",
		"storage", dm.storage, "name", name, "data", config.DataDir)

	if name == "" {
		return &EmptyArgErr{"name"}
	}

	if found, err := pathExists(string(dm.storage)); err != nil {
		return err
	} else if !found {
		if makeDirErr := util.MakeDir(string(dm.storage)); makeDirErr != nil {
			return fmt.Errorf("making %s directory failed: %w", dm.storage, makeDirErr)
		}
		klog.InfoS("Backup storage directory created", "path", dm.storage)
	} else {
		klog.InfoS("Backup storage directory already existed", "path", dm.storage)
	}

	dest := dm.GetBackupPath(name)
	tmp := dest + ".tmp"
	old := dest + ".old"

	// Make sure /storage/backup.tmp does not exist, so data isn't copied into that directory
	if err := os.RemoveAll(tmp); err != nil {
		return fmt.Errorf("failed to remove %s: %w", tmp, err)
	}

	if err := copyDataDir(tmp); err != nil {
		return err
	}

	backupExists, err := dm.BackupExists(name)
	if err != nil {
		return err
	} else if backupExists {
		if err := renamePath(dest, old); err != nil {
			return err
		}
		klog.InfoS("Temporarily renamed existing backup", "backup", dest, "renamed", old)
	}

	if err := renamePath(tmp, dest); err != nil {
		klog.Errorf("Renaming path failed - renaming %s back and deleting %s: %v", old, tmp, err)
		renameErr := renamePath(old, dest)
		rmErr := removePath(tmp)
		return errors.Join(err, renameErr, rmErr)
	}

	if backupExists {
		if err := removePath(old); err != nil {
			return err
		}
	}

	klog.InfoS("Backup finished", "backup", dest, "data", config.DataDir)
	return nil
}

func (dm *manager) Restore(n BackupName) error {
	return fmt.Errorf("Restore not implemented")
}

func removePath(path string) error {
	exists, err := pathExists(path)
	if err != nil {
		return err
	}

	if exists {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
		klog.InfoS("Removed path", "path", path)
	}
	return nil
}

func copyDataDir(dest string) error {
	cmd := exec.Command("cp", append(cpArgs, config.DataDir, dest)...) //nolint:gosec
	klog.InfoS("Executing command", "cmd", cmd)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()

	klog.InfoS("Command finished running", "cmd", cmd,
		"stdout", strings.ReplaceAll(outb.String(), "\n", `, `),
		"stderr", errb.String())

	if err != nil {
		return fmt.Errorf("command %s failed: %w", cmd, err)
	}

	klog.InfoS("Command successful", "cmd", cmd)
	return nil
}

func renamePath(from, to string) error {
	if err := os.Rename(from, to); err != nil {
		return fmt.Errorf("renaming %s to %s failed: %w", from, to, err)
	}
	return nil
}

func pathExists(path string) (bool, error) {
	exists, err := util.PathExists(path)
	if err != nil {
		return false, fmt.Errorf("checking if %s exists failed: %w", path, err)
	}
	return exists, nil
}
