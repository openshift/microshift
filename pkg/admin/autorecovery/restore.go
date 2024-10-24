package autorecovery

import (
	"fmt"
	"os"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

const (
	failedSubstorageName   = "failed"
	restoredSubstorageName = "restored"
)

type Manager struct {
	storage    data.StoragePath
	saveFailed bool
}

func NewManager(storage data.StoragePath, saveFailed bool) (*Manager, error) {
	if storage == "" {
		return nil, fmt.Errorf("`storage` argument is empty")
	}

	return &Manager{storage: storage, saveFailed: saveFailed}, nil
}

func (m *Manager) PerformRestore() error {
	if err := storageShouldExist(m.storage); err != nil {
		return err
	}

	existingState, err := GetState(m.storage)
	if err != nil {
		return err
	}

	restoreCandidate, err := getCandidateForRestore(m.storage, existingState)
	if err != nil {
		return err
	}

	/*
		Preparations - initial file system operations:
		  COPY    $STORAGE/$CANDIDATE -> /var/lib/microshift.tmp
		  COPY    /var/lib/microshift -> $STORAGE/failed/$DATE_DEPLOY-ID.tmp
		  CREATE  $STORAGE/state.json.new
		  -       $STORAGE/$PREVIOUSLY_RESTORED									Should already exist, nothing to do.

		"Closing the transaction"
		  RENAME    /var/lib/microshift.tmp -> /var/lib/microshift
						This operation is first, because it's the most important one - whole point of the restore is to get to running state.
		  RENAME    $STORAGE/failed/DATE_DEPLOY-ID.tmp -> $STORAGE/failed/DATE_DEPLOY-ID
		  RENAME    $STORAGE/state.json.new -> $STORAGE/state.json
		  RENAME    $STORAGE/$PREVIOUSLY_RESTORED -> $STORAGE/restored/$PREVIOUSLY_RESTORED
	*/

	if err := data.CheckIfEnoughSpaceToRestore(m.storage.GetBackupPath(restoreCandidate.Name())); err != nil {
		return err
	}

	// Copies/creations into intermediate destinations
	var oldData *data.AtomicDirCopy
	if m.saveFailed {
		if err := data.CheckIfEnoughSpaceToBackUp(string(m.storage)); err != nil {
			return err
		}

		oldDataBackupName, err := GetBackupName()
		if err != nil {
			return err
		}
		failedStorage := m.storage.SubStorage(failedSubstorageName)
		if err := os.MkdirAll(string(failedStorage), 0600); err != nil {
			return fmt.Errorf("failed to create %q subdirectory: %w", failedSubstorageName, err)
		}
		oldData = &data.AtomicDirCopy{Source: config.DataDir, Destination: failedStorage.GetBackupPath(oldDataBackupName)}
		if err := oldData.CopyToIntermediate(); err != nil {
			return fmt.Errorf("old microshift data: %w", err)
		}
	}

	newData := data.AtomicDirCopy{Source: m.storage.GetBackupPath(restoreCandidate.Name()), Destination: config.DataDir}
	if err := newData.CopyToIntermediate(); err != nil {
		if rollbackErr := oldData.RollbackIntermediate(); rollbackErr != nil {
			klog.ErrorS(rollbackErr, "Failed to rollback intermediate state for old data")
		}
		return fmt.Errorf("new microshift data: %w", err)
	}

	newState := NewState(m.storage, restoreCandidate.Name())
	if err := newState.SaveToIntermediate(); err != nil {
		if rollbackErr := oldData.RollbackIntermediate(); rollbackErr != nil {
			klog.ErrorS(rollbackErr, "Failed to rollback intermediate state for old data")
		}
		if rollbackErr := newData.RollbackIntermediate(); rollbackErr != nil {
			klog.ErrorS(rollbackErr, "Failed to rollback intermediate state for new data")
		}
		return fmt.Errorf("new state file: %w", err)
	}

	// Renames into final destinations
	if err := newData.RenameToFinal(); err != nil {
		return fmt.Errorf("new microshift data: %w", err)
	}
	// Update the state file right after /var/lib/microshift is restored from a backup
	if err := newState.MoveToFinal(); err != nil {
		return fmt.Errorf("new state file: %w", err)
	}
	if err := oldData.RenameToFinal(); err != nil {
		return fmt.Errorf("old microshift data: %w", err)
	}

	//nolint:nestif
	if existingState != nil {
		previous := m.storage.GetBackupPath(existingState.LastBackup)
		exists, err := util.PathExists(previous)
		if err != nil {
			klog.ErrorS(err, "Failed to check if previously restored backup exists")
			return err
		}
		if exists {
			restoredStorage := m.storage.SubStorage(restoredSubstorageName)
			if err := os.MkdirAll(string(restoredStorage), 0600); err != nil {
				return fmt.Errorf("failed to create `restored` subdirectory: %w", err)
			}

			previouslyRestored := data.AtomicDirCopy{
				Source:      m.storage.GetBackupPath(existingState.LastBackup),
				Destination: restoredStorage.GetBackupPath(existingState.LastBackup),
			}
			if err := previouslyRestored.RenameToFinal(); err != nil {
				return fmt.Errorf("previously restored backup: %w", err)
			}
		}
	}

	klog.InfoS("Auto-recovery restore completed")
	return nil
}

func getCandidateForRestore(storagePath data.StoragePath, existingState *state) (Backup, error) {
	backups, err := GetBackups(storagePath)
	if err != nil {
		return Backup{}, err
	}

	if existingState != nil {
		backups = backups.RemoveBackup(existingState.LastBackup)
	}

	if len(backups) == 0 {
		klog.InfoS("There are no candidate backups for restoring!")
		return Backup{}, fmt.Errorf("no backups for restoring")
	}

	ver, err := getVersion()
	if err != nil {
		return Backup{}, err
	}

	backups = backups.FilterByVersion(ver)
	if len(backups) == 0 {
		klog.InfoS("There are no candidate backups for restoring after filtering by version!")
		return Backup{}, fmt.Errorf("no backups for restoring")
	}
	klog.InfoS("Potential backups", "bz", backups)

	restoreCandidate := backups.GetMostRecent()
	klog.InfoS("Candidate backup for restore", "b", restoreCandidate)

	return restoreCandidate, nil
}

func storageShouldExist(storage data.StoragePath) error {
	path := string(storage)
	exists, err := util.PathExists(path)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%q doesn't exist", path)
	}
	return nil
}
