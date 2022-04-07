package config

import (
	"errors"
	goflag "flag"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/kelseyhightower/envconfig"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/openshift/microshift/pkg/util"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

const (
	defaultUserConfigFile   = "~/.microshift/config.yaml"
	defaultUserDataDir      = "~/.microshift/data"
	defaultGlobalConfigFile = "/etc/microshift/config.yaml"
	defaultGlobalDataDir    = "/var/lib/microshift"
)

var (
	defaultRoles = validRoles
	validRoles   = []string{"controlplane", "node"}
)

type ClusterConfig struct {
	URL string `yaml:"url"`

	ClusterCIDR          string `yaml:"clusterCIDR"`
	ServiceCIDR          string `yaml:"serviceCIDR"`
	ServiceNodePortRange string `yaml:"serviceNodePortRange"`
	DNS                  string `yaml:"dns"`
	Domain               string `yaml:"domain"`
}

type ControlPlaneConfig struct {
	// Token string `yaml:"token", envconfig:"CONTROLPLANE_TOKEN"`
}

type NodeConfig struct {
	// Token string `yaml:"token", envconfig:"NODE_TOKEN"`
}

type MicroshiftConfig struct {
	ConfigFile string
	DataDir    string `yaml:"dataDir"`

	AuditLogDir string `yaml:"auditLogDir"`
	LogVLevel   int    `yaml:"logVLevel"`

	Roles []string `yaml:"roles"`

	NodeName string `yaml:"nodeName"`
	NodeIP   string `yaml:"nodeIP"`

	Cluster      ClusterConfig      `yaml:"cluster"`
	ControlPlane ControlPlaneConfig `yaml:"controlPlane"`
	Node         NodeConfig         `yaml:"node"`
}

func NewMicroshiftConfig() *MicroshiftConfig {
	nodeName, err := os.Hostname()
	if err != nil {
		klog.Fatalf("Failed to get hostname %v", err)
	}
	nodeIP, err := util.GetHostIP()
	if err != nil {
		klog.Warningf("failed to get host IP: %v, using: %q", err, nodeIP)
	}

	return &MicroshiftConfig{
		ConfigFile:  findConfigFile(),
		DataDir:     findDataDir(),
		AuditLogDir: "",
		LogVLevel:   0,
		Roles:       defaultRoles,
		NodeName:    nodeName,
		NodeIP:      nodeIP,
		Cluster: ClusterConfig{
			URL:                  "https://127.0.0.1:6443",
			ClusterCIDR:          "10.42.0.0/16",
			ServiceCIDR:          "10.43.0.0/16",
			ServiceNodePortRange: "30000-32767",
			DNS:                  "10.43.0.10",
			Domain:               "cluster.local",
		},
		ControlPlane: ControlPlaneConfig{},
		Node:         NodeConfig{},
	}
}

// extract the api server port from the cluster URL
func (c *ClusterConfig) ApiServerPort() (string, error) {
	parsed, err := url.Parse(c.URL)
	if err != nil {
		return "", err
	}

	// default empty URL to port 6443
	if parsed.Port() == "" {
		return "6443", nil
	}
	return parsed.Port(), nil
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
	if auditLogDir, err := flags.GetString("audit-log-dir"); err == nil && flags.Changed("audit-log-dir") {
		c.AuditLogDir = auditLogDir
	}
	if vLevelFlag := flags.Lookup("v"); vLevelFlag != nil && flags.Changed("v") {
		c.LogVLevel, _ = strconv.Atoi(vLevelFlag.Value.String())
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

func InitGlobalFlags() {
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	goflag.CommandLine.VisitAll(func(goflag *goflag.Flag) {
		if StringInList(goflag.Name, []string{"v", "log_file"}) {
			pflag.CommandLine.AddGoFlag(goflag)
		}
	})

	pflag.CommandLine.MarkHidden("log-flush-frequency")
	pflag.CommandLine.MarkHidden("log_file")
	pflag.CommandLine.MarkHidden("version")
}
