package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	ctrl "k8s.io/kubernetes/pkg/controlplane"
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

func (c *Config) ReadFromConfigFile(configFile string) error {
	contents, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("reading config file %q: %v", configFile, err)
	}
	var config Config
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return fmt.Errorf("decoding config file %s: %v", configFile, err)
	}

	// Wire new Config type to existing Config
	c.Node = config.Node
	c.Debugging = config.Debugging
	c.Network = config.Network
	if err := c.computeAndUpdateClusterDNS(); err != nil {
		return fmt.Errorf("Failed to validate configuration file %s: %v", configFile, err)
	}

	c.DNS = config.DNS
	c.ApiServer = config.ApiServer
	c.ApiServer.URL = "https://localhost:6443"

	c.Etcd = config.Etcd
	if c.Etcd.DefragCheckFreq != "" {
		d, err := time.ParseDuration(c.Etcd.DefragCheckFreq)
		if err != nil {
			return fmt.Errorf("failed to parse etcd defragCheckFreq: %v", err)
		}
		c.Etcd.DefragCheckDuration = d
	}
	if c.Etcd.MinDefragSize != "" {
		q, err := resource.ParseQuantity(c.Etcd.MinDefragSize)
		if err != nil {
			return fmt.Errorf("failed to parse etcd minDefragSize: %v", err)
		}
		if !q.IsZero() {
			c.Etcd.MinDefragBytes = q.Value()
		}
	}
	if c.Etcd.QuotaBackendSize != "" {
		q, err := resource.ParseQuantity(c.Etcd.QuotaBackendSize)
		if err != nil {
			return fmt.Errorf("failed to parse etcd quotaBackendSize: %v", err)
		}
		if !q.IsZero() {
			c.Etcd.QuotaBackendBytes = q.Value()
		}
	}

	return nil
}

// Note: add a configFile parameter here because of unit test requiring custom
// local directory
func (c *Config) ReadAndValidate(configFile string) error {
	if configFile != "" {
		if err := c.ReadFromConfigFile(configFile); err != nil {
			return err
		}
	}

	if err := c.computeAndUpdateClusterDNS(); err != nil {
		return fmt.Errorf("Failed to validate configuration file %s: %v", configFile, err)
	}

	// If KAS advertise address is not configured then grab it from the service
	// CIDR automatically.
	if len(c.ApiServer.AdvertiseAddress) == 0 {
		// unchecked error because this was done when getting cluster DNS
		_, svcNet, _ := net.ParseCIDR(c.Network.ServiceNetwork[0])
		_, apiServerServiceIP, err := ctrl.ServiceIPRange(*svcNet)
		if err != nil {
			return fmt.Errorf("error getting apiserver IP: %v", err)
		}
		c.ApiServer.AdvertiseAddress = apiServerServiceIP.String()
		c.ApiServer.SkipInterface = false
	} else {
		c.ApiServer.SkipInterface = true
	}

	if len(c.ApiServer.SubjectAltNames) > 0 {
		// Any entry in SubjectAltNames will be included in the external access certificates.
		// Any of the hostnames and IPs (except the node IP) listed below conflicts with
		// other certificates, such as the service network and localhost access.
		// The node IP is a bit special. Apiserver k8s service, which holds a service IP
		// gets resolved to the node IP. If we include the node IP in the SAN then we have
		// an ambiguity, the same IP matches two different certificates and there are errors
		// when trying to reach apiserver from within the cluster using the service IP.
		// Apiserver will decide which certificate to return to client hello based on SNI
		// (which client-go does not use) or raw IP mappings. As soon as there is a match for
		// the node IP it returns that certificate, which is the external access one. This
		// breaks all pods trying to reach apiserver, as hostnames dont match and the certificate
		// is invalid.
		u, err := url.Parse(c.ApiServer.URL)
		if err != nil {
			return fmt.Errorf("failed to parse cluster URL: %v", err)
		}
		if u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" {
			if stringSliceContains(c.ApiServer.SubjectAltNames, "localhost", "127.0.0.1") {
				return fmt.Errorf("subjectAltNames must not contain localhost, 127.0.0.1")
			}
		} else {
			if stringSliceContains(c.ApiServer.SubjectAltNames, c.Node.NodeIP) {
				return fmt.Errorf("subjectAltNames must not contain node IP")
			}
			if !stringSliceContains(c.ApiServer.SubjectAltNames, u.Host) || u.Host != c.Node.HostnameOverride {
				return fmt.Errorf("Cluster URL host %v is not included in subjectAltNames or nodeName", u.String())
			}
		}

		if stringSliceContains(
			c.ApiServer.SubjectAltNames,
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			"openshift",
			"openshift.default",
			"openshift.default.svc",
			"openshift.default.svc.cluster.local",
			c.ApiServer.AdvertiseAddress,
		) {
			return fmt.Errorf("subjectAltNames must not contain apiserver kubernetes service names or IPs")
		}
	}
	// Validate NodeName in config file, node-name should not be changed for an already
	// initialized MicroShift instance. This can lead to Pods being re-scheduled, storage
	// being orphaned or lost, and other side effects.
	if err := c.validateNodeName(c.isDefaultNodeName()); err != nil {
		klog.Fatalf("Error in validating node name: %v", err)
	}

	return nil
}
