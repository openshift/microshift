package prerun

import (
	"testing"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/stretchr/testify/assert"
)

func Test_BackupName(t *testing.T) {
	testData := []struct {
		healthInfo         HealthInfo
		expectedBackupName string
	}{
		{
			healthInfo: HealthInfo{
				Health:       "healthy",
				DeploymentID: "did",
				BootID:       "bid",
			},
			expectedBackupName: "did_bid",
		},
		{
			healthInfo: HealthInfo{
				Health:       "unhealthy",
				DeploymentID: "did",
				BootID:       "bid",
			},
			expectedBackupName: "did_bid_unhealthy",
		},
	}

	for _, td := range testData {
		assert.Equal(t, data.BackupName(td.expectedBackupName), td.healthInfo.BackupName())
	}
}
