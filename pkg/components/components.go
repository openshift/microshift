package components

import (
	"github.com/sirupsen/logrus"
)

func StartComponents() error {
	if err := startServiceCAController(); err != nil {
		logrus.Warningf("failed to start service-ca controller: %v", err)
		return err
	}

	if err := startHostpathProvisioner(); err != nil {
		logrus.Warningf("failed to start hostpath provisioner: %v", err)
		return err
	}

	if err := startIngressController(); err != nil {
		logrus.Warningf("failed to start ingress router controller: %v", err)
		return err
	}
	if err := startDNSController(); err != nil {
		logrus.Warningf("failed to start DNS controller: %v", err)
		return err
	}
	if err := startFlannel(); err != nil {
		logrus.Warning("failed to start Flannel: %v", err)
		return err
	}
	return nil
}
