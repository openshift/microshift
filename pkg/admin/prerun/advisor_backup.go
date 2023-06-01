package prerun

import (
	"github.com/openshift/microshift/pkg/admin/history"
)

// BackupAdvisor helps answering questions related to deciding "should data be backed up?"
type BackupAdvisor interface {
	DeviceBootedForTheFirstTime() bool
	BootHistoryExists() bool
	BootHistoryContainsPreviousBoot() bool
	PreviousBootWasHealthy() bool
}

var _ BackupAdvisor = (*backupAdvisor)(nil)

type backupAdvisor struct {
	data decisionData
}

func NewBackupAdvisor(decisionData decisionData) *backupAdvisor {
	return &backupAdvisor{data: decisionData}
}

func (a *backupAdvisor) DeviceBootedForTheFirstTime() bool {
	return a.data.PreviousBoot == nil
}

func (a *backupAdvisor) BootHistoryExists() bool {
	return a.data.BootHistory != nil
}

func (a *backupAdvisor) BootHistoryContainsPreviousBoot() bool {
	return a.data.PreviousBootInfo != nil
}

func (a *backupAdvisor) PreviousBootWasHealthy() bool {
	if !a.BootHistoryContainsPreviousBoot() {
		panic("PreviousBootWasHealthy() should not be called if BootHistoryContainsPreviousBoot() returned false")
	}
	return a.data.PreviousBootInfo.Health == history.Healthy
}
