package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const (
	defaultUserConfigFile   = "~/.ushift/ushift.yaml"
	defaultUserDataDir      = "~/.ushift"
	defaultGlobalConfigFile = "/etc/ushift/ushift.yaml"
	defaultGlobalDataDir    = "/var/lib/ushift"
)

var (
	defaultRoles = validRoles
	validRoles   = []string{"controlplane", "node"}
)

type MicroshiftConfig struct {
	ConfigFile string
	DataDir    string `yaml:"dataDir"`

	Roles []string `yaml:"roles"`

	// ClusterURL    string `yaml:"url", envconfig:"CLUSTER_URL"`
	// ClusterSecret string `yaml:"secret", envconfig:"CLUSTER_SECRET."`

	// ControlPlaneConfig struct {
	// 	Port string `yaml:"port", envconfig:"SERVER_PORT"`
	// 	Host string `yaml:"host", envconfig:"SERVER_HOST"`
	// } `yaml:"controlplane"`
	// NodeConfig struct {
	// 	Username string `yaml:"user", envconfig:"DB_USERNAME"`
	// 	Password string `yaml:"pass", envconfig:"DB_PASSWORD"`
	// } `yaml:"node"`
}

func NewMicroshiftConfig() *MicroshiftConfig {
	return &MicroshiftConfig{
		ConfigFile: findConfigFile(),
		DataDir:    findDataDir(),
		Roles:      defaultRoles,
	}
}

// Returns the default user config file if that exists, else the default global
// global config file, else the empty string.
func findConfigFile() string {
	userConfigFile, _ := homedir.Expand(defaultUserConfigFile)
	if _, err := os.Stat(userConfigFile); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Stat(defaultGlobalConfigFile); errors.Is(err, os.ErrNotExist) {
			return ""
		} else {
			return defaultGlobalConfigFile
		}
	} else {
		return userConfigFile
	}
}

// Returns the default user data dir if it exists or if neither the default
// user nor global data dir exists yet and the user is non-root.
// Returns the default global data dir otherwise.
func findDataDir() string {
	userDataDir, _ := homedir.Expand(defaultUserDataDir)
	if _, err := os.Stat(userDataDir); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Stat(defaultGlobalDataDir); errors.Is(err, os.ErrNotExist) {
			if os.Geteuid() == 0 {
				return defaultGlobalDataDir
			} else {
				return userDataDir
			}
		} else {
			return defaultGlobalDataDir
		}
	} else {
		return userDataDir
	}
}

func StringInList(s string, list []string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func (c *MicroshiftConfig) ReadFromConfigFile() error {
	if len(c.ConfigFile) == 0 {
		return nil
	}

	f, err := os.Open(c.ConfigFile)
	if err != nil {
		return fmt.Errorf("opening config file %s: %v", c.ConfigFile, err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(c); err != nil {
		return fmt.Errorf("decoding config file %s: %v", c.ConfigFile, err)
	}

	return nil
}

func (c *MicroshiftConfig) ReadFromEnv() error {
	return envconfig.Process("microshift", c)
}

func (c *MicroshiftConfig) ReadFromCmdLine(flags *pflag.FlagSet) error {
	if dataDir, err := flags.GetString("data-dir"); err == nil && flags.Changed("data-dir") {
		c.DataDir = dataDir
	}
	if roles, err := flags.GetStringSlice("roles"); err == nil && flags.Changed("roles") {
		c.Roles = roles
	}
	return nil
}

func (c *MicroshiftConfig) ReadAndValidate(flags *pflag.FlagSet) error {
	if err := c.ReadFromConfigFile(); err != nil {
		return err
	}
	if err := c.ReadFromEnv(); err != nil {
		return err
	}
	if err := c.ReadFromCmdLine(flags); err != nil {
		return err
	}

	for _, role := range c.Roles {
		if !StringInList(role, validRoles) {
			return fmt.Errorf("config error: '%s' is not a valid role, must be in ['controlplane','node']", role)
		}
	}

	return nil
}
