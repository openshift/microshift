package autorecovery

import (
	"fmt"
	"os"
	"time"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/admin/prerun"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

const (
	// Date format: yyyymmddHHMMSS
	backupCreationTimeFormat = "20060102150405"
)

func getVersion() (string, error) {
	isOstree, err := util.IsOSTree()
	if err != nil {
		return "", fmt.Errorf("failed to check if system is ostree: %w", err)
	}

	if isOstree {
		deployID, err := prerun.GetCurrentDeploymentID()
		if err != nil {
			return "", err
		}
		return deployID, nil
	}

	execVer, err := prerun.GetVersionOfExecutable()
	if err != nil {
		return "", fmt.Errorf("failed to get version of MicroShift executable: %w", err)
	}
	return execVer.String(), nil
}

func GetBackupName() (data.BackupName, error) {
	versionPart, err := getVersion()
	if err != nil {
		return "", err
	}

	t := time.Now().Format(backupCreationTimeFormat)
	return data.BackupName(fmt.Sprintf("%s_%s", t, versionPart)), nil
}

func CreateStorageIfAbsent(storagePath data.StoragePath) error {
	path := string(storagePath)
	exists, err := util.PathExists(path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	klog.Infof("%q doesn't exist - creating", path)
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("failed to create auto-recovery backup storage %q: %w", path, err)
	}

	return nil
}
