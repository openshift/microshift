package prerun

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

var (
	errHealthFileDoesNotExist = errors.New("health file does not exist")
)

type HealthInfo struct {
	Health       string `json:"health"`
	DeploymentID string `json:"deployment_id"`
	BootID       string `json:"boot_id"`
}

func (hi *HealthInfo) BackupName() data.BackupName {
	return data.BackupName(fmt.Sprintf("%s_%s", hi.DeploymentID, hi.BootID))
}

func (hi *HealthInfo) IsHealthy() bool {
	return hi.Health == "healthy"
}

func Perform() error {
	health, err := getHealthInfo()
	if err != nil {
		if errors.Is(err, errHealthFileDoesNotExist) {
			klog.InfoS("Health file does not exist - skipping backup")
			return nil
		}
		klog.ErrorS(err, "Failed to load health from disk")
		return err
	}
	klog.InfoS("Loaded health info from the disk", "health", health)

	if isCurr, err := containsCurrentBootID(health.BootID); err != nil {
		return err
	} else if isCurr {
		klog.InfoS("Health file contains current boot - skipping backup")
		return nil
	}

	if !health.IsHealthy() {
		klog.InfoS("System was not healthy - skipping backup")
		return nil
	}

	dataManager, err := data.NewManager(config.BackupsDir)
	if err != nil {
		return err
	}

	existingBackups, err := dataManager.GetBackupList()
	if err != nil {
		return err
	}

	// get list of already existing backups for deployment ID persisted in health file
	// after creating backup, the list will be used to remove older backups
	// (so only the most recent one for specific deployment is kept)
	backupsForDeployment := getExistingBackupsForTheDeployment(existingBackups, health.DeploymentID)

	newBackupName := health.BackupName()
	if backupAlreadyExists(backupsForDeployment, newBackupName) {
		klog.InfoS("Backup already exists", "name", newBackupName)
		return nil
	}

	if err := dataManager.Backup(newBackupName); err != nil {
		return err
	}

	removeOldBackups(dataManager, backupsForDeployment)

	return nil
}

func containsCurrentBootID(id string) (bool, error) {
	path := "/proc/sys/kernel/random/boot_id"
	content, err := os.ReadFile(path)
	if err != nil {
		klog.ErrorS(err, "Failed to read file", "path", path)
		return false, fmt.Errorf("reading file %s failed: %w", path, err)
	}
	currentBootID := strings.ReplaceAll(strings.TrimSpace(string(content)), "-", "")
	klog.InfoS("Comparing boot IDs", "current", currentBootID, "toCompare", id)
	return id == currentBootID, nil
}

func getHealthInfo() (*HealthInfo, error) {
	path := "/var/lib/microshift-backups/health.json"
	if exists, err := util.PathExists(path); err != nil {
		return nil, err
	} else if !exists {
		return nil, errHealthFileDoesNotExist
	}

	content, err := os.ReadFile(path)
	if err != nil {
		klog.ErrorS(err, "Failed to read file", "path", path)
		return nil, err
	}

	health := &HealthInfo{}
	if err := json.Unmarshal(content, &health); err != nil {
		klog.ErrorS(err, "Failed to unmarshal file to json", "content", string(content))
		return nil, err
	}
	return health, nil
}

func getExistingBackupsForTheDeployment(existingBackups []data.BackupName, deployID string) []data.BackupName {
	existingDeploymentBackups := make([]data.BackupName, 0)

	for _, existingBackup := range existingBackups {
		if strings.HasPrefix(string(existingBackup), deployID) {
			existingDeploymentBackups = append(existingDeploymentBackups, existingBackup)
		}
	}

	return existingDeploymentBackups
}

func backupAlreadyExists(existingBackups []data.BackupName, name data.BackupName) bool {
	for _, backup := range existingBackups {
		if backup == name {
			return true
		}
	}
	return false
}

func removeOldBackups(dataManager data.Manager, backups []data.BackupName) {
	for _, b := range backups {
		klog.InfoS("Removing older backup", "name", b)
		if err := dataManager.RemoveBackup(b); err != nil {
			klog.ErrorS(err, "Failed to remove backup", "name", b)
		}
	}
}
