package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	"github.com/openshift/microshift/pkg/util"
)

const (
	defaultUserConfigFile   = "~/.microshift/config.yaml"
	defaultUserDataDir      = "~/.microshift/data"
	defaultGlobalConfigFile = "/etc/microshift/config.yaml"
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

type ClusterConfig struct {
	URL string `json:"url"`

	ClusterCIDR          string `json:"clusterCIDR"`
	ServiceCIDR          string `json:"serviceCIDR"`
	ServiceNodePortRange string `json:"serviceNodePortRange"`
	DNS                  string `json:"dns"`
	Domain               string `json:"domain"`
}

type IngressConfig struct {
	ServingCertificate []byte
	ServingKey         []byte
}

type MicroshiftConfig struct {
	LogVLevel int `json:"logVLevel"`

	NodeName string `json:"nodeName"`
	NodeIP   string `json:"-"`

	Cluster ClusterConfig `json:"cluster"`

	Ingress IngressConfig `json:"-"`
}

func GetConfigFile() string {
	return configFile
}

func GetDataDir() string {
	return dataDir
}

func GetManifestsDir() []string {
	return manifestsDir
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
	return filepath.Join(dataDir, "resources", string(id), "kubeconfig")
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
	// Use the node IP to set default URL
	// If user wants to configure node IP manually, they can set it in URL via config
	// file, env var and cli
	url := "https://" + nodeIP + ":6443"
	return &MicroshiftConfig{
		LogVLevel: 0,
		NodeName:  nodeName,
		Cluster: ClusterConfig{
			URL:                  url,
			ClusterCIDR:          "10.42.0.0/16",
			ServiceCIDR:          "10.43.0.0/16",
			ServiceNodePortRange: "30000-32767",
			DNS:                  "10.43.0.10",
			Domain:               "cluster.local",
		},
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

// extract the api server port from the cluster URL
func (c *ClusterConfig) ApiServerHostName() (string, error) {
	parsed, err := url.Parse(c.URL)
	if err != nil {
		return "", err
	}

	return parsed.Hostname(), nil
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

func StringInList(s string, list []string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func (c *MicroshiftConfig) ReadFromConfigFile(configFile string) error {
	contents, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("reading config file %s: %v", configFile, err)
	}

	if err := yaml.Unmarshal(contents, c); err != nil {
		return fmt.Errorf("decoding config file %s: %v", configFile, err)
	}

	return nil
}

func (c *MicroshiftConfig) ReadFromEnv() error {
	if err := envconfig.Process("microshift", c); err != nil {
		return err
	}
	return nil
}

func (c *MicroshiftConfig) ReadFromCmdLine(flags *pflag.FlagSet) error {
	if f := flags.Lookup("v"); f != nil && flags.Changed("v") {
		c.LogVLevel, _ = strconv.Atoi(f.Value.String())
	}
	if s, err := flags.GetString("node-name"); err == nil && flags.Changed("node-name") {
		c.NodeName = s
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

	return nil
}

// Note: add a configFile parameter here because of unit test requiring custom
// local directory
func (c *MicroshiftConfig) ReadAndValidate(configFile string, flags *pflag.FlagSet) error {
	if configFile != "" {
		if err := c.ReadFromConfigFile(configFile); err != nil {
			return err
		}
	}
	if err := c.ReadFromEnv(); err != nil {
		return err
	}
	if err := c.ReadFromCmdLine(flags); err != nil {
		return err
	}

	return nil
}

func HideUnsupportedFlags(flags *pflag.FlagSet) {
	// hide logging flags that we do not use/support
	loggingFlags := pflag.NewFlagSet("logging-flags", pflag.ContinueOnError)
	logs.AddFlags(loggingFlags)

	supportedLoggingFlags := sets.NewString("v")

	loggingFlags.VisitAll(func(pf *pflag.Flag) {
		if !supportedLoggingFlags.Has(pf.Name) {
			flags.MarkHidden(pf.Name)
		}
	})

	flags.MarkHidden("version")
}
