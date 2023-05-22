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

func NewDataManager() *IDataManager {
	return &IDataManager{}
}

type IDataManager struct{}

func (dm *IDataManager) Backup(cfg BackupConfig) error {
	return makeBackup(cfg)
}

func (dm *IDataManager) Restore(cfg BackupConfig) error {
	return fmt.Errorf("not implemented")
}
