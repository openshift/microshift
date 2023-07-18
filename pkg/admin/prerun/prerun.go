package prerun

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/multilogger"
)

var (
	preRunLogFilepath = filepath.Join(config.BackupsDir, "action_log.txt")
	fileKlog          = multilogger.NewFileLoggerWithFallback(preRunLogFilepath)
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
	fileKlog.InfoS("Starting pre-run")
	defer fileKlog.InfoS("Pre-run complete")

	if isOstree, err := util.PathExists("/run/ostree-booted"); err != nil {
		return fmt.Errorf("failed to check if system is ostree: %w", err)
	} else if !isOstree {
		fileKlog.InfoS("System is not OSTree-based")
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

	fileKlog.InfoS("Existence of important paths",
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
			fileKlog.InfoS("Neither data dir nor health file exist - assuming first start of MicroShift")
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

		fileKlog.InfoS("Data exists, but version file is missing - assuming upgrade from 4.13")
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
		fileKlog.InfoS("Health info is missing - continuing start up")
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

	fileKlog.InfoS("Found boot details",
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
		fileKlog.InfoS("Skipping pre-run: Health file refers to the current boot ID")
		return nil
	}

	if health.IsHealthy() {
		fileKlog.InfoS("Previous boot was healthy")
		if err := pr.backup(health); err != nil {
			return fmt.Errorf("failed to backup during pre-run: %w", err)
		}

		if health.DeploymentID != currentDeploymentID {
			fileKlog.InfoS("Current and previously booted deployments are different")
			return pr.handleDeploymentSwitch(currentDeploymentID)
		}
		return nil
	}

	fileKlog.InfoS("Previous boot was not healthy")
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

	fileKlog.InfoS("MicroShift data doesn't exist, but health info exists",
		"health", health.Health,
		"deploymentID", health.DeploymentID,
		"previousBootID", health.BootID,
	)

	if health.IsHealthy() {
		fileKlog.InfoS("'Healthy' is persisted, but there's nothing to back up - continuing start up")
		return nil
	}

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}
	existingBackups, err := getBackups(pr.dataManager)
	if err != nil {
		return err
	}
	backup := existingBackups.getForDeployment(currentDeploymentID).getOnlyHealthyBackups().getOneOrNone()
	if backup == "" {
		fileKlog.InfoS("There is no backup to restore - continuing start up")
		return nil
	}

	if err := pr.dataManager.Restore(backup); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}
	return nil
}

func (pr *PreRun) backup(health *HealthInfo) error {
	newBackupName := health.BackupName()
	fileKlog.InfoS("Preparing to backup",
		"deploymentID", health.DeploymentID,
		"name", newBackupName,
	)

	existingBackups, err := getBackups(pr.dataManager)
	if err != nil {
		return err
	}

	if existingBackups.has(newBackupName) {
		fileKlog.InfoS("Skipping backup: Backup already exists",
			"deploymentID", health.DeploymentID,
			"name", newBackupName,
		)
		return nil
	}

	if err := pr.dataManager.Backup(newBackupName); err != nil {
		return fmt.Errorf("failed to create new backup %q: %w", newBackupName, err)
	}

	// after making a new backup, remove all old backups for the deployment
	// including unhealthy ones
	existingBackups.getForDeployment(health.DeploymentID).removeAll(pr.dataManager)
	if err := pr.removeBackupsWithoutExistingDeployments(existingBackups); err != nil {
		fileKlog.ErrorS(err, "Failed to remove backups belonging to no longer existing deployments - ignoring")
	}

	fileKlog.InfoS("Finished backup",
		"deploymentID", health.DeploymentID,
		"destination", newBackupName,
	)
	return nil
}

