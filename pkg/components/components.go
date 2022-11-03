package components

import (
	"github.com/openshift/microshift/pkg/config"
	"k8s.io/klog/v2"
)

var microshiftDataDir = config.GetDataDir()

func StartComponents(cfg *config.MicroshiftConfig) error {
	kubeAdminConfig := cfg.KubeConfigPath(config.KubeAdmin)

	if err := startServiceCAController(cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start service-ca controller: %v", err)
		return err
	}

	if err := startCSIPlugin(cfg, cfg.KubeConfigPath(config.KubeAdmin)); err != nil {
		klog.Warningf("Failed to start csi plugin: %v", err)
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

	if err := startCNIPlugin(cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to start CNI plugin: %v", err)
		return err
	}
	return nil
}
