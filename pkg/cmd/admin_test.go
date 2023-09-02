package cmd

import (
	"testing"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/stretchr/testify/assert"
)

func Test_backupPathToStorageAndName(t *testing.T) {
	testData := []struct {
		name               string
		input              string
		errExpected        bool
		expectedStorage    data.StoragePath
		expectedBackupName data.BackupName
	}{
		{
			name:               "Absolute path with trailing slash",
			input:              "/var/lib/microshift-backups/custom-backup/",
			errExpected:        false,
			expectedStorage:    "/var/lib/microshift-backups",
			expectedBackupName: "custom-backup",
		},
		{
			name:               "Absolute path without trailing slash",
			input:              "/var/lib/microshift-backups/custom-backup",
			errExpected:        false,
			expectedStorage:    "/var/lib/microshift-backups",
			expectedBackupName: "custom-backup",
		},
		{
			name:               "Relative path with trailing slash",
			input:              "../microshift-backups/custom-backup/",
			errExpected:        false,
			expectedStorage:    "../microshift-backups",
			expectedBackupName: "custom-backup",
		},
		{
			name:               "Relative path without trailing slash",
			input:              "../microshift-backups/custom-backup",
			errExpected:        false,
			expectedStorage:    "../microshift-backups",
			expectedBackupName: "custom-backup",
		},
		{
			name:               "Storage is /",
			input:              "/custom-backup",
			errExpected:        false,
			expectedStorage:    "/",
			expectedBackupName: "custom-backup",
		},
		{
			name:               "Storage is .",
			input:              "custom-backup",
			errExpected:        false,
			expectedStorage:    ".",
			expectedBackupName: "custom-backup",
		},
		{
			name:        "Empty path",
			input:       "",
			errExpected: true,
		},
		{
			name:        "Backup is /",
			input:       "/",
			errExpected: true,
		},
		{
			name:               "Path ends with a dot",
			input:              "/var/lib/microshift-backups/bk/.",
			errExpected:        false,
			expectedStorage:    "/var/lib/microshift-backups",
			expectedBackupName: "bk",
		},
		{
			name:               "Path ends with a dot-dot",
			input:              "/var/lib/microshift-backups/bk/..",
			errExpected:        false,
			expectedStorage:    "/var/lib",
			expectedBackupName: "microshift-backups",
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			storage, backupName, err := backupPathToStorageAndName(td.input)
			if td.errExpected {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, td.expectedStorage, storage)
			assert.Equal(t, td.expectedBackupName, backupName)
		})
	}
}
