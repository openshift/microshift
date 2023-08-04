package prerun

import (
	"fmt"
	"strings"

	datadir "github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

type DataManager struct {
	dataManager datadir.Manager
}

func NewDataManager(dataManager datadir.Manager) *DataManager {
	return &DataManager{
		dataManager,
	}
}

func (dm *DataManager) Run() error {
	if isOstree, err := util.PathExists("/run/ostree-booted"); err != nil {
		return fmt.Errorf("failed to check if system is ostree: %w", err)
	} else if !isOstree {
		klog.InfoS("System is not OSTree-based - skipping data management")
		return nil
	}

	handled, err := dm.handleSpecialCases()
	if err != nil {
		return fmt.Errorf("handling special cases failed: %w", err)
	}

	if handled {
		klog.InfoS("Data management handled by one of the special cases - not performing regular procedure")
		return nil
	}

	// 8 - regular flow
	return dm.regularPrerun()
}

func (dm *DataManager) handleSpecialCases() (bool, error) {
	klog.InfoS("Handling special cases of pre-run data management")
	defer klog.InfoS("Completed handling special cases of pre-run data management")

	dataExists, err := util.PathExistsAndIsNotEmpty(config.DataDir, ".nodename")
	if err != nil {
		return true, fmt.Errorf("failed to check if data directory already exists: %w", err)
	}

	versionExists, err := util.PathExistsAndIsNotEmpty(versionFilePath)
	if err != nil {
		return true, fmt.Errorf("checking if version metadata exists failed: %w", err)
	}

	healthExists, err := util.PathExistsAndIsNotEmpty(healthFilepath)
	if err != nil {
		return true, fmt.Errorf("failed to check if health file already exists: %w", err)
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
			klog.InfoS("Neither data dir nor health file exist (assuming first start of MicroShift)" +
				" - skipping data management")
			return true, nil
		}

		// 2
		return true, dm.missingDataExistingHealth()
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

		klog.InfoS("Data exists, but version file is missing (assuming data version is 4.13)" +
			" - backing up the data and continuing start up")
		return true, dm.backup413()
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
		klog.InfoS("Health info is missing - skipping data management and continuing start up")
		return true, nil
	}

	return false, nil
}

// regularPrerun performs actions in prerun flow that is most expected in day to day usage
// (i.e. data, version metadata, and health information exist)
func (dm *DataManager) regularPrerun() error {
	klog.InfoS("Starting pre-run data management")
	defer klog.InfoS("Completed pre-run data management")

	health, err := getHealthInfo()
	if err != nil {
		return fmt.Errorf("failed to get health info: %w", err)
	}

	currentBootID, err := getCurrentBootID()
	if err != nil {
		return fmt.Errorf("failed to determine the current boot ID: %w", err)
	}

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}

	klog.InfoS("Obtained health info and current boot details",
		"health.Health", health.Health,
		"health.DeploymentID", health.DeploymentID,
		"health.BootID", health.BootID,
		"currentBootID", currentBootID,
		"currentDeploymentID", currentDeploymentID,
	)

	if currentBootID == health.BootID {
		// This might happen if microshift is (re)started after greenboot finishes running.
		// Green script will overwrite the health.json with
		// current boot's ID, deployment ID, and health.
		klog.InfoS("Boot ID in health file matches current boot - skipping data management and continuing start up")
		return nil
	}

	if health.IsHealthy() {
		if err = dm.handleHealthy(health, currentDeploymentID); err != nil {
			return fmt.Errorf("failed to handle healthy system: %w", err)
		}

		return nil
	}

	if err = dm.handleUnhealthy(health); err != nil {
		return fmt.Errorf("failed to handle unhealthy system: %w", err)
	}

	return nil
}

