package testutil

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"

	"k8s.io/klog/v2"
)

func RunCommand(c ...string) (string, string, error) {
	cmd := exec.Command(c[0], c[1:]...)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	klog.V(2).InfoS("Running command", "cmd", cmd)
	err := cmd.Run()
	out := strings.Trim(outb.String(), "\n")
	serr := errb.String()
	klog.InfoS("Command complete", "cmd", cmd, "stdout", redactOutput(out), "stderr", redactOutput(serr), "err", err)
	if err != nil {
		return "", "", err
	}

	return out, serr, nil
}

// redactOutput overwrites sensitive data when logging command outputs
func redactOutput(output string) string {
	rx := regexp.MustCompile("gpgkeys.*")
	return rx.ReplaceAllString(output, "gpgkeys = REDACTED")
}
