package data

import "fmt"

type BackupConfig struct {
	// BackupsStorage is the base directory storing all backups
	BackupsStorage string

	// Name is backup's directory name
	Name string
}

type DataManager interface {
	Backup(BackupConfig) error
	Restore(BackupConfig) error
}

func NewDataManager() DataManager {
	return &dataManager{}
}

type dataManager struct{}

func (dm *dataManager) Backup(cfg BackupConfig) error {
	return makeBackup(cfg)
}

func (dm *dataManager) Restore(cfg BackupConfig) error {
	return fmt.Errorf("not implemented")
}
