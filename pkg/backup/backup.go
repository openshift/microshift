package backup

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	"k8s.io/klog/v2"
)

const (
	latestBackupDir = "latest"
)

func MakeBackup(subdir string) error {
	if err := config.EnsureAuxDirExists(); err != nil {
		klog.Errorf("Failed to make aux data dir: %v", err)
		return err
	}

	dest := filepath.Join(config.AuxDataDir, subdir)
	cmd := exec.Command("cp", "--recursive", "--reflink=auto", config.DataDir, dest) //nolint:gosec

	if out, err := cmd.CombinedOutput(); err != nil {
		klog.Errorf("Failed to make backup - copy failed: %s err: %v", out, err)
		return err
	}

	link := filepath.Join(config.AuxDataDir, latestBackupDir)
	if err := os.Remove(link); err != nil {
		klog.Errorf("Failed to remove old symlink: %v", link, err)
		return err
	}
	if err := os.Symlink(dest, link); err != nil {
		klog.Errorf("Failed to make symlink: %v", link, err)
		return err
	}

	klog.Infof("Backed up data to %s", dest)
	return nil
}
