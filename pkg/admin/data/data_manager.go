package data

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/pkg/errors"
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

func (dm *manager) RemoveBackup(name BackupName) error {
	klog.InfoS("Removing backup",
		"name", name,
	)
	if err := os.RemoveAll(dm.GetBackupPath(name)); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to delete backup %q", name))
	}
	klog.InfoS("Removed backup",
		"name", name,
	)
	return nil
}

func (dm *manager) GetBackupList() ([]BackupName, error) {
	files, err := os.ReadDir(config.BackupsDir)
	if err != nil {
		return nil, err
	}

	backups := make([]BackupName, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			backups = append(backups, BackupName(file.Name()))
		}
	}

	return backups, nil
}

func (dm *manager) Backup(name BackupName) error {
	klog.InfoS("Starting backup",
		"storage", dm.storage,
		"name", name,
		"data", config.DataDir,
	)

	if name == "" {
		return &EmptyArgErr{"name"}
	}

	if exists, err := dm.BackupExists(name); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to determine if backup %s exists", name))
	} else if exists {
		return fmt.Errorf("Failed to create backup destination %q because it already exists", name)
	}

	if found, err := pathExists(string(dm.storage)); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to determine if storage location %q exists", dm.storage))
	} else if !found {
		if makeDirErr := util.MakeDir(string(dm.storage)); makeDirErr != nil {
			return errors.Wrap(makeDirErr, fmt.Sprintf("Failed to create backup storage directory %q", dm.storage))
		}
		klog.InfoS("Created backup storage directory", "path", dm.storage)
	} else {
		klog.InfoS("Found existing backup storage directory", "path", dm.storage)
	}

	dest := dm.GetBackupPath(name)

	if err := copyPath(config.DataDir, dest); err != nil {
		return err
	}

	klog.InfoS("Backup finished", "backup", dest, "data", config.DataDir)
	return nil
}

func (dm *manager) Restore(name BackupName) error {
	klog.InfoS("Starting restore",
		"storage", dm.storage,
		"name", name,
		"data", config.DataDir,
	)

	if name == "" {
		return &EmptyArgErr{"name"}
	}

	if exists, err := dm.BackupExists(name); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to determine if backup %q exists failed", name))
	} else if !exists {
		return fmt.Errorf("Failed to restore backup, %q does not exist", name)
	}

	tmp := fmt.Sprintf("%s.saved", config.DataDir)
	klog.InfoS("Renaming existing data dir", "data", config.DataDir, "renamedTo", tmp)
	if err := os.Rename(config.DataDir, tmp); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to rename existing data directory %q to %q", config.DataDir, tmp))
	}

	src := dm.GetBackupPath(name)
	if err := copyPath(src, config.DataDir); err != nil {
		klog.ErrorS(err, "Failed to restore from backup, restoring current data dir")

		if err := os.RemoveAll(config.DataDir); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to remove data directory %q", config.DataDir))
		}

		if err := os.Rename(tmp, config.DataDir); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to rename temporary directory %q to %q", tmp, config.DataDir))
		}

		return errors.Wrap(err, "Failed to restore backup")
	}

	klog.InfoS("Removing temporary data directory", "path", tmp)
	if err := os.RemoveAll(tmp); err != nil {
		klog.ErrorS(err, "Failed to remove temporary data directory, leaving in place", "path", tmp)
	}

	klog.InfoS("Finished restore",
		"name", name,
		"data", config.DataDir,
	)
	return nil
}

func copyPath(src, dest string) error {
	cmd := exec.Command("cp", append(cpArgs, src, dest)...) //nolint:gosec
	klog.InfoS("Starting copy", "cmd", cmd)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()

	if err != nil {
		klog.InfoS("Failed to copy", "cmd", cmd,
			"stdout", strings.ReplaceAll(outb.String(), "\n", `, `),
			"stderr", errb.String())
		return errors.Wrap(err, fmt.Sprintf("Failed to copy %q to %q", src, dest))
	}

	klog.InfoS("Finished copy", "cmd", cmd)
	return nil
}

func pathExists(path string) (bool, error) {
	exists, err := util.PathExists(path)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("failed to check if %q exists", path))
	}
	return exists, nil
}
