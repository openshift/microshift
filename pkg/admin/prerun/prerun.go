package prerun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

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
		return pr.missingDataExistingHealth()
	}

	if !versionExists {
		// 5
		// Expected 4.13 upgrade flow: data exists, but neither the version nor health files

		// 6
		// Missing version means that version of the data is "4.13" as future
		// versions doesn't start without creating that file.
		//
		// Let's assume that existence of health.json is result of incomplete
		// manual intervention after system rolled back to 4.13
		// (i.e. backup was manually restored but health.json not deleted).

		klog.InfoS("Data exists, but version file is missing - assuming upgrade from 4.13")
		return pr.backup413()
	}

	// 7
	if !healthExists {
		// MicroShift might end up here if FIRST RUN of MicroShift gets interrupted
		// before green/red script manages to write the health file.
		//
		// Example scenarios:
		// - host rebooted before the end of greenboot's procedure
		// - test restarting MicroShift (e.g. to reload the config)
		//
		// For non-first boots the health file will exist, just contain slightly outdated boot ID
		// which might result in repeating the action (backup (which should already exist) or restore).
		//
		// Continuing start up seems to be the best course of action in this situation;
		// there is no health.json to steer the logic into backup or restore,
		// and deleting the files is too invasive.
		klog.InfoS("Health info is missing - continuing start up")
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

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}

	klog.InfoS("Found boot details",
		"health", health.Health,
		"deploymentID", health.DeploymentID,
		"previousBootID", health.BootID,
		"currentBootID", currentBootID,
		"currentDeploymentID", currentDeploymentID,
	)

	if currentBootID == health.BootID {
		// This might happen if microshift is (re)started after greenboot finishes running.
		// Green script will overwrite the health.json with
		// current boot's ID, deployment ID, and health.
		klog.InfoS("Skipping pre-run: Health file refers to the current boot ID")
		return nil
	}

	if health.IsHealthy() {
		klog.Info("Previous boot was healthy")
		if err := pr.backup(health); err != nil {
			return fmt.Errorf("failed to backup during pre-run: %w", err)
		}

		if health.DeploymentID != currentDeploymentID {
			klog.Info("Current and previously booted deployments are different")
			return pr.handleDeploymentSwitch(currentDeploymentID)
		}
		return nil
	}

	klog.Info("Previous boot was not healthy")
	if err = pr.handleUnhealthy(health); err != nil {
		return fmt.Errorf("failed to handle unhealthy data during pre-run: %w", err)
	}

	return nil
}

func (pr *PreRun) backup413() error {
	backupName := data.BackupName("4.13")

	if err := pr.dataManager.Backup(backupName); err != nil {
		return fmt.Errorf("failed to create new backup %q: %w", backupName, err)
	}

	return nil
}

// missingDataExistingHealth handles situation where MicroShift's data doesn't exist
// but health file is present.
//
// Depending on health persisted in health file it might:
//   - try to restore a backup for current deployment (if exists), or
//   - proceed with fresh start if "healthy" was persisted (nothing to back up)
//     or backup does not exists (nothing to restore)
func (pr *PreRun) missingDataExistingHealth() error {
	health, err := getHealthInfo()
	if err != nil {
		return fmt.Errorf("failed to determine the current system health: %w", err)
	}

	klog.InfoS("MicroShift data doesn't exist, but health info exists",
		"health", health.Health,
		"deploymentID", health.DeploymentID,
		"previousBootID", health.BootID,
	)

	if health.IsHealthy() {
		klog.InfoS("'Healthy' is persisted, but there's nothing to back up - continuing start up")
		return nil
	}

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}
	backup, err := pr.getBackupToRestore(currentDeploymentID)
	if err != nil {
		return err
	}
	if backup == "" {
		klog.InfoS("There is no backup to restore - continuing start up")
		return nil
	}

	if err := pr.dataManager.Restore(backup); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
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
	allBackupsForDeployment := getBackupsForTheDeployment(existingBackups, health.DeploymentID)
	healthyBackupsForDeployment := getOnlyHealthyBackups(allBackupsForDeployment)

	if backupAlreadyExists(healthyBackupsForDeployment, newBackupName) {
		klog.InfoS("Skipping backup: Backup already exists",
			"deploymentID", health.DeploymentID,
			"name", newBackupName,
		)
		return nil
	}

	if err := pr.dataManager.Backup(newBackupName); err != nil {
		return fmt.Errorf("failed to create new backup %q: %w", newBackupName, err)
	}

	// if making a new backup, remove all old backups for the deployment
	// including unhealthy ones
	pr.removeOldBackups(allBackupsForDeployment)

	klog.InfoS("Finished backup",
		"deploymentID", health.DeploymentID,
		"destination", newBackupName,
	)
	return nil
}

func (pr *PreRun) getBackupToRestore(deploymentID string) (data.BackupName, error) {
	existingBackups, err := pr.dataManager.GetBackupList()
	if err != nil {
		return "", fmt.Errorf("failed to get the existing backups: %w", err)
	}
	klog.InfoS("Found existing backups", "backups", existingBackups)

	allBackupsForDeployment := getBackupsForTheDeployment(existingBackups, deploymentID)
	healthyBackupsForDeployment := getOnlyHealthyBackups(allBackupsForDeployment)

	if len(healthyBackupsForDeployment) == 0 {
		return "", nil
	}

	if len(healthyBackupsForDeployment) > 1 {
		// could happen during backing up when removing older backups failed
		klog.InfoS("TODO: more than 1 backup, need to pick most recent one")
	}

	return healthyBackupsForDeployment[0], nil
}

