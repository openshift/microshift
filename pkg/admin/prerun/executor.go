package prerun

import (
	"fmt"
	"strings"
	"time"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/admin/history"
	"github.com/openshift/microshift/pkg/admin/system"
	"k8s.io/klog/v2"
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
	deployID := e.decisionData.PreviousBootInfo.DeploymentID

	backups, err := e.dataManager.GetBackupList()
	if err != nil {
		return err
	}

	existingDeploymentBackups := make([]data.BackupName, 0)
	for _, b := range backups {
		if strings.Contains(string(b), string(deployID)) {
			existingDeploymentBackups = append(existingDeploymentBackups, b)
		}
	}

	name := fmt.Sprintf("%s_%s", deployID, time.Now().UTC().Format("20060102_150405"))
	if err := e.dataManager.Backup(data.BackupName(name)); err != nil {
		return err
	}

	if len(existingDeploymentBackups) > 0 {
		klog.InfoS("Removing old deployment backups",
			"deployment", deployID,
			"backups", existingDeploymentBackups)
		for _, b := range existingDeploymentBackups {
			if err := e.dataManager.RemoveBackup(b); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *executor) UpdatePreRunStatus(status history.PreRunStatus) error {
	return e.historyManager.Update(
		history.NewDeploymentBoot(*e.decisionData.CurrentBoot, e.decisionData.CurrentDeploymentID),
		history.BootInfo{PreRun: status},
	)
}
