package components

import (
	"context"

	"github.com/sirupsen/logrus"
)

func StartComponents() error {
	if err := startHostpathProvisioner(); err != nil {
		logrus.Warningf("failed to start hostpath provisioner: %v", err)
		return err
	}
	// create service-ca cert
	if dir, err := genCerts("service-ca", "/var/lib/openshift/service-ca/key",
		"tls.crt", "tls.key"); err != nil {
		logrus.Warningf("failed to create service-ca svc cert: %v", err)
		return err
	} else {
		openshift.ServiceCAKeyDir = dir
	}
	// create service-ca ca bundle
	if dir, err := genCerts("service-ca-signer", "/var/lib/openshift/service-ca/ca-cabundle",
		"ca-bundle.crt", "ca-bundle.key"); err != nil {
		logrus.Warningf("failed to create service-ca ca bundle cert: %v", err)
		return err
	}
	if err := startServiceCAController(); err != nil {
		logrus.Warningf("failed to start service-ca controller: %v", err)
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
	return nil
}
