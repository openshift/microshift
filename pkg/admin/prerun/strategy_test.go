package prerun

import (
	"errors"
	"testing"

	"github.com/openshift/microshift/pkg/admin/history"
	"github.com/openshift/microshift/pkg/admin/prerun/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_Strategy(t *testing.T) {
	testData := []struct {
		name        string
		setupMocks  func(pa *mocks.PreconditionsAdvisor, ba *mocks.BackupAdvisor, ex *mocks.Executor)
		shouldRun   bool
		errExpected bool
	}{
		{
			name: "Does not run on non-ostree systems",
			setupMocks: func(pa *mocks.PreconditionsAdvisor, ba *mocks.BackupAdvisor, ex *mocks.Executor) {
				pa.EXPECT().IsOSTree().Return(false)
			},
			errExpected: false,
		},
		{
			name: "Does not run if already successfully ran during current boot",
			setupMocks: func(pa *mocks.PreconditionsAdvisor, ba *mocks.BackupAdvisor, ex *mocks.Executor) {
				pa.EXPECT().IsOSTree().Return(true)
				pa.EXPECT().PreRunAlreadyRanCurrentBoot().Return(true)
				pa.EXPECT().PreRunWasSuccessful().Return(true)
			},
			errExpected: false,
		},
		{
			name: "Does not run if already ran and was unsuccessful",
			setupMocks: func(pa *mocks.PreconditionsAdvisor, ba *mocks.BackupAdvisor, ex *mocks.Executor) {
				pa.EXPECT().IsOSTree().Return(true)
				pa.EXPECT().PreRunAlreadyRanCurrentBoot().Return(true)
				pa.EXPECT().PreRunWasSuccessful().Return(false)
				pa.EXPECT().PreRunStatus().Return(history.PreRunBackupFailed)
			},
			errExpected: true,
		},
		{
			name: "Data is not backed up if this is first device boot",
			setupMocks: func(pa *mocks.PreconditionsAdvisor, ba *mocks.BackupAdvisor, ex *mocks.Executor) {
				pa.EXPECT().IsOSTree().Return(true)
				pa.EXPECT().PreRunAlreadyRanCurrentBoot().Return(false)

				ba.EXPECT().DeviceBootedForTheFirstTime().Return(true)

				ex.EXPECT().UpdatePreRunStatus(history.PreRunSuccess).Return(nil)
			},
			errExpected: false,
		},
		{
			name: "Data is not backed up if boot history doesn't exist",
			setupMocks: func(pa *mocks.PreconditionsAdvisor, ba *mocks.BackupAdvisor, ex *mocks.Executor) {
				pa.EXPECT().IsOSTree().Return(true)
				pa.EXPECT().PreRunAlreadyRanCurrentBoot().Return(false)

				ba.EXPECT().DeviceBootedForTheFirstTime().Return(false)
				ba.EXPECT().BootHistoryExists().Return(false)

				ex.EXPECT().UpdatePreRunStatus(history.PreRunSuccess).Return(nil)
			},
			errExpected: false,
		},
		{
			name: "Data is not backed up if boot history doesn't know about previous boot",
			setupMocks: func(pa *mocks.PreconditionsAdvisor, ba *mocks.BackupAdvisor, ex *mocks.Executor) {
				pa.EXPECT().IsOSTree().Return(true)
				pa.EXPECT().PreRunAlreadyRanCurrentBoot().Return(false)

				ba.EXPECT().DeviceBootedForTheFirstTime().Return(false)
				ba.EXPECT().BootHistoryExists().Return(true)
				ba.EXPECT().BootHistoryContainsPreviousBoot().Return(false)

				ex.EXPECT().UpdatePreRunStatus(history.PreRunSuccess).Return(nil)
			},
			errExpected: false,
		},
		{
			name: "Data is not backed up if previous boot wasn't healthy",
			setupMocks: func(pa *mocks.PreconditionsAdvisor, ba *mocks.BackupAdvisor, ex *mocks.Executor) {
				pa.EXPECT().IsOSTree().Return(true)
				pa.EXPECT().PreRunAlreadyRanCurrentBoot().Return(false)

				ba.EXPECT().DeviceBootedForTheFirstTime().Return(false)
				ba.EXPECT().BootHistoryExists().Return(true)
				ba.EXPECT().BootHistoryContainsPreviousBoot().Return(true)
				ba.EXPECT().PreviousBootWasHealthy().Return(false)

				ex.EXPECT().UpdatePreRunStatus(history.PreRunSuccess).Return(nil)
			},
			errExpected: false,
		},
		{
			name: "Data is backed up if previous boot was healthy",
			setupMocks: func(pa *mocks.PreconditionsAdvisor, ba *mocks.BackupAdvisor, ex *mocks.Executor) {
				pa.EXPECT().IsOSTree().Return(true)
				pa.EXPECT().PreRunAlreadyRanCurrentBoot().Return(false)

				ba.EXPECT().DeviceBootedForTheFirstTime().Return(false)
				ba.EXPECT().BootHistoryExists().Return(true)
				ba.EXPECT().BootHistoryContainsPreviousBoot().Return(true)
				ba.EXPECT().PreviousBootWasHealthy().Return(true)

				ex.EXPECT().BackupPreviousBoot().Return(nil)
				ex.EXPECT().UpdatePreRunStatus(history.PreRunSuccess).Return(nil)
			},
			errExpected: false,
		},
		{
			name: "If backing up wasn't successful, pre run result should be adequate",
			setupMocks: func(pa *mocks.PreconditionsAdvisor, ba *mocks.BackupAdvisor, ex *mocks.Executor) {
				pa.EXPECT().IsOSTree().Return(true)
				pa.EXPECT().PreRunAlreadyRanCurrentBoot().Return(false)

				ba.EXPECT().DeviceBootedForTheFirstTime().Return(false)
				ba.EXPECT().BootHistoryExists().Return(true)
				ba.EXPECT().BootHistoryContainsPreviousBoot().Return(true)
				ba.EXPECT().PreviousBootWasHealthy().Return(true)

				ex.EXPECT().BackupPreviousBoot().Return(errors.New("no reason"))
				ex.EXPECT().UpdatePreRunStatus(history.PreRunBackupFailed).Return(nil)
			},
			errExpected: true,
		},
	}

	for _, td := range testData {
		pa := &mocks.PreconditionsAdvisor{}
		ba := &mocks.BackupAdvisor{}
		ex := &mocks.Executor{}

		td.setupMocks(pa, ba, ex)

		prs := NewStrategy(pa, ba, ex)
		err := prs.Run()

		if td.errExpected {
			assert.Error(t, err, td.name)
		} else {
			assert.NoError(t, err, td.name)
		}

		ba.AssertExpectations(t)
		ex.AssertExpectations(t)
	}
}
