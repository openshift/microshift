package prerun

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"

	"k8s.io/klog/v2"
)

var (
	healthFilepath = filepath.Join(config.BackupsDir, "health.json")
)

type HealthInfo struct {
	Health       string `json:"health"`
	DeploymentID string `json:"deployment_id"`
	BootID       string `json:"boot_id"`
}

// updateHealthInfo updates health.json files with hardcoded "healthy" and
// deployment and boot IDs from provided argument.
func updateHealthInfo(vf versionFile) error {
	// health.json in kept in place to support rollback to 4.14.0~rc.1 and earlier,
	// as lack of the file would result in MicroShift neither backing up the data
	// nor restoring a backup and we want to always create a backup.
	// We should be able to remove the health.json file when 4.14 is no longer supported.

	hi := HealthInfo{
		Health:       "healthy",
		DeploymentID: vf.DeploymentID,
		BootID:       vf.BootID,
	}
	klog.InfoS("Updating health.json", "contents", hi)

	data, err := json.Marshal(hi)
	if err != nil {
		return fmt.Errorf("failed to marshal %v: %w", hi, err)
	}

	if err := os.WriteFile(healthFilepath, data, 0600); err != nil {
		return fmt.Errorf("writing %q to %q failed: %w", string(data), healthFilepath, err)
	}

	return nil
}
