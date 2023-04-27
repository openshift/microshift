package ostree

import (
	"fmt"
	"os/exec"

	"github.com/openshift/microshift/pkg/backup"
	"k8s.io/klog/v2"
)

func microshiftShouldNotRun() error {
	cmd := exec.Command("sh", "-c", "ps aux | grep  -v grep | grep 'microshift run'")
	out, err := cmd.CombinedOutput()
	sout := string(out)

	if err == nil {
		klog.Infof("Detected that MicroShift is already running: %v", sout)
		return fmt.Errorf("command requires that MicroShift must not run")
	}
	if err != nil && sout != "" {
		klog.Info("Unexpected error when checking if MicroShift is running already: %v", err)
		return err
	}
	return nil
}

func PreRun() error {
	klog.Info("Pre-run procedure starting")

	if err := microshiftShouldNotRun(); err != nil {
		return err
	}

	action, err := preRunActionFromDisk()
	if err != nil {
		klog.Errorf("Failed to get pre-run-action: %v", err)
		return err
	}

	switch action.Action {
	case actionBackup:
		if err := backup.MakeBackup(action.OstreeID); err != nil {
			return err
		}
	case actionRestore:
		return fmt.Errorf("not implemented")
	case actionMissing:
		klog.Infof("Pre-run-action file is missing")
	}

	if err := action.RemoveFromDisk(); err != nil {
		klog.Errorf("Failed to remove pre-run-action file: %v", err)
		return err
	}

	return nil
}
