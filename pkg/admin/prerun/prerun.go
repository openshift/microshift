package prerun

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	datadir "github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

var (
	restoreFilepath = filepath.Join(config.BackupsDir, "restore")
)

func DataManagement(dataManager datadir.Manager) error {
	klog.InfoS("START pre-run data management")

	dm := dataManagement{
		dataManager: dataManager,
	}

	if err := dm.newPerform(); err != nil {
		klog.ErrorS(err, "FAIL pre-run data management")
		return err
	}

	klog.InfoS("END pre-run data management")
	return nil
}

type importantPathsExistence struct {
	data, version, health bool
}

type dataManagement struct {
	dataManager datadir.Manager
}

func (dm *dataManagement) newPerform() error {
	if isOstree, err := util.PathExists("/run/ostree-booted"); err != nil {
		return fmt.Errorf("failed to check if system is ostree: %w", err)
	} else if !isOstree {
		klog.InfoS("System is not OSTree-based - skipping data management")
		return nil
	}

	klog.Info("START creating backup")
	if err := dm.alwaysBackup(); err != nil {
		klog.ErrorS(err, "FAIL creating backup")
		return err
	}
	klog.Info("END creating backup")

	klog.InfoS("START optional restore")
	if err := dm.optionalRestore(); err != nil {
		klog.ErrorS(err, "FAIL optional restore")
		return err
	}
	klog.InfoS("END optional restore")

	return nil
}

func (dm *dataManagement) alwaysBackup() error {
	dataExists, err := util.PathExistsAndIsNotEmpty(config.DataDir, ".nodename")
	if err != nil {
		return fmt.Errorf("failed to check if data directory exists: %w", err)
	}
	if !dataExists {
		klog.InfoS("MicroShift data does not exist - skipping backup, continuing startup")
		return nil
	}

	versionFileExists, err := util.PathExistsAndIsNotEmpty(versionFilePath)
	if err != nil {
		return fmt.Errorf("checking if version metadata exists failed: %w", err)
	}
	if !versionFileExists {
		klog.InfoS("Data exists, but version file is missing - assuming pre-4.14 data")
		return dm.backup413()
	}

	versionFile, err := getVersionFile()
	if err != nil {
		return fmt.Errorf("loading version metadata failed: %w", err)
	}
	klog.InfoS("Contents of version file", "contents", versionFile)
	return dm.newBackup(versionFile)
}

