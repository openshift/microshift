package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVerbosity(t *testing.T) {
	var ttests = []struct {
		setting string
		level   int
	}{
		{
			setting: "Normal",
			level:   2,
		},
		{
			setting: "normal",
			level:   2,
		},
		{
			setting: "Debug",
			level:   4,
		},
		{
			setting: "debug",
			level:   4,
		},
		{
			setting: "Trace",
			level:   6,
		},
		{
			setting: "trace",
			level:   6,
		},
		{
			setting: "TraceAll",
			level:   8,
		},
		{
			setting: "traceall",
			level:   8,
		},
		{
			setting: "Unknown",
			level:   2,
		},
		{
			setting: "unknown",
			level:   2,
		},
		{
			setting: "",
			level:   2,
		},
	}

	for _, tt := range ttests {
		t.Run(tt.setting, func(t *testing.T) {
			config := NewDefault()
			config.Debugging.LogLevel = tt.setting
			verbosity := config.GetVerbosity()
			assert.Equal(t, tt.level, verbosity)
		})
	}
}
