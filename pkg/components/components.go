package components

import (
	"time"

	"github.com/openshift/microshift/pkg/config"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/util/iptables"
)

const iptablesCheckInterval = time.Second * 5

var microshiftDataDir = config.GetDataDir()

func StartComponents(cfg *config.MicroshiftConfig, iptClients []iptables.Interface) error {
	kubeAdminConfig := cfg.KubeConfigPath(config.KubeAdmin)

	for i := range iptClients {
		iptClient := iptClients[i]
		go iptClient.Monitor(
			iptables.Chain("MICROSHIFT-CANARY"),
			[]iptables.Table{iptables.TableMangle, iptables.TableNAT, iptables.TableFilter},
			func() {
				klog.Warningf("Iptables flush is detected, reloading affected components")
				if err := reloadOnIptableFlush(cfg); err != nil {
					klog.Errorf("Failed to reload affected components: %v", err)
				}
			},
			iptablesCheckInterval,
			wait.NeverStop,
		)
	}

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

func reloadOnIptableFlush(cfg *config.MicroshiftConfig) error {
	kubeAdminConfig := cfg.KubeConfigPath(config.KubeAdmin)

	klog.Infof("Reload ingress controller")
	if err := startIngressController(cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to reload ingress router controller: %v", err)
		return err
	}
	klog.Infof("Reload CNI plugin")
	if err := startCNIPlugin(cfg, kubeAdminConfig); err != nil {
		klog.Warningf("Failed to reload CNI plugin: %v", err)
		return err
	}
	return nil
}
