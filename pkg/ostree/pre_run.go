package ostree

import (
	"k8s.io/klog/v2"
)

func PreRun() error {
	klog.Info("Pre run procedure starting")

	return nil
}