func (dm *dataManagement) optionalRestore() error {
	restoreFileExists, err := util.PathExists(restoreFilepath)
	if err != nil {
		return err
	}

	if !restoreFileExists {
		klog.InfoS("Restore marker file does not exist - skipping restore, "+
			"continuing startup with current data", "path", restoreFilepath)
		return nil
	}
	klog.InfoS("Restore marker file exists - attempting to restore",
		"path", restoreFilepath)

	currentDeploymentID, err := getCurrentDeploymentID()
	if err != nil {
		return err
	}

	existingBackups, err := getBackups(dm.dataManager)
	if err != nil {
		return err
	}

	backup := existingBackups.getForDeployment(currentDeploymentID).getOnlyHealthyBackups().getOneOrNone() // TODO: remove getOnlyHealthyBackups

	if backup == "" {
		klog.InfoS("WARNING: MicroShift was instructed to restore a backup, "+
			"but there is no backup for the deployment - continuing start up with current data",
			"deploymentID", currentDeploymentID)
		return nil
	}

	dataExists, err := util.PathExistsAndIsNotEmpty(config.DataDir, ".nodename")
	if err != nil {
		return fmt.Errorf("failed to check if data directory exists: %w", err)
	}
	if !dataExists {
		// Data doesn't exist, we have suitable backup, let's restore
		if err := dm.restore(backup); err != nil {
			klog.ErrorS(err, "Failed to restore")
			return err
		}

		if err := dm.removeRestoreFile(); err != nil {
			return err
		}

		return nil
	}

	versionFileExists, err := util.PathExistsAndIsNotEmpty(versionFilePath)
	if err != nil {
		return fmt.Errorf("checking if version metadata exists failed: %w", err)
	}
	if !versionFileExists {
		klog.InfoS("Backup found for active deployment, MicroShift data exists, but version file does not - " +
			"assuming data is pre-4.14 and this new attempt to upgrade - not restoring, continuing start up with current data. " +
			" If restore is expected perform manual restore.")
		// Data exists, but version file does not - this suggests previous boot was running pre-4.14 MicroShift,
		// however backup for this deployment exists.
		// It could be a scenario:
		// - pre-4.14 deployment
		// - post-4.14 deployment that failed and rolled back
		//   (though it rebooted several times resulting in backups)
		// - pre-4.14 backup was restored manually, but post-4.14 backup was not removed
		// - the same post-4.14 commit was deployed.
		// Looks more like an upgrade, rather than rollback, so we should not restore and just let MicroShift run.
		if err := dm.removeRestoreFile(); err != nil {
			return err
		}

		return nil
	}

	versionFile, err := getVersionFile()
	if err != nil {
		return fmt.Errorf("loading version metadata failed: %w", err)
	}

	klog.InfoS("Contents of version file", "contents", versionFile)

	if currentDeploymentID == versionFile.DeploymentID {
		klog.InfoS("Current active deployment ID and deployment ID in version file are the same - " +
			"not restoring, continuing startup with current data.")
		// MicroShift just created a backup according to information in versionFile,
		// so if current active deployment is the same as previous boot's deployment,
		// it would restore the very same data it backed up, so let's just skip restore.

		// This could also happen if version checks blocked the upgrade (e.g. upgrading to Y+2),
		// would refuse to run and not update version file. After rolling back, the data would be unchanged.
		if err := dm.removeRestoreFile(); err != nil {
			return err
		}

		return nil
	}

	// At this point:
	// - MicroShift was instructed to restore by existence of the restoreFilepath
	// - Backup for current deployment exists
	// - versionFile.DeploymentID is different from current deployment according to ostree
	//	 (meaning last time MicroShift ran was on different deployment).
	//
	// Let's restore and remove the `restore` file

	if err := dm.restore(backup); err != nil {
		klog.ErrorS(err, "Failed to restore")
		return err
	}

	if err := dm.removeRestoreFile(); err != nil {
		return err
	}

	return nil
}

func (dm *dataManagement) removeRestoreFile() error {
	klog.InfoS("Removing restore marker filepath", "path", restoreFilepath)
	if err := os.Remove(restoreFilepath); err != nil {
		klog.ErrorS(err, "FATAL ERROR: Failed to remove file - existence of the file will result in unexpected data restores: "+
			"remove the file manually and make sure microshift.service can manipulate it",
			"file", restoreFilepath)
		return err
	}
	klog.InfoS("Removed restore marker filepath", "path", restoreFilepath)
	return nil
}

func (dm *dataManagement) Perform() error {
	if isOstree, err := util.PathExists("/run/ostree-booted"); err != nil {
		return fmt.Errorf("failed to check if system is ostree: %w", err)
	} else if !isOstree {
		klog.InfoS("System is not OSTree-based - skipping data management")
		return nil
	}

	existence, err := dm.getPathsExistence()
	if err != nil {
		return fmt.Errorf("failed to get existence of important paths: %w", err)
	}

	if existence.data && existence.version && existence.health {
		// 8 - regular flow
		klog.InfoS("START regular data management process")

		if err := dm.regularPrerun(); err != nil {
			klog.ErrorS(err, "FAIL regular data management process")
			return err
		}

		klog.InfoS("END regular data management")
		return nil
	}

	klog.InfoS("START handling special case of data management")
	if err := dm.handleSpecialCases(existence); err != nil {
		klog.ErrorS(err, "FAIL handling special case of data management")
		return err
	}
	klog.InfoS("END handling special case of data management")
	return nil
}

