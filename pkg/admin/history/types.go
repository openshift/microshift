package history

import (
	"errors"
	"fmt"

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

func StringToHealth(s string) (Health, error) {
	h := Health(s)

	switch h {
	case Healthy, Unhealthy:
		return h, nil
	default:
		return Health(""),
			fmt.Errorf("invalid value: expected %s or %s", Healthy, Unhealthy)
	}
}

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
