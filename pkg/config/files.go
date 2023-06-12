package config

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

const (
	ConfigFile = "/etc/microshift/config.yaml"
	DataDir    = "/var/lib/microshift"
	BackupsDir = "/var/lib/microshift-backups"
)

func parse(contents []byte) (*Config, error) {
	c := &Config{}
	if err := yaml.Unmarshal(contents, c); err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %v", err)
	}
	return c, nil
}

func getActiveConfigFromYAML(contents []byte) (*Config, error) {
	userSettings, err := parse(contents)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %q: %v", ConfigFile, err)
	}

	// Start with the defaults, then apply the user settings and
	// recompute dynamic values.
	results := &Config{}
	if err := results.fillDefaults(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}
	results.incorporateUserSettings(userSettings)
	if err := results.updateComputedValues(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}
	if err := results.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}
	return results, nil
}

// ActiveConfig returns the active configuration. If the configuration
// file exists, read it and require it to be valid. Otherwise return
// the default settings.
func ActiveConfig() (*Config, error) {
	_, err := os.Stat(ConfigFile)
	if os.IsNotExist(err) {
		// No configuration file, use the default settings
		return NewDefault(), nil
	} else if err != nil {
		return nil, err
	}

	// Read the file and merge user-provided settings with the defaults
	contents, err := os.ReadFile(ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %q: %v", ConfigFile, err)
	}
	return getActiveConfigFromYAML(contents)
}
