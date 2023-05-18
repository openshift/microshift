package admin

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"

	"k8s.io/klog/v2"
)

var (
	cpArgs = []string{
		"--verbose",
		"--recursive",
		"--preserve",
		"--reflink=auto",
	}
)

func MakeBackup(target, name string) error {
	if err := config.CreateDir(target); err != nil {
		klog.Errorf("Failed to create dir %s: %v", target, err)
		return err
	}

	dest := filepath.Join(target, name)
	tmp_dest := dest + "-tmp"

	args := append(cpArgs, config.DataDir, tmp_dest)
	out, err := exec.Command("cp", args...).CombinedOutput() //nolint:gosec
	klog.Infof("Output of cp (%v): %v", args, string(out))
	if err != nil {
		klog.Errorf("Failed to make backup - copy failed: %v", err)
		return err
	}

	if err := os.RemoveAll(dest); err != nil {
		klog.Errorf("Failed to remove %s directory: %v", tmp_dest, err)
		return err
	}

	if err := os.Rename(tmp_dest, dest); err != nil {
		klog.Errorf("Failed to rename %s to %s: %v", tmp_dest, dest, err)
		return err
	}
	klog.Infof("Backed %s to %s", config.DataDir, dest)
	return nil
}
