package prerun

import (
	"errors"
	"fmt"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/admin/history"
	"github.com/openshift/microshift/pkg/admin/system"

	"k8s.io/klog/v2"
)

type decisionData struct {
	IsOSTree bool

	BootHistory *history.History

	CurrentBoot  *system.Boot
	PreviousBoot *system.Boot

	CurrentBootInfo  *history.Boot
	PreviousBootInfo *history.Boot
}

func NewDecisionData(
	dataManager data.Manager,
	systemInfo system.SystemInfo,
	historyManager history.HistoryManager,
) (*decisionData, error) {
	d := &decisionData{}
	var err error

	d.IsOSTree, err = systemInfo.IsOSTree()
	if err != nil {
		return nil, fmt.Errorf("getting ostree information failed: %w", err)
	}

	d.BootHistory, err = historyManager.Get()
	if err != nil {
		if errors.Is(err, history.ErrNoHistory) {
			klog.InfoS("History does not exist")
			d.BootHistory = nil
		} else {
			return nil, fmt.Errorf("getting boot history failed: %w", err)
		}
	}

	d.CurrentBoot, err = systemInfo.GetCurrentBoot()
	if err != nil {
		return nil, fmt.Errorf("getting current boot info failed: %w", err)
	} else {
		currBootInfo, found := d.BootHistory.GetBootByID(d.CurrentBoot.ID)
		if found {
			d.CurrentBootInfo = &currBootInfo
		}
	}

	d.PreviousBoot, err = systemInfo.GetPreviousBoot()
	if err != nil {
		return nil, fmt.Errorf("getting previous boot info failed: %w", err)
	} else {
		prevBootInfo, found := d.BootHistory.GetBootByID(d.PreviousBoot.ID)
		if found {
			d.PreviousBootInfo = &prevBootInfo
		}
	}
	klog.InfoS("Pre run decision data", "data", *d)

	return d, nil
}
