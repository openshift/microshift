package healthcheck

import (
	"context"
	"encoding/json"
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

func CustomWorkloadHealthcheck(ctx context.Context, timeout time.Duration, definition string) error {
	workloads := map[string]NamespaceWorkloads{}

	err := json.Unmarshal([]byte(definition), &workloads)
	if err != nil {
		return err
	}
	klog.V(2).Infof("Deserialized '%s' into %+v", definition, workloads)

	if err := waitForWorkloads(timeout, workloads); err != nil {
		logPodsAndEvents()
		return err
	}
	return nil
}

func EasyCustomWorkloadHealthcheck(ctx context.Context, timeout time.Duration, namespace string, deployments, daemonsets, statefulsets []string) error {
	workloads := map[string]NamespaceWorkloads{
		namespace: {
			Deployments:  deployments,
			DaemonSets:   daemonsets,
			StatefulSets: statefulsets,
		},
	}
	if err := waitForWorkloads(timeout, workloads); err != nil {
		logPodsAndEvents()
		return err
	}
	return nil
}
