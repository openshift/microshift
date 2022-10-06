package components

import (
	"github.com/openshift/microshift/pkg/config"
	"k8s.io/klog/v2"
)

func StartComponents(cfg *config.MicroshiftConfig) error {
	kubeAdminConfig := cfg.KubeConfigPath(config.KubeAdmin)

	if err := startServiceCAController(cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start service-ca controller: %v", err)
		return err
	}

	if err := startCSIPlugin(cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start csi plugin: %v", err)
		return err
	}

	if err := startCSISnapshot(cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start CSI snapshot controller: %v", err)
		return err
	}

	if err := startIngressController(cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start ingress router controller: %v", err)
		return err
	}
	if err := startDNSController(cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start DNS controller: %v", err)
		return err
	}

	if err := startOVNKubernetes(cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start OVNKubernetes: %v", err)
		return err
	}

	return nil
}