func (pr *PreRun) handleUnhealthy(health *HealthInfo) error {
	// TODO: Check if containers are already running (i.e. microshift.service was restarted)?
	klog.Info("Handling previously unhealthy system")

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}

	backup, err := pr.getBackupToRestore(currentDeploymentID)
	if err != nil {
		return err
	}
	if backup != "" {
		err = pr.dataManager.Restore(backup)
		if err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
		klog.Info("Finished handling unhealthy system")
		return nil
	}

	klog.InfoS("There is no backup to restore for current deployment - trying to restore backup for rollback deployment")
	rollbackDeployID, err := getRollbackDeploymentID()
	if err != nil {
		return err
	}

	if rollbackDeployID == "" {
		// No backup for current deployment and there is no rollback deployment.
		// This could be a unhealthy system that was manually rebooted to
		// remediate the situation - let's not interfere: no backup, no restore, just proceed.
		klog.InfoS("System has no rollback but health.json suggests system was rebooted - skipping prerun")
		return nil
	}

	klog.InfoS("Obtained rollback deployment",
		"rollback-deployment-id", rollbackDeployID,
		"current-deployment-id", currentDeploymentID,
		"health-deployment-id", health.DeploymentID)

	if health.DeploymentID == rollbackDeployID {
		return fmt.Errorf("deployment ID stored in health.json is the same as rollback's" +
			" - MicroShift should not be updated from unhealthy system")
	}

	if health.DeploymentID == currentDeploymentID {
		rollbackBackup, err := pr.getBackupToRestore(rollbackDeployID)
		if err != nil {
			return err
		}
		if rollbackBackup == "" {
			// This could happen if current deployment is unhealthy and rollback didn't run MicroShift
			klog.InfoS("There is no backup for rollback deployment as well - removing existing data for clean start")
			return pr.dataManager.RemoveData()
		}

		// There is no backup for current deployment, but there is a backup for the rollback.
		// Let's restore it and act like it's first boot of newly staged deployment
		klog.InfoS("Restoring backup for a rollback deployment to perform migration and try starting again",
			"backup-name", rollbackBackup)
		if err := pr.dataManager.Restore(rollbackBackup); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
		return nil
	}

	// DeployID in health.json is neither booted nor rollback deployment,
	// so current deployment was staged over deployment without MicroShift
	// but MicroShift data exists (created by another deployment that rolled back).
	klog.InfoS("Deployment in health metadata is neither currently booted nor rollback deployment - backing up, then removing existing data for clean start")
	if err := pr.backup(health); err != nil {
		klog.ErrorS(err, "Failed to backup data of unhealthy system - ignoring")
	}
	return pr.dataManager.RemoveData()
}

func (pr *PreRun) handleDeploymentSwitch(currentDeploymentID string) error {
	// System booted into different deployment
	// It might be a rollback (restore existing backup), or
	// it might be a newly staged one (continue start up)
	existingBackups, err := pr.dataManager.GetBackupList()
	if err != nil {
		return fmt.Errorf("failed to determine the existing backups: %w", err)
	}
	backupsForDeployment := getBackupsForTheDeployment(existingBackups, currentDeploymentID)

	if len(backupsForDeployment) > 0 {
		klog.Info("Backup exists for current deployment - restoring")

		err = pr.dataManager.Restore(backupsForDeployment[0])
		if err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}

		klog.Info("Restored existing backup for current deployment")
	} else {
		klog.Info("There is no backup for current deployment - continuing start up")
	}

	return nil
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

func filterBackups(backups []data.BackupName, predicate func(data.BackupName) bool) []data.BackupName {
	out := make([]data.BackupName, 0, len(backups))
	for _, backup := range backups {
		if predicate(backup) {
			out = append(out, backup)
		}
	}
	return out
}

func getOnlyHealthyBackups(backups []data.BackupName) []data.BackupName {
	return filterBackups(backups, func(bn data.BackupName) bool {
		return !strings.HasSuffix(string(bn), "unhealthy")
	})
}

func getBackupsForTheDeployment(backups []data.BackupName, deployID string) []data.BackupName {
	return filterBackups(backups, func(bn data.BackupName) bool {
		return strings.HasPrefix(string(bn), deployID)
	})
}

func backupAlreadyExists(existingBackups []data.BackupName, name data.BackupName) bool {
	for _, backup := range existingBackups {
		if backup == name {
			return true
		}
	}
	return false
}

func getRollbackDeploymentID() (string, error) {
	cmd := exec.Command("rpm-ostree", "status", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%q failed: %s: %w", cmd, stderr.String(), err)
	}

	type deploy struct {
		ID     string `json:"id"`
		Booted bool   `json:"booted"`
	}
	type statusOutput struct {
		Deployments []deploy `json:"deployments"`
	}

	var status statusOutput
	if err := json.Unmarshal(stdout.Bytes(), &status); err != nil {
		return "", fmt.Errorf("failed to unmarshal %q: %w", cmd, err)
	}

	if len(status.Deployments) == 0 {
		return "", fmt.Errorf("unexpected amount (0) of deployments from rpm-ostree status output")
	}

	if len(status.Deployments) == 1 {
		return "", nil
	}

	afterBooted := false
	for _, d := range status.Deployments {
		if afterBooted {
			return d.ID, nil
		}

		if d.Booted {
			afterBooted = true
		}
	}

	return "", fmt.Errorf("could not find rollback deployment in %v", status)
}
