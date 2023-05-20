package admin

import (
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

func microshiftShouldNotRun() error {
	cmd := exec.Command("sh", "-c", "ps aux | grep  -v grep | grep 'microshift run'")
	out, err := cmd.CombinedOutput()
	sout := strings.TrimSpace(string(out))

	if err == nil {
		klog.InfoS("MicroShift is running", "process", sout)
		return fmt.Errorf("command requires that MicroShift must not run")
	}
	if err != nil && sout != "" {
		klog.ErrorS(err, "Unexpected error when checking if MicroShift is running already", "output", sout)
		return fmt.Errorf("checking if MicroShift is running failed: %w", err)
	}
	return nil
}
