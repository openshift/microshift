package data

import (
	"errors"
	"fmt"
)

type BackupConfig struct {
	// Storage is the base directory storing all backups
	Storage string

	// Name is backup's directory name
	Name string
}

func (bc BackupConfig) Validate() error {
	var err error
	if bc.Storage == "" {
		err = fmt.Errorf("backup storage must not be empty")
	}
	if bc.Name == "" {
		err = errors.Join(err, fmt.Errorf("backup name must not be empty"))
	}
	return err
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
