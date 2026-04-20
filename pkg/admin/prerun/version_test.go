package prerun

import (
	"encoding/json"
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
			name:        "binary must not be older than data",
			execVer:     versionMetadata{Major: 4, Minor: 14},
			dataVer:     versionMetadata{Major: 4, Minor: 15},
			errExpected: true,
		},
		{
			name:        "binary may be newer by one minor version",
			execVer:     versionMetadata{Major: 4, Minor: 15},
			dataVer:     versionMetadata{Major: 4, Minor: 14},
			errExpected: false,
		},
		{
			name:        "binary may be newer by two minor versions",
			execVer:     versionMetadata{Major: 4, Minor: 16},
			dataVer:     versionMetadata{Major: 4, Minor: 14},
			errExpected: false,
		},
		{
			name:        "binary must not be newer by more than 2 minor versions",
			execVer:     versionMetadata{Major: 4, Minor: 16},
			dataVer:     versionMetadata{Major: 4, Minor: 13},
			errExpected: true,
		},
		{
			name:        "cross-major upgrade 4.22 to 5.0 allowed (distance 1)",
			execVer:     versionMetadata{Major: 5, Minor: 0},
			dataVer:     versionMetadata{Major: 4, Minor: 22},
			errExpected: false,
		},
		{
			name:        "cross-major upgrade 4.21 to 5.0 allowed (distance 2)",
			execVer:     versionMetadata{Major: 5, Minor: 0},
			dataVer:     versionMetadata{Major: 4, Minor: 21},
			errExpected: false,
		},
		{
			name:        "cross-major upgrade 4.20 to 5.0 blocked (distance 3)",
			execVer:     versionMetadata{Major: 5, Minor: 0},
			dataVer:     versionMetadata{Major: 4, Minor: 20},
			errExpected: true,
		},
		{
			name:        "cross-major downgrade 5.0 to 4.22 blocked",
			execVer:     versionMetadata{Major: 4, Minor: 22},
			dataVer:     versionMetadata{Major: 5, Minor: 0},
			errExpected: true,
		},
		{
			name:        "cross-major upgrade 4.22 to 5.1 allowed (distance 2)",
			execVer:     versionMetadata{Major: 5, Minor: 1},
			dataVer:     versionMetadata{Major: 4, Minor: 22},
			errExpected: false,
		},
		{
			name:        "same major upgrade 5.0 to 5.1 allowed",
			execVer:     versionMetadata{Major: 5, Minor: 1},
			dataVer:     versionMetadata{Major: 5, Minor: 0},
			errExpected: false,
		},
	}

	for _, td := range testData {
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

func TestVersionFileSerialization(t *testing.T) {
	expected := `{"version":"4.14.0","deployment_id":"deploy-id","boot_id":"b-id"}`

	v := versionFile{
		Version:      versionMetadata{Major: 4, Minor: 14, Patch: 0},
		DeploymentID: "deploy-id",
		BootID:       "b-id",
	}

	serialized, err := json.Marshal(v)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(serialized))

	deserialized := &versionFile{}
	err = json.Unmarshal(serialized, deserialized)
	assert.NoError(t, err)
	assert.Equal(t, v, *deserialized)
}

func TestParseVersionFile(t *testing.T) {
	testData := []struct {
		input          string
		expectedResult versionFile
		errorExpected  bool
	}{
		{
			input: `{"version":"4.14.0","deployment_id":"deploy-id","boot_id":"b-id"}`,
			expectedResult: versionFile{
				Version:      versionMetadata{Major: 4, Minor: 14, Patch: 0},
				DeploymentID: "deploy-id",
				BootID:       "b-id",
			},
			errorExpected: false,
		},
		{
			input: `4.14.0`,
			expectedResult: versionFile{
				Version:      versionMetadata{Major: 4, Minor: 14, Patch: 0},
				DeploymentID: "",
				BootID:       "",
			},
			errorExpected: false,
		},
		{
			input:         `4.14.`,
			errorExpected: true,
		},
		{
			input:         `4.14`,
			errorExpected: true,
		},
		{
			input:         `4`,
			errorExpected: true,
		},
		{
			input:         ``,
			errorExpected: true,
		},
	}

	for _, td := range testData {
		out, err := parseVersionFile([]byte(td.input))
		if td.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, td.expectedResult, out)
		}
	}
}
