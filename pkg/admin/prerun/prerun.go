package prerun

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

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

	info := struct {
		Health       string `json:"health"`
		DeploymentID string `json:"deployment_id"`
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
	if info.Health != "healthy" {
		return nil
	}

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
		if strings.Contains(string(b), info.DeploymentID) {
			existingDeploymentBackups = append(existingDeploymentBackups, b)
		}
	}

	name := fmt.Sprintf("%s_%s", info.DeploymentID, time.Now().UTC().Format("20060102_150405"))
	if err := dm.Backup(data.BackupName(name)); err != nil {
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
