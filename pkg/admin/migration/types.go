package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MigratonStatus string

const (
	MigrationSuccess MigratonStatus = "success"
	MigrationFailure MigratonStatus = "failure"
	MigrationRunning MigratonStatus = "running"
)

// Container for individual migration attempts
type MigrationResult struct {
	Error     error `json:"Error,omitempty"`
	Timestamp time.Time
	Status    MigratonStatus
	schema.GroupVersionResource
}

type MigrationResultList struct {
	Status MigratonStatus
	Items  []MigrationResult
}

func (m *MigrationResultList) WriteStatusFile(filePath string) error {
	data := fmt.Sprintf(`{"Status": "%s"}`, m.Status)
	return os.WriteFile(filePath, []byte(data), 0644)
}

func (m *MigrationResultList) WriteDataFile(filePath string) error {
	fileData, err := json.Marshal(m.Items)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, fileData, 0644)
}

type ErrRetriable struct {
	error
}

func (ErrRetriable) Temporary() bool { return true }

type ErrNotRetriable struct {
	error
}

func (ErrNotRetriable) Temporary() bool { return false }

// TemporaryError is a wrapper interface that is used to determine if an error can be retried.
type TemporaryError interface {
	error
	// Temporary should return true if this is a temporary error
	Temporary() bool
}
