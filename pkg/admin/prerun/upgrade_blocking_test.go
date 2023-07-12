package prerun

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UnmarshalBlockedUpgrades(t *testing.T) {
	testData := []struct {
		input          string
		expectedOutput map[string][]string
	}{
		{
			input:          `{"4.14.10": ["4.14.5", "4.14.4"]}`,
			expectedOutput: map[string][]string{"4.14.10": {"4.14.5", "4.14.4"}},
		},
		{
			input:          `{}`,
			expectedOutput: make(map[string][]string),
		},
	}

	for _, td := range testData {
		result, err := unmarshalBlockedUpgrades([]byte(td.input))
		assert.NoError(t, err)
		assert.Equal(t, td.expectedOutput, result)
	}
}

func Test_IsBlocked(t *testing.T) {
	edges := map[string][]string{
		"4.14.10": {"4.14.5", "4.14.6"},
		"4.15.5":  {"4.15.2"},
	}

	testData := []struct {
		dataVersion string
		execVersion string
		errExpected bool
	}{
		{
			dataVersion: "4.14.4",
			execVersion: "4.14.10",
			errExpected: false,
		},
		{
			dataVersion: "4.14.5",
			execVersion: "4.14.10",
			errExpected: true,
		},
		{
			dataVersion: "4.14.6",
			execVersion: "4.14.10",
			errExpected: true,
		},
		{
			dataVersion: "4.14.7",
			execVersion: "4.14.10",
			errExpected: false,
		},
		{
			dataVersion: "4.14.7",
			execVersion: "4.15.0",
			errExpected: false,
		},
		{
			dataVersion: "4.15.2",
			execVersion: "4.15.5",
			errExpected: true,
		},
	}

	for _, td := range testData {
		if td.errExpected {
			assert.Error(t, isBlocked(edges, td.execVersion, td.dataVersion))
		} else {
			assert.NoError(t, isBlocked(edges, td.execVersion, td.dataVersion))
		}
	}
}
