package config

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
	ctrl "k8s.io/kubernetes/pkg/controlplane"
	"sigs.k8s.io/yaml"

	"github.com/openshift/microshift/pkg/util"
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
	// default DNS resolve file when systemd-resolved is used
	DefaultSystemdResolvedFile = "/run/systemd/resolve/resolv.conf"
)

var (
	configFile   = findConfigFile()
	dataDir      = findDataDir()
	manifestsDir = findManifestsDir()
)

type ClusterConfig struct {
	URL                  string `json:"-"`
	ClusterCIDR          string `json:"clusterCIDR"`
	ServiceCIDR          string `json:"serviceCIDR"`
	ServiceNodePortRange string `json:"serviceNodePortRange"`
	DNS                  string `json:"-"`
}

type IngressConfig struct {
	ServingCertificate []byte
	ServingKey         []byte
}

type InternalEtcdConfig struct {
	// The limit on the size of the etcd database; etcd will start failing writes if its size on disk reaches this value
	QuotaBackendBytes int64
	// If the backend is fragmented more than `maxFragmentedPercentage`
	//		and the database size is greater than `minDefragBytes`, do a defrag.
	MinDefragBytes          int64
	MaxFragmentedPercentage float64
	// How often to check the conditions for defragging (0 means no defrags, except for a single on startup if `doStartupDefrag` is set).
	DefragCheckFreq time.Duration
	// Whether or not to do a defrag when the server finishes starting
	DoStartupDefrag bool
}

type EtcdConfig struct {
	// The limit on the size of the etcd database; etcd will start failing writes if its size on disk reaches this value
	QuotaBackendSize string
	// If the backend is fragmented more than `maxFragmentedPercentage`
	//		and the database size is greater than `minDefragSize`, do a defrag.
	MinDefragSize           string
	MaxFragmentedPercentage float64
	// How often to check the conditions for defragging (0 means no defrags, except for a single on startup if `doStartupDefrag` is set).
	DefragCheckFreq string
	// Whether or not to do a defrag when the server finishes starting
	DoStartupDefrag bool
}

type MicroshiftConfig struct {
	SubjectAltNames []string `json:"subjectAltNames"`
	// Kube apiserver advertise address to work around the certificates issue
	// when requiring external access using the node IP. This will turn into
	// the IP configured in the endpoint slice for kubernetes service. Must be
	// a reachable IP from pods. Defaults to service network CIDR first
	// address.
	KASAdvertiseAddress string `json:"kasAdvertiseAddress"`
	// Determines if kube-apiserver controller should configure the
	// KASAdvertiseAddress in the loopback interface. Automatically computed.
	SkipKASInterface bool          `json:"-"`
	Cluster          ClusterConfig `json:"cluster"`

	Ingress IngressConfig      `json:"-"`
	Etcd    InternalEtcdConfig `json:"etcd"`

	DNS       DNS       `json:"-"`
	Node      Node      `json:"-"`
	Debugging Debugging `json:"debugging"`
}

// Top level config file
type Config struct {
	DNS       DNS        `json:"dns"`
	Network   Network    `json:"network"`
	Node      Node       `json:"node"`
	ApiServer ApiServer  `json:"apiServer"`
	Debugging Debugging  `json:"debugging"`
	Etcd      EtcdConfig `json:"etcd"`
}

type Network struct {
	// IP address pool to use for pod IPs.
	// This field is immutable after installation.
	ClusterNetwork []ClusterNetworkEntry `json:"clusterNetwork,omitempty"`

	// IP address pool for services.
	// Currently, we only support a single entry here.
	// This field is immutable after installation.
	ServiceNetwork []string `json:"serviceNetwork,omitempty"`

	// The port range allowed for Services of type NodePort.
	// If not specified, the default of 30000-32767 will be used.
	// Such Services without a NodePort specified will have one
	// automatically allocated from this range.
	// This parameter can be updated after the cluster is
	// installed.
	// +kubebuilder:validation:Pattern=`^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])-([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`
	ServiceNodePortRange string `json:"serviceNodePortRange,omitempty"`
}

type ClusterNetworkEntry struct {
	// The complete block for pod IPs.
	CIDR string `json:"cidr,omitempty"`
}

type DNS struct {
	// baseDomain is the base domain of the cluster. All managed DNS records will
	// be sub-domains of this base.
	//
	// For example, given the base domain `example.com`, router exposed
	// domains will be formed as `*.apps.example.com` by default,
	// and API service will have a DNS entry for `api.example.com`,
	// as well as "api-int.example.com" for internal k8s API access.
	//
	// Once set, this field cannot be changed.
	BaseDomain string `json:"baseDomain"`
}

type ApiServer struct {
	// SubjectAltNames added to API server certs
	SubjectAltNames []string `json:"subjectAltNames"`
	// AdvertiseAddress for endpoint slices in kubernetes service. Developer
	// only parameter, wont show in show-config commands or docs.
	AdvertiseAddress string `json:"advertiseAddress,omitempty"`
}

