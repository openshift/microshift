package testutil

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

func RunCommand(c ...string) (string, string, error) {
	cmd := exec.Command(c[0], c[1:]...)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	klog.InfoS("Running command", "cmd", cmd)
	start := time.Now()
	err := cmd.Run()
	dur := time.Since(start)

	out := strings.Trim(outb.String(), "\n")
	serr := errb.String()
	klog.InfoS("Command complete", "duration", dur, "cmd", cmd, "stdout", redactOutput(out), "stderr", redactOutput(serr), "err", err)

	return out, serr, err
}

// redactOutput overwrites sensitive data when logging command outputs
func redactOutput(output string) string {
	rx := regexp.MustCompile("gpgkeys.*")
	return rx.ReplaceAllString(output, "gpgkeys = REDACTED")
}
