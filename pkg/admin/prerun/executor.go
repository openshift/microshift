package prerun

import (
	"fmt"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/admin/history"
	"github.com/openshift/microshift/pkg/admin/system"
)

type Executor interface {
	BackupPreviousBoot() error
	UpdatePreRunStatus(history.PreRunStatus) error
}

var _ Executor = (*executor)(nil)

type executor struct {
	dataManager    data.Manager
	systemInfo     system.SystemInfo
	historyManager history.HistoryManager
	decisionData   decisionData
}

func NewExecutor(
	dataManager data.Manager,
	systemInfo system.SystemInfo,
	historyManager history.HistoryManager,
	decisionData decisionData,
) *executor {
	return &executor{
		dataManager:    dataManager,
		systemInfo:     systemInfo,
		historyManager: historyManager,
		decisionData:   decisionData,
	}
}

func (e *executor) BackupPreviousBoot() error {
	if e.decisionData.PreviousBootInfo == nil {
		return fmt.Errorf("unexpected request to backup the data - previous boot info is missing")
	}
	name := e.decisionData.PreviousBootInfo.DeploymentID
	return e.dataManager.Backup(data.BackupName(name))
}

func (e *executor) UpdatePreRunStatus(status history.PreRunStatus) error {
	return e.historyManager.Update(
		*e.decisionData.CurrentBoot,
		history.BootInfo{PreRun: status},
	)
}
