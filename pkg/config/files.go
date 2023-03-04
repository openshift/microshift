package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"sigs.k8s.io/yaml"
)

const (
	DefaultUserConfigFile   = "~/.microshift/config.yaml"
	defaultUserDataDir      = "~/.microshift/data"
	DefaultGlobalConfigFile = "/etc/microshift/config.yaml"
	defaultGlobalDataDir    = "/var/lib/microshift"
	// for files managed via management system in /etc, i.e. user applications
	defaultManifestDirEtc = "/etc/microshift/manifests"
	// for files embedded in ostree. i.e. cni/other component customizations
	defaultManifestDirLib = "/usr/lib/microshift/manifests"
)

var (
	configFile   = findConfigFile()
	dataDir      = findDataDir()
	manifestsDir = findManifestsDir()
)

func GetConfigFile() string {
	return configFile
}

func GetDataDir() string {
	return dataDir
}

func GetManifestsDir() []string {
	return manifestsDir
}

// Returns the default user config file if that exists, else the default global
// config file, else the empty string.
func findConfigFile() string {
	userConfigFile, _ := homedir.Expand(DefaultUserConfigFile)
	if _, err := os.Stat(userConfigFile); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Stat(DefaultGlobalConfigFile); errors.Is(err, os.ErrNotExist) {
			return ""
		} else {
			return DefaultGlobalConfigFile
		}
	} else {
		return userConfigFile
	}
}

// Returns the default user data dir if it exists or the user is non-root.
// Returns the default global data dir otherwise.
func findDataDir() string {
	userDataDir, _ := homedir.Expand(defaultUserDataDir)
	if _, err := os.Stat(userDataDir); errors.Is(err, os.ErrNotExist) {
		if os.Geteuid() > 0 {
			return userDataDir
		} else {
			return defaultGlobalDataDir
		}
	} else {
		return userDataDir
	}
}

// Returns the default manifests directories
func findManifestsDir() []string {
	var manifestsDir = []string{defaultManifestDirLib, defaultManifestDirEtc}
	return manifestsDir
}

func parse(contents []byte) (*Config, error) {
	c := &Config{}
	fmt.Printf("parsing %s\n", string(contents))
	if err := yaml.Unmarshal(contents, c); err != nil {
		return nil, fmt.Errorf("Unable to decode configuration: %v", err)
	}
	if err := c.fillDefaults(); err != nil {
		return nil, fmt.Errorf("Invalid configuration: %v", err)
	}
	if err := c.updateComputedValues(); err != nil {
		return nil, fmt.Errorf("Invalid configuration: %v", err)
	}
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("Invalid configuration: %v", err)
	}
	return c, nil
}

func Read(configFile string) (*Config, error) {
	contents, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %v", configFile, err)
	}
	return parse(contents)
}

// Get the active configuration. If the configuration file exists,
// read it and require it to be valid. Otherwise return the default
// settings.
func GetActiveConfig() (*Config, error) {
	var cfg *Config
	filename := GetConfigFile()
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// No configuration file, use the default settings
		return NewMicroshiftConfig(), nil
	} else if err != nil {
		return nil, err
	}
	cfg, err = Read(filename)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
