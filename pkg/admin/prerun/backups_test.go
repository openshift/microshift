package prerun

import (
	"testing"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/stretchr/testify/assert"
)

func Test_isAutomatedBackup(t *testing.T) {
	testData := []struct {
		name        string
		backupName  data.BackupName
		isAutomated bool
	}{
		{
			name:        "Correct backup name",
			backupName:  data.BackupName("rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0_80364fcf3df54284a6902687e2cdd4c2"),
			isAutomated: true,
		},
		{
			name:        "Serial number is missing at the end of deployment ID",
			backupName:  data.BackupName("rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048._80364fcf3df54284a6902687e2cdd4c2"),
			isAutomated: false,
		},
		{
			name:        "Serial number is double digit",
			backupName:  data.BackupName("rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.11_80364fcf3df54284a6902687e2cdd4c2"),
			isAutomated: true,
		},
		{
			name:        "Serial number is triple digit",
			backupName:  data.BackupName("rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.111_80364fcf3df54284a6902687e2cdd4c2"),
			isAutomated: true,
		},
		{
			name:        "Deployment checksum is one char short",
			backupName:  data.BackupName("rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf9904.0_80364fcf3df54284a6902687e2cdd4c2"),
			isAutomated: false,
		},
		{
			name:        "Boot ID is one char short",
			backupName:  data.BackupName("rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0_80364fcf3df54284a6902687e2cdd4c"),
			isAutomated: false,
		},
		{
			name:        "Different osname",
			backupName:  data.BackupName("fedora-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0_80364fcf3df54284a6902687e2cdd4c2"),
			isAutomated: true,
		},
		{
			name:        "Different osname with extra dash",
			backupName:  data.BackupName("fedora-silverblue-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0_80364fcf3df54284a6902687e2cdd4c2"),
			isAutomated: true,
		},
		{
			name:        "Former default manual name",
			backupName:  data.BackupName("4.14.0_20230815000000"),
			isAutomated: false,
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			assert.Equal(t, td.isAutomated, isAutomatedBackup(td.backupName))
		})
	}
}

func Test_getDeploymentIDForTheBackup(t *testing.T) {
	testData := []struct {
		backupName     data.BackupName
		expectedResult string
	}{
		{
			backupName:     data.BackupName("rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0_80364fcf3df54284a6902687e2cdd4c2"),
			expectedResult: "rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0",
		},
		{
			backupName:     data.BackupName("fedora-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0_80364fcf3df54284a6902687e2cdd4c2"),
			expectedResult: "fedora-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0",
		},
		{
			backupName:     data.BackupName("fedora-silverblue-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0_80364fcf3df54284a6902687e2cdd4c2"),
			expectedResult: "fedora-silverblue-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0",
		},
		{
			backupName:     data.BackupName("r-35d7b5c80f0f1378d6846f6dc1304bbf.0_80364fcf3df54284"),
			expectedResult: "",
		},
		{
			backupName:     data.BackupName("4.14.0_20230815000000"),
			expectedResult: "",
		},
		{
			backupName:     data.BackupName("custom001"),
			expectedResult: "",
		},
		{
			backupName:     data.BackupName("custom_001"),
			expectedResult: "",
		},
	}

	for _, td := range testData {
		assert.Equal(t, td.expectedResult, getDeploymentIDForTheBackup(td.backupName))
	}
}