func (dm *dataManagement) getPathsExistence() (importantPathsExistence, error) {
	var err error
	pathsExistence := importantPathsExistence{}

	pathsExistence.data, err = util.PathExistsAndIsNotEmpty(config.DataDir, ".nodename")
	if err != nil {
		return pathsExistence, fmt.Errorf("failed to check if data directory exists: %w", err)
	}

	pathsExistence.version, err = util.PathExistsAndIsNotEmpty(versionFilePath)
	if err != nil {
		return pathsExistence, fmt.Errorf("checking if version metadata exists failed: %w", err)
	}

	pathsExistence.health, err = util.PathExistsAndIsNotEmpty(healthFilepath)
	if err != nil {
		return pathsExistence, fmt.Errorf("failed to check if health file exists: %w", err)
	}

	klog.InfoS("Existence of important paths",
		"data-exists?", pathsExistence.data,
		"version-exists?", pathsExistence.version,
		"health-exists?", pathsExistence.health,
	)

	return pathsExistence, nil
}

func (dm *dataManagement) handleSpecialCases(existence importantPathsExistence) error {
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

	if !existence.data {
		// Implies !existence.version

		// 1
		if !existence.health {
			klog.InfoS("Neither data dir nor health file exist (assuming first start of MicroShift)" +
				" - skipping data management")
			return nil
		}

		// 2
		klog.InfoS("START handling missing data but existing health file")
		if err := dm.missingDataExistingHealth(); err != nil {
			klog.ErrorS(err, "FAIL handling missing data but existing health file")
			return err
		}
		klog.InfoS("END handling missing data but existing health file")
	}

	if !existence.version {
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
		klog.InfoS("START creating 4.13 backup")
		if err := dm.backup413(); err != nil {
			klog.ErrorS(err, "FAIL creating 4.13 backup")
			return err
		}
		klog.InfoS("END creating 4.13 backup")
		return nil
	}

	// 7 - !existence.health
	//
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
	return nil
}

// regularPrerun performs actions in prerun flow that is most expected in day to day usage
// (i.e. data, version metadata, and health information exist)
func (dm *dataManagement) regularPrerun() error {
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
		klog.Info("START handling healthy system")
		if err = dm.handleHealthy(health, currentDeploymentID); err != nil {
			klog.ErrorS(err, "FAIL handling healthy system")
			return err
		}

		klog.Info("END handling healthy system")
		return nil
	}

	klog.Info("START handling unhealthy system")
	if err = dm.handleUnhealthy(health); err != nil {
		klog.ErrorS(err, "FAIL handling unhealthy system")
		return err
	}
	klog.InfoS("END handling unhealthy system")

	return nil
}

func (dm *dataManagement) backup413() error {
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
	return nil
}

// missingDataExistingHealth handles situation where MicroShift's data doesn't exist
// but health file is present.
//
// Depending on health persisted in health file it might:
//   - try to restore a backup for current deployment (if exists), or
//   - proceed with fresh start if "healthy" was persisted (nothing to back up)
//     or backup does not exists (nothing to restore)
func (dm *dataManagement) missingDataExistingHealth() error {
	health, err := getHealthInfo()
	if err != nil {
		return fmt.Errorf("failed to determine the current system health: %w", err)
	}

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

	klog.InfoS("START restoring backup", "name", backup)
	if err := dm.restore(backup); err != nil {
		klog.ErrorS(err, "FAIL restoring backup", "name", backup)
		return err
	}
	klog.InfoS("END restoring backup", "name", backup)
	return nil
}

func (dm *dataManagement) newBackup(vf versionFile) error {
	existingBackups, err := getBackups(dm.dataManager)
	if err != nil {
		return err
	}

	newBackupName := vf.BackupName()
	if existingBackups.has(newBackupName) {
		klog.InfoS("Backup already exists", "name", newBackupName)
		return nil
	}

	if err := dm.dataManager.Backup(newBackupName); err != nil {
		return fmt.Errorf("failed to create backup %q: %w", newBackupName, err)
	}

	// after making a new backup, remove all old backups for the deployment
	existingBackups.getForDeployment(vf.DeploymentID).removeAll(dm.dataManager)

	// prune old backups
	if err := dm.removeBackupsWithoutExistingDeployments(existingBackups); err != nil {
		klog.ErrorS(err, "Failed to remove backups belonging to no longer existing deployments - ignoring")
	}

	return nil
}

