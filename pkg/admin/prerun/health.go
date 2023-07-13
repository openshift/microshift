package prerun

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/util"
)

var (
	healthFilepath            = "/var/lib/microshift-backups/health.json"
	errHealthFileDoesNotExist = errors.New("health file does not exist")
)

type HealthInfo struct {
	Health       string `json:"health"`
	DeploymentID string `json:"deployment_id"`
	BootID       string `json:"boot_id"`
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
		return nil, fmt.Errorf("failed to read health data from %q: %w", healthFilepath, err)
	}

	health := &HealthInfo{}
	if err := json.Unmarshal(content, &health); err != nil {
		return nil, fmt.Errorf("failed to parse health data %q: %w", strings.TrimSpace(string(content)), err)
	}
	return health, nil
}
