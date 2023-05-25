package data

import "fmt"

type BackupConfig struct {
	// Storage is the base directory storing all backups
	Storage string

	// Name is backup's directory name
	Name string
}

type Manager interface {
	Backup(BackupConfig) error
	Restore(BackupConfig) error
}

func NewManager() *manager {
	return &manager{}
}

type manager struct{}

func (dm *manager) Backup(cfg BackupConfig) error {
	return makeBackup(cfg)
}

func (dm *manager) Restore(cfg BackupConfig) error {
	return fmt.Errorf("not implemented")
}