type Node struct {
	// If non-empty, will use this string to identify the node instead of the hostname
	HostnameOverride string `json:"hostnameOverride"`

	// IP address of the node, passed to the kubelet.
	// If not specified, kubelet will use the node's default IP address.
	NodeIP string `json:"nodeIP"`
}

type Debugging struct {
	// Valid values are: "Normal", "Debug", "Trace", "TraceAll".
	// Defaults to "Normal".
	LogLevel string `json:"logLevel"`
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
	KubeAdmin               KubeConfigID = "kubeadmin"
	KubeControllerManager   KubeConfigID = "kube-controller-manager"
	KubeScheduler           KubeConfigID = "kube-scheduler"
	Kubelet                 KubeConfigID = "kubelet"
	ClusterPolicyController KubeConfigID = "cluster-policy-controller"
	RouteControllerManager  KubeConfigID = "route-controller-manager"
)

// KubeConfigPath returns the path to the specified kubeconfig file.
func (cfg *MicroshiftConfig) KubeConfigPath(id KubeConfigID) string {
	return filepath.Join(dataDir, "resources", string(id), "kubeconfig")
}

func (cfg *MicroshiftConfig) KubeConfigAdminPath(id string) string {
	return filepath.Join(dataDir, "resources", string(KubeAdmin), id, "kubeconfig")
}

func getAllHostnames() ([]string, error) {
	cmd := exec.Command("/bin/hostname", "-A")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Error when executing 'hostname -A': %v", err)
	}
	outString := out.String()
	outString = strings.Trim(outString[:len(outString)-1], " ")
	// Remove duplicates to avoid having them in the certificates.
	names := strings.Split(outString, " ")
	set := sets.NewString(names...)
	return set.List(), nil
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
	subjectAltNames, err := getAllHostnames()
	if err != nil {
		klog.Fatalf("failed to get all hostnames: %v", err)
	}

	return &MicroshiftConfig{
		Debugging: Debugging{
			LogLevel: "Normal",
		},
		SubjectAltNames: subjectAltNames,
		Node: Node{
			HostnameOverride: nodeName,
			NodeIP:           nodeIP,
		},
		DNS: DNS{
			BaseDomain: "example.com",
		},
		Cluster: ClusterConfig{
			URL:                  "https://localhost:6443",
			ClusterCIDR:          "10.42.0.0/16",
			ServiceCIDR:          "10.43.0.0/16",
			ServiceNodePortRange: "30000-32767",
		},
		Etcd: InternalEtcdConfig{
			MinDefragBytes:          100 * 1024 * 1024, // 100MB
			MaxFragmentedPercentage: 45,                // percent
			DefragCheckFreq:         5 * time.Minute,
			DoStartupDefrag:         true,
			QuotaBackendBytes:       2 * 1024 * 1024 * 1024, // 2GB
		},
	}
}

// Determine if the config file specified a NodeName (by default it's assigned the hostname)
func (c *MicroshiftConfig) isDefaultNodeName() bool {
	hostname, err := os.Hostname()
	if err != nil {
		klog.Fatalf("Failed to get hostname %v", err)
	}
	return c.Node.HostnameOverride == hostname
}

// Read or set the NodeName that will be used for this MicroShift instance
func (c *MicroshiftConfig) establishNodeName() (string, error) {
	filePath := filepath.Join(GetDataDir(), ".nodename")
	contents, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		// ensure that dataDir exists
		os.MkdirAll(GetDataDir(), 0700)
		if err := os.WriteFile(filePath, []byte(c.Node.HostnameOverride), 0444); err != nil {
			return "", fmt.Errorf("failed to write nodename file %q: %v", filePath, err)
		}
		return c.Node.HostnameOverride, nil
	} else if err != nil {
		return "", err
	}
	return string(contents), nil
}

