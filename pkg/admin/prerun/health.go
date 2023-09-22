package prerun

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"

	"k8s.io/klog/v2"
)

var (
	healthFilepath            = filepath.Join(config.BackupsDir, "health.json")
	errHealthFileDoesNotExist = errors.New("health file does not exist")
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
	// We should be able to remove the health.json file in 4.16 or even 4.15
	// if we state that upgrading from rc.0 and rc.1 directly to 4.15 is not supported,
	// but should be done with intermediate step to released 4.14.0.

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

func (hi *HealthInfo) BackupName() data.BackupName {
	name := fmt.Sprintf("%s_%s", hi.DeploymentID, hi.BootID)

	if hi.IsHealthy() {
		return data.BackupName(name)
	}

	return data.BackupName(fmt.Sprintf("%s_unhealthy", name))
}

func (hi *HealthInfo) IsHealthy() bool {
	return hi.Health == "healthy"
}

func getHealthInfo() (*HealthInfo, error) {
	if exists, err := util.PathExistsAndIsNotEmpty(healthFilepath); err != nil {
		return nil, err
	} else if !exists {
		return nil, errHealthFileDoesNotExist
	}

	content, err := os.ReadFile(healthFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", healthFilepath, err)
	}

	health := &HealthInfo{}
	if err := json.Unmarshal(content, &health); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q: %w", strings.TrimSpace(string(content)), err)
	}

	klog.InfoS("Read health info from file",
		"health", health.Health,
		"deploymentID", health.DeploymentID,
		"previousBootID", health.BootID,
	)

	return health, nil
}
