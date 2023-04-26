package ostree

import (
	"k8s.io/klog/v2"
)

func PreRun() error {
	klog.Info("Pre run procedure starting")

	action, err := nextBootFromDisk()
	if err != nil {
		klog.Errorf("Failed to get boot action: %v", err)
		return err
	}
	klog.Info("Boot action: %v", action)

	return nil
}
