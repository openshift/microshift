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

	if err := dm.perform(); err != nil {
		klog.ErrorS(err, "FAIL pre-run data management")
		return err
	}

	klog.InfoS("END pre-run data management")
	return nil
}

type dataManagement struct {
	dataManager datadir.Manager
}

func (dm *dataManagement) perform() error {
	if isOstree, err := util.PathExists("/run/ostree-booted"); err != nil {
		return fmt.Errorf("failed to check if system is ostree: %w", err)
	} else if !isOstree {
		klog.InfoS("System is not OSTree-based - skipping data management")
		return nil
	}

	klog.Info("START creating backup")
	if err := dm.backup(); err != nil {
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

func (dm *dataManagement) backup() error {
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

	currentBootID, err := getCurrentBootID()
	if err != nil {
		return err
	}
	if currentBootID == versionFile.BootID {
		// We don't want to create a backup if versionFile has current boot ID:
		// backup would be created for current boot - in the middle of it.
		// Because backups are not updated/overwritten, existence of such backup would prevent
		// creation of proper backup, i.e. backup for previous boot after reboot.
		klog.InfoS("Current boot ID and one stored in version file are the same - skipping backup",
			"current", currentBootID)
		return nil
	}

	existingBackups, err := getBackups(dm.dataManager)
	if err != nil {
		return err
	}

	newBackupName := versionFile.BackupName()
	if existingBackups.has(newBackupName) {
		klog.InfoS("Backup already exists", "name", newBackupName)
		return nil
	}

	if _, err := dm.dataManager.Backup(newBackupName); err != nil {
		return fmt.Errorf("failed to create backup %q: %w", newBackupName, err)
	}

	// after making a new backup, remove all old backups for the deployment
	existingBackups.getForDeployment(versionFile.DeploymentID).removeAll(dm.dataManager)

	// prune old backups
	if err := dm.removeBackupsWithoutExistingDeployments(existingBackups); err != nil {
		klog.ErrorS(err, "Failed to remove backups belonging to no longer existing deployments - ignoring")
	}

	return nil
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

	currentDeploymentID, err := GetCurrentDeploymentID()
	if err != nil {
		return err
	}

	existingBackups, err := getBackups(dm.dataManager)
	if err != nil {
		return err
	}

	backup := existingBackups.getForDeployment(currentDeploymentID).getOneOrNone()

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
		if err := dm.restore(backup, nil); err != nil {
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

	if err := dm.restore(backup, &versionFile); err != nil {
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

	if _, err := dm.dataManager.Backup(backupName); err != nil {
		return fmt.Errorf("failed to create new 4.13 backup: %w", err)
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

func (dm *dataManagement) restore(backup datadir.BackupName, vf *versionFile) error {
	if vf != nil {
		// `version` was successfully read, so we can compare
		// with deployment and boot IDs extracted from backup's name

		deploy, boot := func() (string, string) {
			spl := strings.Split(string(backup), "_")
			return spl[0], spl[1]
		}()

		if vf.DeploymentID == deploy && vf.BootID == boot {
			klog.InfoS("Restore skipped - data directory already matches backup to restore",
				"backup-name", backup,
				"version", *vf)
			return nil
		}
	}

	if err := dm.dataManager.Restore(backup); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}
	return nil
}
