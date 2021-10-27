package components

import (
	"github.com/openshift/microshift/pkg/config"
	"github.com/sirupsen/logrus"
)

func StartComponents(cfg *config.MicroshiftConfig) error {
	if err := startServiceCAController(cfg, cfg.DataDir+"/resources/kubeadmin/kubeconfig"); err != nil {
		logrus.Warningf("failed to start service-ca controller: %v", err)
		return err
	}

	if err := startHostpathProvisioner(cfg.DataDir + "/resources/kubeadmin/kubeconfig"); err != nil {
		logrus.Warningf("failed to start hostpath provisioner: %v", err)
		return err
	}

	if err := startIngressController(cfg, cfg.DataDir+"/resources/kubeadmin/kubeconfig"); err != nil {
		logrus.Warningf("failed to start ingress router controller: %v", err)
		return err
	}
	if err := startDNSController(cfg, cfg.DataDir+"/resources/kubeadmin/kubeconfig"); err != nil {
		logrus.Warningf("failed to start DNS controller: %v", err)
		return err
	}
	if err := startFlannel(cfg.DataDir + "/resources/kubeadmin/kubeconfig"); err != nil {
		logrus.Warningf("failed to start Flannel: %v", err)
		return err
	}
	return nil
}
