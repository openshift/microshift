package config

type Debugging struct {
	// Valid values are: "Normal", "Debug", "Trace", "TraceAll".
	// Defaults to "Normal".
	LogLevel string `json:"logLevel"`
}

// GetVerbosity returns the numerical value for LogLevel which is an enum
func (c *Config) GetVerbosity() int {
	var verbosity int
	switch c.Debugging.LogLevel {
	case "Normal":
		verbosity = 2
	case "Debug":
		verbosity = 4
	case "Trace":
		verbosity = 6
	case "TraceAll":
		verbosity = 8
	default:
		verbosity = 2
	}
	return verbosity
}
