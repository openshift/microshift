package prerun

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
		if err := pr.backup(health); err != nil {
			return err
		}

		migrationNeeded, err := pr.checkVersions()
		if err != nil {
			return err
		}

		klog.InfoS("Version checks successful", "is-migration-needed?", migrationNeeded)

		if migrationNeeded {
			_ = migrationNeeded
			// TODO: data migration
		}

		return nil
	}

	return pr.restore()
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

func (pr *PreRun) restore() error {
	// TODO: Check if containers are already running (i.e. microshift.service was restarted)?

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}
	klog.InfoS("Restoring data for current deployment", "deployID", currentDeploymentID)

	existingBackups, err := pr.dataManager.GetBackupList()
	if err != nil {
		return err
	}
	klog.InfoS("List of existing backups", "backups", existingBackups)
	backupsForDeployment := getExistingBackupsForTheDeployment(existingBackups, currentDeploymentID)

	if len(backupsForDeployment) == 0 {
		return fmt.Errorf("there is no backup to restore for current deployment (%s)", currentDeploymentID)
	}

	if len(backupsForDeployment) > 1 {
		// could happen during backing up when removing older backups failed
		klog.InfoS("TODO: more than 1 backup, need to pick most recent one")
	}

	return pr.dataManager.Restore(backupsForDeployment[0])
}

// checkVersions compares version of data and executable
//
// It returns true if migration should be performed.
// It returns non-nil error if difference between versions is unsupported.
func (pr *PreRun) checkVersions() (bool, error) {
	execVer, err := getVersionOfExecutable()
	if err != nil {
		return false, err
	}

	dataVer, err := getVersionOfData()
	if err != nil {
		if errors.Is(err, errDataVersionDoesNotExist) {
			klog.InfoS("Version file of data does not exist - assuming data version is 4.13")
			// TODO: 4.13
			return true, nil
		}
		return false, err
	}

	klog.InfoS("Checking version difference between data and executable", "data", dataVer, "exec", execVer)

	if execVer == dataVer {
		return false, nil
	}

	if execVer.X != dataVer.X {
		return false, fmt.Errorf("major (X) versions are different: %d and %d", dataVer.X, execVer.X)
	}

	if execVer.Y < dataVer.Y {
		return false, fmt.Errorf("executable (%s) is older than existing data (%s): migrating data to older version is not supported", execVer.String(), dataVer.String())
	}

	return false, nil
}

func getCurrentDeploymentID() (string, error) {
	cmd := exec.Command("rpm-ostree", "status", "--jsonpath=$.deployments[0].id", "--booted")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command %s failed: %s", strings.TrimSpace(cmd.String()), strings.TrimSpace(stderr.String()))
	}

	ids := []string{}
	if err := json.Unmarshal(stdout.Bytes(), &ids); err != nil {
		return "", fmt.Errorf("unmarshalling '%s' to json failed: %w", strings.TrimSpace(stdout.String()), err)
	}

	if len(ids) != 1 {
		// this shouldn't happen if running on ostree system, but just in case
		klog.ErrorS(nil, "Unexpected amount of deployments in rpm-ostree output",
			"cmd", cmd.String(),
			"stdout", strings.TrimSpace(stdout.String()),
			"stderr", strings.TrimSpace(stderr.String()),
			"unmarshalledIDs", ids)
		return "", fmt.Errorf("rpm-ostree returned unexpected amount of deployment IDs: %d", len(ids))
	}

	return ids[0], nil
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
