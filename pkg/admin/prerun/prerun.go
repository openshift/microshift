package prerun

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/openshift/microshift/pkg/admin/data"
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

type PreRun struct {
	dataManager data.Manager
}

func New(dataManager data.Manager) *PreRun {
	return &PreRun{
		dataManager: dataManager,
	}
}

func (pr *PreRun) Perform() error {
	health, err := getHealthInfo()
	if err != nil {
		if errors.Is(err, errHealthFileDoesNotExist) {
			klog.InfoS("Health file does not exist - skipping backup")
			return nil
		}
		return err
	}

	if isCurr, err := containsCurrentBootID(health.BootID); err != nil {
		return err
	} else if isCurr {
		// This might happen if microshift is (re)started after greenboot finishes running.
		// Green script will overwrite the health.json with
		// current boot's ID, deployment ID, and health.
		klog.InfoS("Health file contains current boot - skipping pre-run")
		return nil
	}

	klog.InfoS("Previous boot", "health", health.Health, "deploymentID", health.DeploymentID, "bootID", health.BootID)

	if health.IsHealthy() {
		return pr.backup(health)
	}

	return nil
}

func (pr *PreRun) backup(health *HealthInfo) error {
	klog.InfoS("Backing up the data for deployment", "deployment", health.DeploymentID)

	existingBackups, err := pr.dataManager.GetBackupList()
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

	if err := pr.dataManager.Backup(newBackupName); err != nil {
		return err
	}

	pr.removeOldBackups(backupsForDeployment)

	return nil
}

func (pr *PreRun) removeOldBackups(backups []data.BackupName) {
	for _, b := range backups {
		klog.InfoS("Removing older backup", "name", b)
		if err := pr.dataManager.RemoveBackup(b); err != nil {
			klog.ErrorS(err, "Failed to remove backup", "name", b)
		}
	}
}

func containsCurrentBootID(id string) (bool, error) {
	path := "/proc/sys/kernel/random/boot_id"
	content, err := os.ReadFile(path)
	if err != nil {
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
		return nil, fmt.Errorf("reading file %s failed: %w", path, err)
	}

	health := &HealthInfo{}
	if err := json.Unmarshal(content, &health); err != nil {
		return nil, fmt.Errorf("unmarshalling '%s' failed: %w", strings.TrimSpace(string(content)), err)
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
