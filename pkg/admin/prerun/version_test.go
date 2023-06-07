package prerun

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckVersionDiff(t *testing.T) {

	testData := []struct {
		name                      string
		execVer                   versionMetadata
		dataVer                   versionMetadata
		expectedMigrationRequired bool
		errExpected               bool
	}{
		{
			name:                      "equal versions: no migration, no error",
			execVer:                   versionMetadata{X: 4, Y: 14},
			dataVer:                   versionMetadata{X: 4, Y: 14},
			expectedMigrationRequired: false,
			errExpected:               false,
		},
		{
			name:                      "X versions must be the same",
			execVer:                   versionMetadata{X: 4, Y: 14},
			dataVer:                   versionMetadata{X: 5, Y: 14},
			expectedMigrationRequired: false,
			errExpected:               true,
		},
		{
			name:                      "binary must not be older than data",
			execVer:                   versionMetadata{X: 4, Y: 14},
			dataVer:                   versionMetadata{X: 4, Y: 15},
			expectedMigrationRequired: false,
			errExpected:               true,
		},
	}

	for _, td := range testData {
		td := td
		t.Run(td.name, func(t *testing.T) {
			migrationRequired, err := checkVersionDiff(td.execVer, td.dataVer)

			assert.Equal(t, td.expectedMigrationRequired, migrationRequired)
			if td.errExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
