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
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

var (
	healthFilepath            = "/var/lib/microshift-backups/health.json"
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
	config      *config.Config
}

func New(dataManager data.Manager, config *config.Config) *PreRun {
	return &PreRun{
		dataManager: dataManager,
		config:      config,
	}
}

func (pr *PreRun) Perform() error {
	klog.InfoS("Starting pre-run")
	defer klog.InfoS("Pre-run complete")

	if isOstree, err := util.PathExists("/run/ostree-booted"); err != nil {
		return fmt.Errorf("failed to check if system is ostree: %w", err)
	} else if !isOstree {
		klog.InfoS("System is not OSTree-based")
		return nil
	}

	dataExists, err := util.PathExistsAndIsNotEmpty(config.DataDir, ".nodename")
	if err != nil {
		return fmt.Errorf("failed to check if data directory already exists: %w", err)
	}

	versionExists, err := util.PathExistsAndIsNotEmpty(versionFilePath)
	if err != nil {
		return fmt.Errorf("checking if version metadata exists failed: %w", err)
	}

	healthExists, err := util.PathExistsAndIsNotEmpty(healthFilepath)
	if err != nil {
		return fmt.Errorf("failed to check if health file already exists: %w", err)
	}

	klog.InfoS("Existence of important paths",
		"data-exists?", dataExists,
		"version-exists?", versionExists,
		"health-exists?", healthExists,
	)

	/*
		| id  | data | version | health | comment                                                                                                     |
		| --- | ---- | ------- | ------ | ----------------------------------------------------------------------------------------------------------- |
		| 1   | 0    | 0       | 0      | first, clean start of MicroShift                                                                            |
		| 2   | 0    | 0       | 1      | data removed manually, but health/backups kept                                                              |
		| 3   | 0    | 1       | 0      | impossible to detect right now [0]                                                                          |
		| 4   | 0    | 1       | 1      | impossible to detect right now [0]                                                                          |
		| 5   | 1    | 0       | 0      | upgrade from 4.13                                                                                           |
		| 6   | 1    | 0       | 1      | first boot failed very early, or upgrade from 4.13 failed (e.g. healthcheck didn't see service being ready) |
		| 7   | 1    | 1       | 0      | first start, rebooted before green/red scripts managed to write health info                                 |
		| 8   | 1    | 1       | 1      | regular                                                                                                     |

		[0] it would need a comprehensive check if data exists, not just existence of /var/lib/microshift
	*/

	if !dataExists {
		// Implies !versionExists

		// 1
		if !healthExists {
			klog.InfoS("Neither data dir nor health file exist - assuming first start of MicroShift")
			return nil
		}

		// 2
		// TODO: Health needs to be extended into a history of boots and their health
		// so prerun can look into the past and decide if a backup should be restored
		// for currently running deployment
		return fmt.Errorf("TODO: data is missing, but health file exists")
	}

	if !versionExists {
		// 5
		if !healthExists {
			klog.InfoS("Data dir exists, but health and version files are missing: assuming upgrade from 4.13")
			return pr.upgradeFrom413()
		}

		// 6
		// TODO: Check if health data is for previous boot (according to journalctl --list-boots)
		// This could happen if system rolled back to 4.13, backup was manually restored to attempt upgrade again,
		// but health file not deleted leaving stale data behind.
		return fmt.Errorf("TODO: health file exist, but version metadata does not")
	}

	// 7
	if !healthExists {
		// MicroShift might end up here if first run of MicroShift gets interrupted
		// before green/red script manages to write the health file.
		// Examples include:
		// - host reboot
		// - if e2e test restarts MicroShift (e.g. to reload the config) or reboots the host
		//   - due to the way microshift-etcd now runs, `restart microshift` causes a restart of both
		//     microshift-etcd and microshift - if m-etcd restarts before microshift,
		//     microshift will restart it self as a way to start m-etcd again
		klog.InfoS("TODO: Version file exists, but health info is missing")
		return nil
	}

	// 8 - regular flow
	return pr.regularPrerun()
}

