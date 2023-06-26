package migration

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	apitypes "k8s.io/apimachinery/pkg/types"
)

type MigratonStatus string

const (
	MigrationSuccess MigratonStatus = "success"
	MigrationFailure MigratonStatus = "failure"
)

// Container for individual migration attempts
type MigrationResult struct {
	Error           error                       `json:"error,omitempty"`
	ResourceVersion schema.GroupVersionResource `json:"resourceVersion"`
	Timestamp       time.Time                   `json:"timestamp"`
	NamespacedName  apitypes.NamespacedName     `json:"namespacedName,omitempty"`
}

type MigrationResultList struct {
	Status MigratonStatus
	Items  []MigrationResult
}

func (m MigrationResultList) String() string {
	buffer := strings.Builder{}
	for _, result := range m.Items {
		objectInfo := result.ResourceVersion.String()
		objectInfo = fmt.Sprintf("%s Namespace=%s Name=%s", objectInfo, result.NamespacedName.Namespace, result.NamespacedName.Name)

		info := fmt.Sprintf("%s MigrationStatus=%s %s\n", result.Timestamp.String(), MigrationSuccess, objectInfo)
		if result.Error != nil {
			info = fmt.Sprintf("%s MigrationStatus=%s %s : %v\n", result.Timestamp.String(), MigrationFailure, objectInfo, result.Error)
		}
		buffer.WriteString(info)
	}
	return buffer.String()
}

func (m MigrationResultList) Bytes() []byte {
	return []byte(m.String())
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
