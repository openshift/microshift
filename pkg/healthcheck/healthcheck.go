package healthcheck

import (
	"context"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

func MicroShiftHealthcheck(ctx context.Context, timeout time.Duration) error {
	if enabled, err := microshiftServiceShouldBeOk(ctx, timeout); err != nil {
		printPrerunLog()
		return err
	} else if !enabled {
		return nil
	}

	if err := waitForCoreWorkloads(timeout); err != nil {
		logPodsAndEvents()
		return err
	}

	klog.Info("MicroShift is ready")

	return nil
}

func NamespacesHealthcheck(ctx context.Context, timeout time.Duration, namespaces []string) error {
	if err := waitForNamespaces(timeout, namespaces); err != nil {
		logPodsAndEvents()
		return err
	}
	klog.Infof("Workloads in %s namespaces are ready", strings.Join(namespaces, ", "))
	return nil
}
