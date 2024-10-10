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

func GetBackupName() (data.BackupName, error) {
	isOstree, err := util.PathExists("/run/ostree-booted")
	if err != nil {
		return "", fmt.Errorf("failed to check if system is ostree: %w", err)
	}

	var versionPart string
	if isOstree {
		versionPart, err = prerun.GetCurrentDeploymentID()
		if err != nil {
			return "", err
		}
	} else {
		execVer, err := prerun.GetVersionOfExecutable()
		if err != nil {
			return "", fmt.Errorf("failed to get version of MicroShift executable: %w", err)
		}
		versionPart = execVer.String()
	}

	// Format current time as yyyymmddHHMMSS
	t := time.Now().Format("20060102150405")

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
