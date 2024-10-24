package data

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"syscall"

	"github.com/openshift/microshift/pkg/config"

	"k8s.io/klog/v2"
)

func GetSizeOfMicroShiftData() (uint64, error) {
	return GetSizeOfDir(config.DataDir)
}

func GetSizeOfDir(path string) (uint64, error) {
	var size int64
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get size of %q: %w", path, err)
	}
	klog.Infof("Calculated size of %q: %vM", path, size/1024/1024)
	return uint64(size), nil
}

func GetAvailableDiskSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("failed to get available disk space of %q: %w", path, err)
	}
	available := stat.Bavail * uint64(stat.Bsize)
	klog.Infof("Calculated available disk space for %q: %vM", path, available/1024/1024)
	return available, nil
}

// CheckIfEnoughSpaceToRestore performs a naive check if there is enough
// disk space on /var/lib filesystem to restore the backup.
// It does not accommodate for potential Copy-on-Write disk savings.
func CheckIfEnoughSpaceToRestore(backupPath string) error {
	backupSize, err := GetSizeOfDir(backupPath)
	if err != nil {
		return err
	}

	// Restore process: renames MicroShift data dir, copies backup into place, deletes renamed copy hence the /var/lib
	// needs extra space before old MicroShift data is removed.
	availableSpace, err := GetAvailableDiskSpace("/var/lib")
	if err != nil {
		return err
	}

	if availableSpace < backupSize {
		return fmt.Errorf(
			"not enough disk space in /var/lib to restore the backup: required=%vM available=%vM",
			backupSize/1024/1024,
			availableSpace/1024/1024,
		)
	}

	return nil
}

// CheckIfEnoughSpaceToBackUp performs a naive check if there is enough
// disk space on a filesystem holding the backups for another backup of MicroShift data.
// It does not accommodate for potential Copy-on-Write disk savings.
func CheckIfEnoughSpaceToBackUp(storage string) error {
	dataSize, err := GetSizeOfMicroShiftData()
	if err != nil {
		return err
	}

	availableSpace, err := GetAvailableDiskSpace(storage)
	if err != nil {
		return err
	}

	if availableSpace < dataSize {
		return fmt.Errorf(
			"not enough disk space in %q to create a backup: required=%vM available=%vM",
			storage,
			dataSize/1024/1024,
			availableSpace/1024/1024,
		)
	}

	return nil
}
