package data

import (
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

const (
	expectedState = "inactive"
)

var (
	services = []string{"microshift.service", "microshift-etcd.scope"}
)

func microshiftIsNotRunning() error {
	for _, service := range services {
		cmd := exec.Command("systemctl", "show", "-p", "ActiveState", "--value", service)
		out, err := cmd.CombinedOutput()
		state := strings.TrimSpace(string(out))
		if err != nil {
			return fmt.Errorf("error when checking if %s is active: %w", service, err)
		}

		klog.V(2).InfoS("Service state", "service", service, "state", state)

		if state != expectedState {
			return fmt.Errorf("service %s is %s - expected to be %s", service, state, expectedState)
		}
	}

	return nil
}
