package healthcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_linesToMap(t *testing.T) {
	testData := []struct {
		name        string
		lines       []string
		expectedMap map[string]string
	}{
		{
			name: "No comments",
			lines: []string{
				"MICROSHIFT_WAIT_TIMEOUT_SEC=100",
				"GREENBOOT_MAX_BOOT_ATTEMPTS=5",
			},
			expectedMap: map[string]string{
				"MICROSHIFT_WAIT_TIMEOUT_SEC": "100",
				"GREENBOOT_MAX_BOOT_ATTEMPTS": "5",
			},
		},
		{
			name: "Leading comments",
			lines: []string{
				"#MICROSHIFT_WAIT_TIMEOUT_SEC=100",
				"# GREENBOOT_MAX_BOOT_ATTEMPTS=5",
			},
			expectedMap: map[string]string{},
		},
		{
			name: "Trailing comments",
			lines: []string{
				"MICROSHIFT_WAIT_TIMEOUT_SEC=200 # Some comment",
				"GREENBOOT_MAX_BOOT_ATTEMPTS=7 # Some other comment",
			},
			expectedMap: map[string]string{
				"MICROSHIFT_WAIT_TIMEOUT_SEC": "200",
				"GREENBOOT_MAX_BOOT_ATTEMPTS": "7",
			},
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			m := bashlikeVarsToMap(td.lines, "MICROSHIFT_", "GREENBOOT_")
			assert.Equal(t, td.expectedMap, m)
		})
	}
}

func Test_getIntValueFromMap(t *testing.T) {
	{
		m := map[string]string{
			"menu_auto_hide":     "1",
			"boot_success":       "0",
			"boot_counter":       "2",
			"boot_indeterminate": "0",
		}
		val := getIntValueFromMap(m, "boot_counter", 5)
		assert.Equal(t, 2, val)
	}

	{
		m := map[string]string{
			"GREENBOOT_MAX_BOOT_ATTEMPTS":      "5",
			"GREENBOOT_WATCHDOG_CHECK_ENABLED": "true",
			"MICROSHIFT_WAIT_TIMEOUT_SEC":      "123",
		}
		val := getIntValueFromMap(m, "MICROSHIFT_WAIT_TIMEOUT_SEC", 300)
		assert.Equal(t, 123, val)
	}

	{
		m := map[string]string{
			"GREENBOOT_MAX_BOOT_ATTEMPTS":      "bad",
			"GREENBOOT_WATCHDOG_CHECK_ENABLED": "true",
			"MICROSHIFT_WAIT_TIMEOUT_SEC":      "123",
		}
		val := getIntValueFromMap(m, "GREENBOOT_MAX_BOOT_ATTEMPTS", 3)
		assert.Equal(t, 3, val)
	}
}

func Test_calculateTimeout(t *testing.T) {
	assert.Equal(t, 300, calculateTimeout(300, 3, 2))
	assert.Equal(t, 600, calculateTimeout(300, 3, 1))
	assert.Equal(t, 900, calculateTimeout(300, 3, 0))
	assert.Equal(t, 600, calculateTimeout(300, 4, 2))
}
