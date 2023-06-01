package data

import (
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

var (
	wrongStates = []string{"active", "activating", "deactivating"}
	services    = []string{"microshift.service", "microshift-etcd.scope"}
)

func MicroShiftIsNotRunning() error {
	for _, service := range services {
		cmd := exec.Command("systemctl", "show", "-p", "ActiveState", "--value", service)
		out, err := cmd.CombinedOutput()
		state := strings.TrimSpace(string(out))
		if err != nil {
			return fmt.Errorf("error when checking if %s is active: %w", service, err)
		}

		klog.InfoS("Service state", "service", service, "state", state)

		for _, s := range wrongStates {
			if state == s {
				return fmt.Errorf("service %s is %s", service, state)
			}
		}
	}

	return nil
}
