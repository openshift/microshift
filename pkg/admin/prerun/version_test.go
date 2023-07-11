package prerun

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckVersionDiff(t *testing.T) {
	testData := []struct {
		name        string
		execVer     versionMetadata
		dataVer     versionMetadata
		errExpected bool
	}{
		{
			name:        "equal versions: no migration, no error",
			execVer:     versionMetadata{Major: 4, Minor: 14},
			dataVer:     versionMetadata{Major: 4, Minor: 14},
			errExpected: false,
		},
		{
			name:        "X versions must be the same",
			execVer:     versionMetadata{Major: 4, Minor: 14},
			dataVer:     versionMetadata{Major: 5, Minor: 14},
			errExpected: true,
		},
		{
			name:        "binary must not be older than data",
			execVer:     versionMetadata{Major: 4, Minor: 14},
			dataVer:     versionMetadata{Major: 4, Minor: 15},
			errExpected: true,
		},
		{
			name:        "binary must be newer only by one minor version",
			execVer:     versionMetadata{Major: 4, Minor: 15},
			dataVer:     versionMetadata{Major: 4, Minor: 14},
			errExpected: false,
		},
		{
			name:        "binary newer more than one minor version is not supported",
			execVer:     versionMetadata{Major: 4, Minor: 15},
			dataVer:     versionMetadata{Major: 4, Minor: 13},
			errExpected: true,
		},
	}

	for _, td := range testData {
		td := td
		t.Run(td.name, func(t *testing.T) {
			err := checkVersionCompatibility(td.execVer, td.dataVer)

			if td.errExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