func (dm *DataManager) backup413() error {
	klog.InfoS("Creating 4.13 backup")

	// If 4.13 backup already exists: remove old and make a new one.
	// As an alternative we might rename existing backup and add some suffix,
	// but 4.13 backups require manual cleanup.

	// If the backup exists, it might mean that:
	// - this is a subsequent boot after first boot of upgrade from 4.13 failed, or
	// - this is first upgrade boot, but there was already attempt to
	//   upgrade from 4.13 that left stale 4.13 backup (i.e. not cleaned by the admin)

	// This function is called when data exists, but version file does not.
	// It means version file wasn't created yet, meaning there was no attempt
	// yet to migrate the data (version file creation is before any of the
	// MicroShift components start), so it's not "corrupted" yet with
	// newer-version-artifacts.

	// Regardless which scenario it is (greenboot reboot after failed
	// upgrade attempt or another attempt to upgrade from 4.13), we can
	// assume that current data is the most up to date because failing to
	// upgrade from 4.13 should be investigated and problems addressed before
	// attempting again.

	// Assuming this data is the most up to date, we prefer it over
	// existing 4.13 backups.

	backupName := datadir.BackupName("4.13")
	if exists, err := dm.dataManager.BackupExists(backupName); err != nil {
		return fmt.Errorf("failed to check if '%q' backup exists: %w",
			backupName, err)
	} else if exists {
		klog.InfoS("Backup 4.13 already exists " +
			"- assuming current data is the most up to date one: " +
			"removing existing backup and creating a new one")
		if err := dm.dataManager.RemoveBackup(backupName); err != nil {
			return fmt.Errorf("failed to remove old 4.13 backup: %w", err)
		}
	}

	if err := dm.dataManager.Backup(backupName); err != nil {
		return fmt.Errorf("failed to create new 4.13 backup: %w", err)
	}

	klog.InfoS("Created 4.13 backup")
	return nil
}

// missingDataExistingHealth handles situation where MicroShift's data doesn't exist
// but health file is present.
//
// Depending on health persisted in health file it might:
//   - try to restore a backup for current deployment (if exists), or
//   - proceed with fresh start if "healthy" was persisted (nothing to back up)
//     or backup does not exists (nothing to restore)
func (dm *DataManager) missingDataExistingHealth() error {
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
		// Skipping data management and not attempting to restore a backup,
		// to be consistent with other areas: healthy - backup, unhealthy - restore.
		klog.InfoS("'Healthy' is persisted, but data does not exist (nothing to back up)" +
			" - skipping data management and continuing start up")
		return nil
	}

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}
	existingBackups, err := getBackups(dm.dataManager)
	if err != nil {
		return err
	}
	backup := existingBackups.getForDeployment(currentDeploymentID).getOnlyHealthyBackups().getOneOrNone()
	if backup == "" {
		klog.InfoS("No backup for current deployment exists (nothing to restore)" +
			" - skipping data management and continuing start up")
		return nil
	}

	if err := dm.restore(backup); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}
	return nil
}

func (dm *DataManager) backup(health *HealthInfo) error {
	klog.Info("Creating backup")

	existingBackups, err := getBackups(dm.dataManager)
	if err != nil {
		return err
	}

	newBackupName := health.BackupName()
	if existingBackups.has(newBackupName) {
		klog.InfoS("Backup already exists", "name", newBackupName)
		return nil
	}

	if err := dm.dataManager.Backup(newBackupName); err != nil {
		return fmt.Errorf("failed to create backup %q: %w", newBackupName, err)
	}

	// after making a new backup, remove all old backups for the deployment
	// including unhealthy ones
	existingBackups.getForDeployment(health.DeploymentID).removeAll(dm.dataManager)
	if err := dm.removeBackupsWithoutExistingDeployments(existingBackups); err != nil {
		klog.ErrorS(err, "Failed to remove backups belonging to no longer existing deployments - ignoring")
	}

	klog.Info("Created backup")
	return nil
}

func (dm *DataManager) handleHealthy(health *HealthInfo, currentDeploymentID string) error {
	klog.Info("Handling healthy system")
	defer klog.Info("Handled healthy system")

	if err := dm.backup(health); err != nil {
		return fmt.Errorf("failed to handle healthy system: %w", err)
	}

	if health.DeploymentID != currentDeploymentID {
		klog.InfoS("Current deployment is different from deploymentID from health file")
		return dm.handleDeploymentSwitch(currentDeploymentID)
	}

	return nil
}

