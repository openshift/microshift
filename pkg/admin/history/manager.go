package history

import (
	"errors"
	"fmt"
	"sort"

	"k8s.io/klog/v2"
)

const (
	maxBootHistory = 10
)

type HistoryManager interface {
	Get() (*History, error)
	Update(DeploymentBoot, BootInfo) error
}

func NewHistoryManager(historyStorage HistoryStorage) *historyManager {
	return &historyManager{
		storage: historyStorage,
	}
}

var _ HistoryManager = (*historyManager)(nil)

type historyManager struct {
	storage HistoryStorage
}

func (dhm *historyManager) Get() (*History, error) {
	history, err := dhm.storage.Load()
	if err != nil {
		return nil, fmt.Errorf("loading history failed: %w", err)
	}

	if history.Boots == nil || len(history.Boots) == 0 {
		// no health history yet
		return nil, ErrNoHistory
	}

	sort.Slice(history.Boots, func(i, j int) bool {
		return history.Boots[i].BootTime.After(history.Boots[j].BootTime)
	})

	return history, nil
}

func (dhm *historyManager) Update(dp DeploymentBoot, info BootInfo) error {
	if dp.ID == "" {
		return fmt.Errorf("missing id")
	}
	if dp.DeploymentID == "" {
		return fmt.Errorf("missing deployment id")
	}

	history, err := dhm.Get()
	if err != nil {
		if !errors.Is(err, ErrNoHistory) {
			return fmt.Errorf("loading history failed: %w", err)
		}
		klog.InfoS("Boot history does not exist (yet)")
		history = &History{}
	}

	klog.InfoS("Current boot history", "history", history)
	history.AddOrUpdate(dp, info)
	history.RemoveOldEntries(maxBootHistory)
	klog.InfoS("Updated boot history", "history", history)

	return dhm.storage.Save(history)
}
