package data

import (
	"fmt"
	"path/filepath"
)

type EmptyArgErr struct {
	argName string
}

func (e *EmptyArgErr) Error() string {
	return fmt.Sprintf("empty argument: %s", e.argName)
}

type BackupName string
type StoragePath string

func (sp StoragePath) GetBackupPath(backupName BackupName) string {
	return filepath.Join(string(sp), string(backupName))
}

func (sp StoragePath) SubStorage(subdir string) StoragePath {
	return StoragePath(filepath.Join(string(sp), subdir))
}

type Manager interface {
	Backup(BackupName) (string, error)
	Restore(BackupName) error

	BackupExists(BackupName) (bool, error)
	GetBackupPath(BackupName) string
	GetBackupList() ([]BackupName, error)
	RemoveBackup(BackupName) error

	RemoveData() error
}
