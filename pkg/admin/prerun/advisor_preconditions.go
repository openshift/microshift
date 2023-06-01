package prerun

import (
	"github.com/openshift/microshift/pkg/admin/history"
)

// PreconditionsAdvisor helps answering questions related to deciding "should pre run procedure be performed?"
type PreconditionsAdvisor interface {
	IsOSTree() bool
	PreRunAlreadyRanCurrentBoot() bool
	PreRunWasSuccessful() bool
	PreRunStatus() history.PreRunStatus
}

var _ PreconditionsAdvisor = (*preconditionsAdvisor)(nil)

type preconditionsAdvisor struct {
	data decisionData
}

func NewPreconditionsAdvisor(decisionData decisionData) *preconditionsAdvisor {
	return &preconditionsAdvisor{data: decisionData}
}

func (pa *preconditionsAdvisor) IsOSTree() bool {
	return pa.data.IsOSTree
}

func (pa *preconditionsAdvisor) PreRunAlreadyRanCurrentBoot() bool {
	if pa.data.CurrentBootInfo == nil {
		return false
	}

	if pa.data.CurrentBootInfo.PreRun == "" {
		return false
	}

	return true
}

func (pa *preconditionsAdvisor) PreRunWasSuccessful() bool {
	if !pa.PreRunAlreadyRanCurrentBoot() {
		panic("PreRunStatus() should not be called if PreRunAlreadyRanCurrentBoot() returned false")
	}
	return pa.data.CurrentBootInfo.PreRun == history.PreRunSuccess
}

func (pa *preconditionsAdvisor) PreRunStatus() history.PreRunStatus {
	if !pa.PreRunAlreadyRanCurrentBoot() {
		panic("PreRunStatus() should not be called if PreRunAlreadyRanCurrentBoot() returned false")
	}
	return pa.data.CurrentBootInfo.PreRun
}
