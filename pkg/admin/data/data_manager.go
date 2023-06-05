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
	"k8s.io/apimachinery/pkg/util/rand"
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
	return os.RemoveAll(dm.GetBackupPath(name))
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
	klog.InfoS("Backing up the data",
		"storage", dm.storage, "name", name, "data", config.DataDir)

	if name == "" {
		return &EmptyArgErr{"name"}
	}

	if exists, err := dm.BackupExists(name); err != nil {
		return fmt.Errorf("checking if backup %s exists failed: %w", name, err)
	} else if exists {
		klog.ErrorS(nil, "Backup already exists - name should be unique", "name", name)
		return fmt.Errorf("backup %s already exists", name)
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

	if err := copy(config.DataDir, dest); err != nil {
		return err
	}

	klog.InfoS("Backup finished", "backup", dest, "data", config.DataDir)
	return nil
}

func (dm *manager) Restore(name BackupName) error {
	klog.InfoS("Restoring the data", "storage", dm.storage, "name", name, "data", config.DataDir)

	if name == "" {
		return &EmptyArgErr{"name"}
	}

	if exists, err := dm.BackupExists(name); err != nil {
		return fmt.Errorf("checking if backup %s exists failed: %w", name, err)
	} else if !exists {
		klog.ErrorS(nil, "Backup to restore does not exist", "name", name)
		return fmt.Errorf("backup %s does not exist", name)
	}

	tmp := fmt.Sprintf("%s.%s", config.DataDir, rand.String(8))
	klog.InfoS("Temporarily renaming data dir", "data", config.DataDir, "renamedTo", tmp)
	if err := os.Rename(config.DataDir, tmp); err != nil {
		return fmt.Errorf("renaming data dir failed: %w", err)
	}

	src := dm.GetBackupPath(name)
	if err := copy(src, config.DataDir); err != nil {
		return err
	}

	klog.InfoS("Removing temporary data dir", "path", tmp)
	if err := os.RemoveAll(tmp); err != nil {
		klog.ErrorS(err, "Failed to remove %s", tmp)
	}

	klog.InfoS("Restore finished", "backup", src, "data", config.DataDir)
	return nil
}

func copy(src, dest string) error {
	cmd := exec.Command("cp", append(cpArgs, src, dest)...) //nolint:gosec
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

func pathExists(path string) (bool, error) {
	exists, err := util.PathExists(path)
	if err != nil {
		return false, fmt.Errorf("checking if %s exists failed: %w", path, err)
	}
	return exists, nil
}
