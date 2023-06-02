package prerun

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

func Perform() error {
	healthFile := "/var/lib/microshift-backups/health.json"
	exists, err := util.PathExists(healthFile)
	if err != nil {
		klog.ErrorS(err, "Could not access health file", "path", healthFile)
		return err
	}
	if !exists {
		klog.InfoS("Boot health information is missing - skipping backup")
		return nil
	}

	currentBootID, err := getCurrentBootID()
	if err != nil {
		klog.ErrorS(err, "Failed to get current boot ID")
		return err
	}
	klog.InfoS("Current boot", "id", currentBootID)

	info := struct {
		Health       string `json:"health"`
		DeploymentID string `json:"deployment_id"`
		BootID       string `json:"boot_id"`
	}{}
	d, err := os.ReadFile(healthFile)
	if err != nil {
		klog.ErrorS(err, "Failed to read file", "path", healthFile)
		return err
	}
	if err := json.Unmarshal(d, &info); err != nil {
		klog.ErrorS(err, "Failed to unmarshal json", "data", string(d))
		return err
	}

	klog.InfoS("Read health info", "info", info)

	if info.BootID == currentBootID {
		klog.InfoS("Current boot in health file - skipping backup")
		return nil
	}

	if info.Health != "healthy" {
		return nil
	}
	name := data.BackupName(fmt.Sprintf("%s_%s", info.DeploymentID, info.BootID))

	dm, err := data.NewManager(config.BackupsDir)
	if err != nil {
		return err
	}

	backups, err := dm.GetBackupList()
	if err != nil {
		return err
	}

	existingDeploymentBackups := make([]data.BackupName, 0)
	for _, b := range backups {
		if name == b {
			klog.InfoS("Backup already exists", "deployment", info.DeploymentID, "boot", info.BootID)
			return nil
		}
		if strings.Contains(string(b), info.DeploymentID) {
			existingDeploymentBackups = append(existingDeploymentBackups, b)
		}
	}

	if err := dm.Backup(name); err != nil {
		return err
	}

	if len(existingDeploymentBackups) > 0 {
		klog.InfoS("Removing old deployment backups",
			"deployment", info.DeploymentID,
			"backups", existingDeploymentBackups)

		for _, b := range existingDeploymentBackups {
			if err := dm.RemoveBackup(b); err != nil {
				klog.ErrorS(err, "Failed to remove backup", "name", b)
			}
		}
	}

	return nil
}

func getCurrentBootID() (string, error) {
	content, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(strings.TrimSpace(string(content)), "-", ""), nil
}