// regularPrerun performs actions in prerun flow that is most expected in day to day usage
// (i.e. data, version metadata, and health information exist)
func (pr *PreRun) regularPrerun() error {
	health, err := getHealthInfo()
	if err != nil {
		return fmt.Errorf("failed to determine the current system health: %w", err)
	}

	currentBootID, err := getCurrentBootID()
	if err != nil {
		return fmt.Errorf("failed to determine the current boot ID: %w", err)
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
			return fmt.Errorf("failed to backup during pre-run: %w", err)
		}

		migrationNeeded, err := pr.checkVersions()
		if err != nil {
			return fmt.Errorf("failed version checks: %w", err)
		}

		klog.InfoS("Completed version checks", "is-migration-needed?", migrationNeeded)

		if migrationNeeded {
			_ = migrationNeeded
			stop, err := runMinimalMicroshift(pr.config)
			if err != nil {
				return fmt.Errorf("minimal MicroShift run failed to be ready: %w", err)
			}
			defer stop()
			// TODO: data migration

			if err := writeExecVersionToData(); err != nil {
				return fmt.Errorf("failed to write MicroShift version to data directory: %w", err)
			}
		}
	} else {
		klog.Info("Previous boot was not healthy")
		if err = pr.restore(); err != nil {
			return fmt.Errorf("failed to restore during pre-run: %w", err)
		}
	}

	return nil
}

func (pr *PreRun) upgradeFrom413() error {
	backupName := data.BackupName("4.13")

	if err := pr.dataManager.Backup(backupName); err != nil {
		return fmt.Errorf("failed to create new backup %q: %w", backupName, err)
	}

	stop, err := runMinimalMicroshift(pr.config)
	if err != nil {
		return fmt.Errorf("minimal MicroShift run failed to be ready: %w", err)
	}
	defer stop()
	// TODO: data migration

	if err := writeExecVersionToData(); err != nil {
		return fmt.Errorf("failed to write MicroShift version to data directory: %w", err)
	}

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
		return fmt.Errorf("failed to determine the existing backups: %w", err)
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
		return fmt.Errorf("failed to create new backup %q: %w", newBackupName, err)
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
		return fmt.Errorf("failed to determine the existing backups: %w", err)
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
		return fmt.Errorf("failed to restore backup: %w", err)
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
		return false, fmt.Errorf("failed to determine the active version of the MicroShift: %w", err)
	}

	dataVer, err := getVersionOfData()
	if err != nil {
		return false, fmt.Errorf("failed to determine the version of the existing data: %w", err)
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
		return "", fmt.Errorf("failed to determine the rpm-ostree deployment id, command %q failed: %s: %w", strings.TrimSpace(cmd.String()), strings.TrimSpace(stderr.String()), err)
	}

	ids := []string{}
	if err := json.Unmarshal(stdout.Bytes(), &ids); err != nil {
		return "", fmt.Errorf("failed to determine the rpm-ostree deployment id from %q: %w", strings.TrimSpace(stdout.String()), err)
	}

	if len(ids) != 1 {
		// this shouldn't happen if running on ostree system, but just in case
		klog.ErrorS(nil, "Unexpected number of deployments in rpm-ostree output",
			"cmd", cmd.String(),
			"stdout", strings.TrimSpace(stdout.String()),
			"stderr", strings.TrimSpace(stderr.String()),
			"unmarshalledIDs", ids)
		return "", fmt.Errorf("expected 1 deployment ID, rpm-ostree returned %d", len(ids))
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
		return "", fmt.Errorf("failed to determine boot ID from %q: %w", path, err)
	}
	return strings.ReplaceAll(strings.TrimSpace(string(content)), "-", ""), nil
}

func getHealthInfo() (*HealthInfo, error) {
	if exists, err := util.PathExistsAndIsNotEmpty(healthFilepath); err != nil {
		return nil, err
	} else if !exists {
		return nil, errHealthFileDoesNotExist
	}

	content, err := os.ReadFile(healthFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read health data from %q: %w", healthFilepath, err)
	}

	health := &HealthInfo{}
	if err := json.Unmarshal(content, &health); err != nil {
		return nil, fmt.Errorf("failed to parse health data %q: %w", strings.TrimSpace(string(content)), err)
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
