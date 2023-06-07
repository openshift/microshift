package prerun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/util"
	"github.com/pkg/errors"
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
	klog.InfoS("Starting pre-run")

	health, err := getHealthInfo()
	if err != nil {
		if errors.Is(err, errHealthFileDoesNotExist) {
			klog.InfoS("Skipping pre-run: Health status file does not exist")
			return nil
		}
		return errors.Wrap(err, "Failed to determine the current system health")
	}

	currentBootID, err := getCurrentBootID()
	if err != nil {
		return errors.Wrap(err, "Failed to determine the current boot ID")
	}

	klog.InfoS("Found boot details",
		"health", health.Health,
		"deploymentID", health.DeploymentID,
		"previousBootID", health.BootID,
		"currentBootID", currentBootID,
	)

	if currentBootID == health.BootID {
		// This might happen if microshift is (re)started after greenboot finishes running.
		// Green script will overwrite the health.json with
		// current boot's ID, deployment ID, and health.
		klog.InfoS("Skipping pre-run: Health file refers to the current boot ID")
		return nil
	}

	// TODO: We may end up needing to split this if statement into
	// functions, but for now let's just tell the linter not to apply
	// the rule.
	//
	//nolint:nestif
	if health.IsHealthy() {
		klog.Info("Previous boot was healthy")
		if err := pr.backup(health); err != nil {
			return errors.Wrap(err, "Failed to backup during pre-run")
		}

		migrationNeeded, err := pr.checkVersions()
		if err != nil {
			return errors.Wrap(err, "Failed version checks")
		}

		klog.InfoS("Completed version checks", "is-migration-needed?", migrationNeeded)

		if migrationNeeded {
			_ = migrationNeeded
			// TODO: data migration
		}

		return nil
	} else {
		klog.Info("Previous boot was not healthy")
		if err = pr.restore(); err != nil {
			return errors.Wrap(err, "Failed to restore during pre-run")
		}
	}

	klog.InfoS("Finished pre-run")
	return nil
}

func (pr *PreRun) backup(health *HealthInfo) error {
	newBackupName := health.BackupName()
	klog.InfoS("Preparing to backup",
		"deploymentID", health.DeploymentID,
		"name", newBackupName,
	)

	existingBackups, err := pr.dataManager.GetBackupList()
	if err != nil {
		return errors.Wrap(err, "Failed to determine the existing backups")
	}

	// get list of already existing backups for deployment ID persisted in health file
	// after creating backup, the list will be used to remove older backups
	// (so only the most recent one for specific deployment is kept)
	backupsForDeployment := getExistingBackupsForTheDeployment(existingBackups, health.DeploymentID)

	if backupAlreadyExists(backupsForDeployment, newBackupName) {
		klog.InfoS("Skipping backup: Backup already exists",
			"deploymentID", health.DeploymentID,
			"name", newBackupName,
		)
		return nil
	}

	if err := pr.dataManager.Backup(newBackupName); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to create new backup %q", newBackupName))
	}

	pr.removeOldBackups(backupsForDeployment)

	klog.InfoS("Finished backup",
		"deploymentID", health.DeploymentID,
		"destination", newBackupName,
	)
	return nil
}

func (pr *PreRun) restore() error {
	// TODO: Check if containers are already running (i.e. microshift.service was restarted)?
	klog.Info("Preparing to restore")

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}

	existingBackups, err := pr.dataManager.GetBackupList()
	if err != nil {
		return errors.Wrap(err, "Failed to determine the existing backups")
	}
	klog.InfoS("Found existing backups",
		"currentDeploymentID", currentDeploymentID,
		"backups", existingBackups,
	)
	backupsForDeployment := getExistingBackupsForTheDeployment(existingBackups, currentDeploymentID)

	if len(backupsForDeployment) == 0 {
		return fmt.Errorf("there is no backup to restore for current deployment %q", currentDeploymentID)
	}

	if len(backupsForDeployment) > 1 {
		// could happen during backing up when removing older backups failed
		klog.InfoS("TODO: more than 1 backup, need to pick most recent one")
	}

	err = pr.dataManager.Restore(backupsForDeployment[0])
	if err != nil {
		return errors.Wrap(err, "Failed to restore backup")
	}

	klog.Info("Finished restore")
	return nil
}

// checkVersions compares version of data and executable
//
// It returns true if migration should be performed.
// It returns non-nil error if difference between versions is unsupported.
func (pr *PreRun) checkVersions() (bool, error) {
	klog.Info("Starting version checks")
	execVer, err := getVersionOfExecutable()
	if err != nil {
		return false, errors.Wrap(err, "Failed to determine the active version of the MicroShift")
	}

	dataVer, err := getVersionOfData()
	if err != nil {
		if errors.Is(err, errDataVersionDoesNotExist) {
			klog.InfoS("Version file of data does not exist - assuming data version is 4.13")
			// TODO: 4.13
			return true, nil
		}
		return false, errors.Wrap(err, "Failed to determine the version of the existing data")
	}

	klog.InfoS("Comparing versions", "data", dataVer, "active", execVer)

	return checkVersionDiff(execVer, dataVer)
}

func getCurrentDeploymentID() (string, error) {
	cmd := exec.Command("rpm-ostree", "status", "--jsonpath=$.deployments[0].id", "--booted")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Failed to determine the rpm-ostree deployment id, command %q failed: %s", strings.TrimSpace(cmd.String()), strings.TrimSpace(stderr.String()))
	}

	ids := []string{}
	if err := json.Unmarshal(stdout.Bytes(), &ids); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to determine the rpm-ostree deployment id from %q", strings.TrimSpace(stdout.String())))
	}

	if len(ids) != 1 {
		// this shouldn't happen if running on ostree system, but just in case
		klog.ErrorS(nil, "Unexpected number of deployments in rpm-ostree output",
			"cmd", cmd.String(),
			"stdout", strings.TrimSpace(stdout.String()),
			"stderr", strings.TrimSpace(stderr.String()),
			"unmarshalledIDs", ids)
		return "", fmt.Errorf("Expected 1 deployment ID, rpm-ostree returned %d", len(ids))
	}

	return ids[0], nil
}

func (pr *PreRun) removeOldBackups(backups []data.BackupName) {
	klog.Info("Preparing to prune backups")
	for _, b := range backups {
		if err := pr.dataManager.RemoveBackup(b); err != nil {
			klog.ErrorS(err, "Failed to remove backup, ignoring", "name", b)
		}
	}
	klog.Info("Finished pruning backups")
}

func getCurrentBootID() (string, error) {
	path := "/proc/sys/kernel/random/boot_id"
	content, err := os.ReadFile(path)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to determine boot ID from %s", path))
	}
	return strings.ReplaceAll(strings.TrimSpace(string(content)), "-", ""), nil
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
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to read health data from %q", path))
	}

	health := &HealthInfo{}
	if err := json.Unmarshal(content, &health); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to parse health data %q", strings.TrimSpace(string(content))))
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