func (dm *DataManager) handleUnhealthy(health *HealthInfo) error {
	klog.Info("Handling unhealthy system")
	defer klog.Info("Handled unhealthy system")

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}

	existingBackups, err := getBackups(dm.dataManager)
	if err != nil {
		return err
	}

	backup := existingBackups.getForDeployment(currentDeploymentID).getOnlyHealthyBackups().getOneOrNone()
	if backup != "" {
		err = dm.restore(backup)
		if err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
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
		rollbackBackup := existingBackups.getForDeployment(rollbackDeployID).getOnlyHealthyBackups().getOneOrNone()
		if err != nil {
			return err
		}
		if rollbackBackup == "" {
			// This could happen if current deployment is unhealthy and rollback didn't run MicroShift
			klog.InfoS("There is no backup for rollback deployment as well - removing existing data for clean start")
			return dm.dataManager.RemoveData()
		}

		// There is no backup for current deployment, but there is a backup for the rollback.
		// Let's restore it and act like it's first boot of newly staged deployment
		klog.InfoS("Restoring rollback deployment's backup to try starting from healthy data")
		if err := dm.restore(rollbackBackup); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
		return nil
	}

	// DeployID in health.json is neither booted nor rollback deployment,
	// so current deployment was staged over deployment without MicroShift
	// but MicroShift data exists (created by another deployment that rolled back).
	klog.InfoS("Deployment in health metadata is neither currently booted nor rollback deployment - backing up, then removing existing data for clean start")
	if err := dm.backup(health); err != nil {
		klog.ErrorS(err, "Failed to backup data of unhealthy system - ignoring")
	}
	return dm.dataManager.RemoveData()
}

func (dm *DataManager) handleDeploymentSwitch(currentDeploymentID string) error {
	// System booted into different deployment
	// It might be a rollback (restore existing backup), or
	// it might be a newly staged one (continue start up)

	klog.InfoS("Handling deployment switch")

	existingBackups, err := getBackups(dm.dataManager)
	if err != nil {
		return err
	}
	backup := existingBackups.getForDeployment(currentDeploymentID).getOnlyHealthyBackups().getOneOrNone()

	if backup != "" {
		klog.Info("Backup exists for current deployment - restoring")

		err = dm.restore(backup)
		if err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
	} else {
		klog.Info("No backup for current deployment to restore - continuing start up")
	}

	klog.Info("Handled deployment switch")
	return nil
}

func (dm *DataManager) removeBackupsWithoutExistingDeployments(backups Backups) error {
	deployments, err := getAllDeploymentIDs()
	if err != nil {
		return err
	}

	toRemove := backups.getDangling(deployments)
	klog.InfoS("Removing backups for no longer existing deployments",
		"backups-to-remove", toRemove)
	toRemove.removeAll(dm.dataManager)

	return nil
}

func (dm *DataManager) restore(backup datadir.BackupName) error {
	klog.InfoS("Restoring backup", "name", backup)

	versionFile, err := getVersionFile()
	if err == nil {
		// `version` was successfully read, so we can compare
		// with deployment and boot IDs extracted from backup's name

		// deployment ID is not enough on its own in scenario:
		// - first boot is okay, admin reboots the machine
		// - second boot is unhealthy, admin reboots the machine
		// - third boot should restore backup of data from #1 boot,
		//   but it would not, because the deployment ID didn't change

		deploy, boot := func() (string, string) {
			spl := strings.Split(string(backup), "_")
			return spl[0], spl[1]
		}()

		if versionFile.DeploymentID == deploy && versionFile.BootID == boot {
			klog.InfoS("Restore skipped - data directory already matches backup to restore",
				"backup-name", backup,
				"version", versionFile)
			return nil
		}
	}

	if err := dm.dataManager.Restore(backup); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	klog.InfoS("Restored backup", "name", backup)
	return nil
}
