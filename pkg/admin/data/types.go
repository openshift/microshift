package data

import (
	"fmt"
)

type EmptyArgErr struct {
	argName string
}

func (e *EmptyArgErr) Error() string {
	return fmt.Sprintf("empty argument: %s", e.argName)
}

type StoragePath string
type BackupName string

type Manager interface {
	Backup(BackupName) error
	Restore(BackupName) error

	BackupExists(BackupName) (bool, error)
	GetBackupPath(BackupName) string
}
