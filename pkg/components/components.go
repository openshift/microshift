package components

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/openshift/microshift/pkg/util"
)

func StartComponents() error {
	if err := startHostpathProvisioner(); err != nil {
		logrus.Warningf("failed to start hostpath provisioner: %v", err)
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %v", err)
	}

	ip, err := util.GetHostIP()
	if err != nil {
		return fmt.Errorf("failed to get host IP: %v", err)
	}
	if err := util.GenCerts("service-ca", "/etc/kubernetes/ushift-resources/service-ca/secrets/service-ca",
		"tls.crt", "tls.key",
		[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
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
