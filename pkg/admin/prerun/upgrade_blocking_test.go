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
		dataVersion    string
		execVersion    string
		expectedResult bool
	}{
		{
			dataVersion:    "4.14.4",
			execVersion:    "4.14.10",
			expectedResult: false,
		},
		{
			dataVersion:    "4.14.5",
			execVersion:    "4.14.10",
			expectedResult: true,
		},
		{
			dataVersion:    "4.14.6",
			execVersion:    "4.14.10",
			expectedResult: true,
		},
		{
			dataVersion:    "4.14.7",
			execVersion:    "4.14.10",
			expectedResult: false,
		},
		{
			dataVersion:    "4.14.7",
			execVersion:    "4.15.0",
			expectedResult: false,
		},
		{
			dataVersion:    "4.15.2",
			execVersion:    "4.15.5",
			expectedResult: true,
		},
	}

	for _, td := range testData {
		assert.Equal(t, td.expectedResult, isBlocked(edges, td.execVersion, td.dataVersion))
	}
}
