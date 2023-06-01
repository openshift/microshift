package prerun

import (
	"errors"
	"fmt"

	"github.com/openshift/microshift/pkg/admin/history"

	"k8s.io/klog/v2"
)

func NewStrategy(pa PreconditionsAdvisor, ba BackupAdvisor, ex Executor) *Strategy {
	return &Strategy{
		pa: pa,
		ba: ba,
		ex: ex,
	}
}

type Strategy struct {
	pa PreconditionsAdvisor
	ba BackupAdvisor
	ex Executor
}

func (a *Strategy) Run() (err error) {
	if !a.pa.IsOSTree() {
		klog.Info("Non ostree-based system - pre-run will not run")
		return nil
	}
	klog.Info("System is ostree-based - executing pre-run")

	if a.pa.PreRunAlreadyRanCurrentBoot() {
		if a.pa.PreRunWasSuccessful() {
			klog.InfoS("Pre run already successfully ran during current boot")
			return nil
		}
		err = fmt.Errorf("pre-run already ran during current boot and was unsuccessful (%s)", a.pa.PreRunStatus())
		return
	}

	klog.Info("Starting pre-run procedure")

	preRunStatus := history.PreRunUnknown
	defer func() {
		klog.InfoS("Pre-run procedure finished", "status", preRunStatus)
		err = errors.Join(err, a.ex.UpdatePreRunStatus(preRunStatus))
	}()

	err = a.backupPreviousBoot()
	if err != nil {
		preRunStatus = history.PreRunBackupFailed
		return
	}

	// TODO: Restore backup compatible with current deployment
	// TODO: Migrate data
	// TODO: Update version metadata

	preRunStatus = history.PreRunSuccess
	return
}

func (a *Strategy) backupPreviousBoot() error {
	klog.Info("Running checks if MicroShift data should be backed up")

	if a.ba.DeviceBootedForTheFirstTime() {
		klog.InfoS("This is first boot of the device (no information about previous boot in systemd) - not backing up the data")
		return nil
	}

	if !a.ba.BootHistoryExists() {
		klog.InfoS("Boot history does not exist - not backing up the data")
		// TODO: No history, check for data existence
		return nil
	}

	if !a.ba.BootHistoryContainsPreviousBoot() {
		klog.InfoS("Boot history does not contain information about previous boot - not backing up the data")
		// TODO: History exists, but missing history entry for previous boot
		return nil
	}

	if a.ba.PreviousBootWasHealthy() {
		klog.InfoS("Previous boot was healthy - backing up the data")
		return a.ex.BackupPreviousBoot()
	}

	return nil
}
