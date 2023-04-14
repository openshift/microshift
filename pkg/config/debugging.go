package config

import "strings"

type Debugging struct {
	// Valid values are: "Normal", "Debug", "Trace", "TraceAll".
	// Defaults to "Normal".
	LogLevel string `json:"logLevel"`
}

// GetVerbosity returns the numerical value for LogLevel which is an enum
func (c *Config) GetVerbosity() int {
	var levelNames = map[string]int{
		"normal":   2,
		"debug":    4,
		"trace":    6,
		"traceall": 8,
	}
	verbosity, ok := levelNames[strings.ToLower(c.Debugging.LogLevel)]
	if !ok {
		verbosity = 2
	}
	return verbosity
}
