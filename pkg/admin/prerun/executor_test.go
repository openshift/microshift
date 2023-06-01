package prerun

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/admin/data/mocks"
	"github.com/openshift/microshift/pkg/admin/history"
	"github.com/openshift/microshift/pkg/admin/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Backup(t *testing.T) {
	did := "system-abcd.0"
	dd := decisionData{
		PreviousBootInfo: &history.Boot{
			DeploymentBoot: history.NewDeploymentBoot(system.Boot{}, system.DeploymentID(did)),
			BootInfo:       history.BootInfo{},
		},
	}
	dm := &mocks.Manager{}

	existingBackups := []data.BackupName{
		data.BackupName(fmt.Sprintf("%s_0", did)),
		data.BackupName(fmt.Sprintf("%s_1", did)),
	}
	dm.EXPECT().GetBackupList().Return(existingBackups, nil)

	dm.EXPECT().Backup(mock.MatchedBy(func(n data.BackupName) bool {
		expectedPrefix := fmt.Sprintf("%s_%s", did, time.Now().UTC().Format("20060102_"))
		return strings.Contains(string(n), expectedPrefix)
	})).Return(nil)

	removedBackups := []data.BackupName{}
	dm.EXPECT().RemoveBackup(mock.MatchedBy(func(n data.BackupName) bool {
		removedBackups = append(removedBackups, n)
		return true
	})).Return(nil)

	ex := &executor{dataManager: dm, decisionData: dd}

	err := ex.BackupPreviousBoot()
	assert.NoError(t, err)
	assert.Subset(t, existingBackups, removedBackups)
	assert.Subset(t, removedBackups, existingBackups)
}