func (pr *PreRun) handleUnhealthy(health *HealthInfo) error {
	// TODO: Check if containers are already running (i.e. microshift.service was restarted)?
	fileKlog.InfoS("Handling previously unhealthy system")

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}

	existingBackups, err := getBackups(pr.dataManager)
	if err != nil {
		return err
	}

	backup := existingBackups.getForDeployment(currentDeploymentID).getOnlyHealthyBackups().getOneOrNone()
	if backup != "" {
		err = pr.dataManager.Restore(backup)
		if err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
		fileKlog.InfoS("Finished handling unhealthy system")
		return nil
	}

	fileKlog.InfoS("There is no backup to restore for current deployment - trying to restore backup for rollback deployment")
	rollbackDeployID, err := getRollbackDeploymentID()
	if err != nil {
		return err
	}

	if rollbackDeployID == "" {
		// No backup for current deployment and there is no rollback deployment.
		// This could be a unhealthy system that was manually rebooted to
		// remediate the situation - let's not interfere: no backup, no restore, just proceed.
		fileKlog.InfoS("System has no rollback but health.json suggests system was rebooted - skipping prerun")
		return nil
	}

	fileKlog.InfoS("Obtained rollback deployment",
		"rollback-deployment-id", rollbackDeployID,
		"current-deployment-id", currentDeploymentID,
		"health-deployment-id", health.DeploymentID)

	if health.DeploymentID == rollbackDeployID {
		return fmt.Errorf("deployment ID stored in health.json is the same as rollback's" +
			" - MicroShift should not be updated from unhealthy system")
	}

	if health.DeploymentID == currentDeploymentID {
		rollbackBackup := existingBackups.getForDeployment(rollbackDeployID).getOnlyHealthyBackups().getOneOrNone()
		if err != nil {
			return err
		}
		if rollbackBackup == "" {
			// This could happen if current deployment is unhealthy and rollback didn't run MicroShift
			fileKlog.InfoS("There is no backup for rollback deployment as well - removing existing data for clean start")
			return pr.dataManager.RemoveData()
		}

		// There is no backup for current deployment, but there is a backup for the rollback.
		// Let's restore it and act like it's first boot of newly staged deployment
		fileKlog.InfoS("Restoring backup for a rollback deployment to perform migration and try starting again",
			"backup-name", rollbackBackup)
		if err := pr.dataManager.Restore(rollbackBackup); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
		return nil
	}

	// DeployID in health.json is neither booted nor rollback deployment,
	// so current deployment was staged over deployment without MicroShift
	// but MicroShift data exists (created by another deployment that rolled back).
	fileKlog.InfoS("Deployment in health metadata is neither currently booted nor rollback deployment - backing up, then removing existing data for clean start")
	if err := pr.backup(health); err != nil {
		fileKlog.ErrorS(err, "Failed to backup data of unhealthy system - ignoring")
	}
	return pr.dataManager.RemoveData()
}

func (pr *PreRun) handleDeploymentSwitch(currentDeploymentID string) error {
	// System booted into different deployment
	// It might be a rollback (restore existing backup), or
	// it might be a newly staged one (continue start up)

	existingBackups, err := getBackups(pr.dataManager)
	if err != nil {
		return err
	}
	backup := existingBackups.getForDeployment(currentDeploymentID).getOnlyHealthyBackups().getOneOrNone()

	if backup != "" {
		fileKlog.InfoS("Backup exists for current deployment - restoring")

		err = pr.dataManager.Restore(backup)
		if err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}

		fileKlog.InfoS("Restored existing backup for current deployment")
	} else {
		fileKlog.InfoS("There is no backup for current deployment - continuing start up")
	}

	return nil
}

func (pr *PreRun) removeBackupsWithoutExistingDeployments(backups Backups) error {
	deployments, err := getAllDeploymentIDs()
	if err != nil {
		return err
	}

	toRemove := backups.getDangling(deployments)
	fileKlog.InfoS("Removing backups for no longer existing deployments",
		"backups-to-remove", toRemove)
	toRemove.removeAll(pr.dataManager)

	return nil
}
