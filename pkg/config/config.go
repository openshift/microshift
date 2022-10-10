package config

import (
	"errors"
	goflag "flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	"github.com/openshift/microshift/pkg/util"
)

const (
	defaultUserConfigFile   = "~/.microshift/config.yaml"
	defaultUserDataDir      = "~/.microshift/data"
	defaultGlobalConfigFile = "/etc/microshift/config.yaml"
	DefaultGlobalDataDir    = "/var/lib/microshift"
	// for files managed via management system in /etc, i.e. user applications
	defaultManifestDirEtc = "/etc/microshift/manifests"
	// for files embedded in ostree. i.e. cni/other component customizations
	defaultManifestDirLib = "/usr/lib/microshift/manifests"
)

var (
	validRoles = []string{"controlplane", "node"}
)

type ClusterConfig struct {
	// URL string `json:"url"`

	// ClusterCIDR          string `json:"clusterCIDR"`
	// ServiceCIDR          string `json:"serviceCIDR"`
	// ServiceNodePortRange string `json:"serviceNodePortRange"`
	// DNS string `json:"dns"`
	// Domain               string `json:"domain"`
	// MTU string `json:"mtu"`
}

type DebugConfig struct {
	// Pprof bool `json:"pprof"`
}

type MicroshiftConfig struct {
	// ConfigFile string `json:"configFile"`
	// DataDir    string `json:"dataDir"`
	//
	// AuditLogDir string `json:"auditLogDir"`
	// LogVLevel   int    `json:"logVLevel"`
	//
	// Roles []string `json:"roles"`
	//
	// NodeName string `json:"nodeName"`
	// NodeIP   string `json:"nodeIP"`
	//
	// Cluster ClusterConfig `json:"cluster"`
	//
	// Manifests []string `json:"manifests"`
	//
	// Debug DebugConfig `json:"debug"`
}

// KubeConfigID identifies the different kubeconfigs managed in the DataDir
type KubeConfigID string

const (
	KubeAdmin             KubeConfigID = "kubeadmin"
	KubeControllerManager KubeConfigID = "kube-controller-manager"
	KubeScheduler         KubeConfigID = "kube-scheduler"
	Kubelet               KubeConfigID = "kubelet"
)

// KubeConfigPath returns the path to the specified kubeconfig file.
func (cfg *MicroshiftConfig) KubeConfigPath(id KubeConfigID) string {
	return filepath.Join(cfg.DataDir, "resources", string(id), "kubeconfig")
}

func NewMicroshiftConfig() *MicroshiftConfig {
	nodeName, err := os.Hostname()
	if err != nil {
		klog.Fatalf("Failed to get hostname %v", err)
	}
	nodeIP, err := util.GetHostIP()
	if err != nil {
		klog.Fatalf("failed to get host IP: %v", err)
	}

	dataDir := findDataDir()

	defaultRoles := make([]string, len(validRoles))
	copy(defaultRoles, validRoles)
	return &MicroshiftConfig{
		ConfigFile:  findConfigFile(),
		DataDir:     dataDir,
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
			MTU:                  "1400",
		},
		Manifests: []string{defaultManifestDirLib, defaultManifestDirEtc},
	}
}

// extract the api server port from the cluster URL
func (c *ClusterConfig) ApiServerPort() (int, error) {
	var port string

	parsed, err := url.Parse(c.URL)
	if err != nil {
		return 0, err
	}

	// default empty URL to port 6443
	port = parsed.Port()
	if port == "" {
		port = "6443"
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return 0, err
	}
	return portNum, nil
}

// Returns the default user config file if that exists, else the default global
// config file, else the empty string.
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
			return DefaultGlobalDataDir
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

	contents, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		return fmt.Errorf("reading config file %s: %v", c.ConfigFile, err)
	}

	if err := yaml.Unmarshal(contents, c); err != nil {
		return fmt.Errorf("decoding config file %s: %v", c.ConfigFile, err)
	}

	c.updateManifestList()

	return nil
}

func (c *MicroshiftConfig) ReadFromEnv() error {
	if err := envconfig.Process("microshift", c); err != nil {
		return err
	}
	c.updateManifestList()
	return nil
}

func (c *MicroshiftConfig) updateManifestList() {
	defaultCfg := NewMicroshiftConfig()
	if c.DataDir != defaultCfg.DataDir && reflect.DeepEqual(defaultCfg.Manifests, c.Manifests) {
		c.Manifests = []string{defaultManifestDirLib, defaultManifestDirEtc, filepath.Join(c.DataDir, "manifests")}
	}
}

func (c *MicroshiftConfig) ReadFromCmdLine(flags *pflag.FlagSet) error {
	if s, err := flags.GetString("data-dir"); err == nil && flags.Changed("data-dir") {
		c.DataDir = s
		// if the defaults are present, rebuild based on the new data-dir
		c.updateManifestList()
	}
	if s, err := flags.GetString("audit-log-dir"); err == nil && flags.Changed("audit-log-dir") {
		c.AuditLogDir = s
	}
	if f := flags.Lookup("v"); f != nil && flags.Changed("v") {
		c.LogVLevel, _ = strconv.Atoi(f.Value.String())
	}
	if ss, err := flags.GetStringSlice("roles"); err == nil && flags.Changed("roles") {
		c.Roles = ss
	}
	if s, err := flags.GetString("node-name"); err == nil && flags.Changed("node-name") {
		c.NodeName = s
	}
	if s, err := flags.GetString("node-ip"); err == nil && flags.Changed("node-ip") {
		c.NodeIP = s
	}
	if s, err := flags.GetString("url"); err == nil && flags.Changed("url") {
		c.Cluster.URL = s
	}
	if s, err := flags.GetString("cluster-cidr"); err == nil && flags.Changed("cluster-cidr") {
		c.Cluster.ClusterCIDR = s
	}
	if s, err := flags.GetString("service-cidr"); err == nil && flags.Changed("service-cidr") {
		c.Cluster.ServiceCIDR = s
	}
	if s, err := flags.GetString("service-node-port-range"); err == nil && flags.Changed("service-node-port-range") {
		c.Cluster.ServiceNodePortRange = s
	}
	if s, err := flags.GetString("cluster-dns"); err == nil && flags.Changed("cluster-dns") {
		c.Cluster.DNS = s
	}
	if s, err := flags.GetString("cluster-domain"); err == nil && flags.Changed("cluster-domain") {
		c.Cluster.Domain = s
	}
	if s, err := flags.GetString("cluster-mtu"); err == nil && flags.Changed("cluster-mtu") {
		c.Cluster.MTU = s
	}
	if b, err := flags.GetBool("debug.pprof"); err == nil && flags.Changed("debug.pprof") {
		c.Debug.Pprof = b
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
			return fmt.Errorf("config error: '%s' is not a valid role, must be in {%s}", role, strings.Join(validRoles, ", "))
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
