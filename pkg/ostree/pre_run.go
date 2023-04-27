package ostree

import (
	"fmt"

	"github.com/openshift/microshift/pkg/backup"
	"k8s.io/klog/v2"
)

func PreRun() error {
	klog.Info("Pre run procedure starting")

	action, err := nextBootFromDisk()
	if err != nil {
		klog.Errorf("Failed to get boot action: %v", err)
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
		klog.Infof("Next boot file is missing")
	}

	if err := action.RemoveFromDisk(); err != nil {
		klog.Errorf("Failed to remove next boot file: %v", err)
		return err
	}

	return nil
}
