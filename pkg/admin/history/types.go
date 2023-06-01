package history

import (
	"errors"

	"github.com/openshift/microshift/pkg/admin/system"
)

var (
	ErrNoHistory = errors.New("no history")
)

type Health string

const (
	Unknown   Health = "unknown"
	Healthy   Health = "healthy"
	Unhealthy Health = "unhealthy"
)

type PreRunStatus string

const (
	PreRunUnknown              PreRunStatus = "unknown"
	PreRunSuccess              PreRunStatus = "success"
	PreRunBackupFailed         PreRunStatus = "backup-failed"
	PreRunRestoreFailed        PreRunStatus = "restore-failed"
	PreRunMigrationFailed      PreRunStatus = "migration-failed"
	PreRunMetadataUpdateFailed PreRunStatus = "metadata-update-failed"
)

type BootInfo struct {
	Health Health       `json:"health"`
	PreRun PreRunStatus `json:"pre_run"`
}

func (bi BootInfo) Update(new BootInfo) BootInfo {
	if new.Health != "" {
		bi.Health = new.Health
	}
	if new.PreRun != "" {
		bi.PreRun = new.PreRun
	}
	return bi
}

type Boot struct {
	system.Boot
	BootInfo
}

type History struct {
	Boots []Boot `json:"boots"`
}

func (h *History) GetBootByID(id system.BootID) (Boot, bool) {
	if h == nil {
		return Boot{}, false
	}
	if h.Boots == nil {
		return Boot{}, false
	}

	for _, b := range h.Boots {
		if id == b.ID {
			return b, true
		}
	}
	return Boot{}, false
}