func (dm *dataManagement) backup(health *HealthInfo) error {
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

	return nil
}

func (dm *dataManagement) handleHealthy(health *HealthInfo, currentDeploymentID string) error {
	klog.Info("START creating backup")
	if err := dm.backup(health); err != nil {
		klog.ErrorS(err, "FAIL creating backup")
		return err
	}
	klog.Info("END creating backup")

	if health.DeploymentID != currentDeploymentID {
		klog.InfoS("START handling deployment switch - current deployment and health.deploymentID are different")
		if err := dm.handleDeploymentSwitch(currentDeploymentID); err != nil {
			klog.ErrorS(err, "FAIL handling deployment switch")
			return err
		}
		klog.Info("END handling deployment switch")
	}

	return nil
}

func (dm *dataManagement) handleUnhealthy(health *HealthInfo) error {
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
		klog.InfoS("START restoring backup", "name", backup)
		err = dm.restore(backup)
		if err != nil {
			klog.ErrorS(err, "FAIL restoring backup", "name", backup)
			return err
		}
		klog.InfoS("END restoring backup", "name", backup)
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
		klog.InfoS("START restoring rollback deployment's backup to try starting from healthy data", "name", rollbackBackup)
		if err := dm.restore(rollbackBackup); err != nil {
			klog.ErrorS(err, "FAIL restoring rollback deployment's backup", "name", rollbackBackup)
			return err
		}
		klog.InfoS("END restoring rollback deployment's backup", "name", rollbackBackup)
		return nil
	}

	// DeployID in health.json is neither booted nor rollback deployment,
	// so current deployment was staged over deployment without MicroShift
	// but MicroShift data exists (created by another deployment that rolled back).
	klog.InfoS("Deployment in health metadata is neither currently booted nor rollback deployment - backing up, then removing existing data for clean start")
	if err := dm.backup(health); err != nil {
		klog.ErrorS(err, "Failed to backup data of unhealthy system - ignoring")
	}

	klog.InfoS("START removing MicroShift data")
	if err := dm.dataManager.RemoveData(); err != nil {
		klog.ErrorS(err, "FAIL removing MicroShift data")
		return err
	}
	klog.InfoS("END removing MicroShift data")

	return nil
}

func (dm *dataManagement) handleDeploymentSwitch(currentDeploymentID string) error {
	// System booted into different deployment
	// It might be a rollback (restore existing backup), or
	// it might be a newly staged one (continue start up)

	existingBackups, err := getBackups(dm.dataManager)
	if err != nil {
		return err
	}
	backup := existingBackups.getForDeployment(currentDeploymentID).getOnlyHealthyBackups().getOneOrNone()

	if backup != "" {
		klog.Info("START restoring existing backup for current deployment", "name", backup)
		if err := dm.restore(backup); err != nil {
			klog.ErrorS(err, "FAIL restoring existing backup for current deployment", "name", backup)
			return err
		}
		klog.Info("END restoring existing backup for current deployment", "name", backup)
	} else {
		klog.Info("No backup for current deployment to restore - continuing start up")
	}

	return nil
}

func (dm *dataManagement) removeBackupsWithoutExistingDeployments(backups Backups) error {
	klog.InfoS("Attempting to remove backups for no longer existing deployments")
	deployments, err := getAllDeploymentIDs()
	if err != nil {
		return err
	}

	toRemove := backups.getDangling(deployments)
	if len(toRemove) == 0 {
		klog.InfoS("Found no backups for no longer existing deployments to remove")
		return nil
	}

	klog.InfoS("Removing backups for no longer existing deployments",
		"backups-to-remove", toRemove)
	toRemove.removeAll(dm.dataManager)

	return nil
}

func (dm *dataManagement) restore(backup datadir.BackupName) error {
	versionFile, err := getVersionFile()
	if err == nil {
		// `version` was successfully read, so we can compare
		// with deployment and boot IDs extracted from backup's name

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
	return nil
}
