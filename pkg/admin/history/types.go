package history

import (
	"errors"
	"sort"

	"github.com/openshift/microshift/pkg/admin/system"
	"k8s.io/klog/v2"
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

func (h *History) AddOrUpdate(boot system.Boot, info BootInfo) {
	if h.Boots == nil {
		h.Boots = make([]Boot, 0, 1)
	}

	for i, b := range h.Boots {
		if b.ID == boot.ID {
			oldInfo := h.Boots[i].BootInfo
			h.Boots[i].BootInfo = oldInfo.Update(info)

			klog.InfoS("Updated boot info", "old", oldInfo, "new", h.Boots[i])
			return
		}
	}

	b := Boot{
		Boot:     boot,
		BootInfo: info,
	}
	h.Boots = append([]Boot{b}, h.Boots...)
	klog.InfoS("Added boot info", "boot", b)

	sort.SliceStable(h.Boots, func(i, j int) bool {
		return h.Boots[i].BootTime.After(h.Boots[j].BootTime)
	})
}

func (h *History) RemoveOldEntries(keep int) {
	if len(h.Boots) > keep {
		removed := h.Boots[keep:]
		h.Boots = h.Boots[:keep]
		klog.InfoS("History to long - removed old entries",
			"max", keep,
			"removed", removed,
			"current", h.Boots)
	}
}