// Validate the NodeName to be used for this MicroShift instances
func (c *MicroshiftConfig) validateNodeName(isDefaultNodeName bool) error {
	if addr := net.ParseIP(c.Node.HostnameOverride); addr != nil {
		return fmt.Errorf("NodeName can not be an IP address: %q", c.Node.HostnameOverride)
	}

	establishedNodeName, err := c.establishNodeName()
	if err != nil {
		return fmt.Errorf("failed to establish NodeName: %v", err)
	}

	if establishedNodeName != c.Node.HostnameOverride {
		if !isDefaultNodeName {
			return fmt.Errorf("configured NodeName %q does not match previous NodeName %q , NodeName cannot be changed for a device once established",
				c.Node.HostnameOverride, establishedNodeName)
		} else {
			c.Node.HostnameOverride = establishedNodeName
			klog.Warningf("NodeName has changed due to a host name change, using previously established NodeName %q."+
				"Please consider using a static NodeName in configuration", c.Node.HostnameOverride)
		}
	}

	return nil
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
		return fmt.Errorf("reading config file %q: %v", configFile, err)
	}
	var config Config
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return fmt.Errorf("decoding config file %s: %v", configFile, err)
	}

	// Wire new Config type to existing MicroshiftConfig
	c.Node = config.Node
	c.Debugging = config.Debugging
	if len(config.Network.ClusterNetwork) != 0 {
		c.Cluster.ClusterCIDR = config.Network.ClusterNetwork[0].CIDR
	}
	if len(config.Network.ServiceNetwork) != 0 {
		c.Cluster.ServiceCIDR = config.Network.ServiceNetwork[0]
	}
	if config.Network.ServiceNodePortRange != "" {
		c.Cluster.ServiceNodePortRange = config.Network.ServiceNodePortRange
	}
	c.DNS = config.DNS
	if len(config.ApiServer.SubjectAltNames) > 0 {
		c.SubjectAltNames = config.ApiServer.SubjectAltNames
	}
	if len(config.ApiServer.AdvertiseAddress) > 0 {
		c.KASAdvertiseAddress = config.ApiServer.AdvertiseAddress
	}

	if config.Etcd.DefragCheckFreq != "" {
		d, err := time.ParseDuration(config.Etcd.DefragCheckFreq)
		if err != nil {
			return fmt.Errorf("failed to parse etcd defragCheckFreq: %v", err)
		}
		c.Etcd.DefragCheckFreq = d
	}
	if config.Etcd.MinDefragSize != "" {
		q, err := resource.ParseQuantity(config.Etcd.MinDefragSize)
		if err != nil {
			return fmt.Errorf("failed to parse etcd minDefragSize: %v", err)
		}
		if !q.IsZero() {
			c.Etcd.MinDefragBytes = q.Value()
		}
	}
	if config.Etcd.MaxFragmentedPercentage > 0 {
		c.Etcd.MaxFragmentedPercentage = config.Etcd.MaxFragmentedPercentage
	}
	if config.Etcd.QuotaBackendSize != "" {
		q, err := resource.ParseQuantity(config.Etcd.QuotaBackendSize)
		if err != nil {
			return fmt.Errorf("failed to parse etcd quotaBackendSize: %v", err)
		}
		if !q.IsZero() {
			c.Etcd.QuotaBackendBytes = q.Value()
		}
	}
	c.Etcd.DoStartupDefrag = config.Etcd.DoStartupDefrag

	return nil
}

// Note: add a configFile parameter here because of unit test requiring custom
// local directory
func (c *MicroshiftConfig) ReadAndValidate(configFile string) error {
	if configFile != "" {
		if err := c.ReadFromConfigFile(configFile); err != nil {
			return err
		}
	}

	// validate serviceCIDR
	clusterDNS, err := getClusterDNS(c.Cluster.ServiceCIDR)
	if err != nil {
		return fmt.Errorf("failed to get DNS IP: %v", err)
	}
	c.Cluster.DNS = clusterDNS

	// If KAS advertise address is not configured then grab it from the service
	// CIDR automatically.
	if len(c.KASAdvertiseAddress) == 0 {
		// unchecked error because this was done when getting cluster DNS
		_, svcNet, _ := net.ParseCIDR(c.Cluster.ServiceCIDR)
		_, apiServerServiceIP, err := ctrl.ServiceIPRange(*svcNet)
		if err != nil {
			return fmt.Errorf("error getting apiserver IP: %v", err)
		}
		c.KASAdvertiseAddress = apiServerServiceIP.String()
		c.SkipKASInterface = false
	} else {
		c.SkipKASInterface = true
	}

	if len(c.SubjectAltNames) > 0 {
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
		u, err := url.Parse(c.Cluster.URL)
		if err != nil {
			return fmt.Errorf("failed to parse cluster URL: %v", err)
		}
		if u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" {
			if stringSliceContains(c.SubjectAltNames, "localhost", "127.0.0.1") {
				return fmt.Errorf("subjectAltNames must not contain localhost, 127.0.0.1")
			}
		} else {
			if stringSliceContains(c.SubjectAltNames, c.Node.NodeIP) {
				return fmt.Errorf("subjectAltNames must not contain node IP")
			}
			if !stringSliceContains(c.SubjectAltNames, u.Host) || u.Host != c.Node.HostnameOverride {
				return fmt.Errorf("Cluster URL host %v is not included in subjectAltNames or nodeName", u.String())
			}
		}

		if stringSliceContains(
			c.SubjectAltNames,
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			"openshift",
			"openshift.default",
			"openshift.default.svc",
			"openshift.default.svc.cluster.local",
			c.KASAdvertiseAddress,
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

// getClusterDNS returns cluster DNS IP that is 10th IP of the ServiceNetwork
func getClusterDNS(serviceCIDR string) (string, error) {
	_, service, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		return "", fmt.Errorf("invalid service cidr %v: %v", serviceCIDR, err)
	}
	dnsClusterIP, err := cidr.Host(service, 10)
	if err != nil {
		return "", fmt.Errorf("service cidr must have at least 10 distinct host addresses %v: %v", serviceCIDR, err)
	}

	return dnsClusterIP.String(), nil
}

func stringSliceContains(list []string, elements ...string) bool {
	for _, value := range list {
		for _, element := range elements {
			if value == element {
				return true
			}
		}
	}
	return false
}

// GetVerbosity returns the numerical value for LogLevel which is an enum
func (c *MicroshiftConfig) GetVerbosity() int {
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
