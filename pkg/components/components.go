package components

import (
	"context"

	"github.com/openshift/microshift/pkg/config"
	"k8s.io/klog/v2"
)

func StartComponents(cfg *config.Config, ctx context.Context) error {
	kubeAdminConfig := cfg.KubeConfigPath(config.KubeAdmin)

	if err := startServiceCAController(ctx, cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start service-ca controller: %v", err)
		return err
	}

	if err := startCSISnapshotController(ctx, cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start csi snapshot controller: %v", err)
		return err
	}

	if err := startCSIPlugin(ctx, cfg, cfg.KubeConfigPath(config.KubeAdmin)); err != nil {
		klog.Warningf("Failed to start csi plugin: %v", err)
		return err
	}

	if err := startIngressController(ctx, cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start ingress router controller: %v", err)
		return err
	}
	if err := startDNSController(ctx, cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start DNS controller: %v", err)
		return err
	}

	if err := startCNIPlugin(ctx, cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start CNI plugin: %v", err)
		return err
	}

	if err := deployMultus(ctx, cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to deploy Multus CNI: %v", err)
		return err
	}

	return nil
}
