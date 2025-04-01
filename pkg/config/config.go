package config

//go:generate ../../scripts/generate-config.sh

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/microshift/pkg/config/apiserver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	netutils "k8s.io/utils/net"
	"k8s.io/utils/ptr"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/openshift/microshift/pkg/util"
	"github.com/vishvananda/netlink"
)

const (
	// default DNS resolve file when systemd-resolved is used
	DefaultSystemdResolvedFile = "/run/systemd/resolve/resolv.conf"
)

var (
	defaultRouterForbiddenCIDRs = []string{
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fe80::/10",
		"fd69::/125",
	}
)

type Config struct {
	DNS       DNS           `json:"dns"`
	Network   Network       `json:"network"`
	Node      Node          `json:"node"`
	ApiServer ApiServer     `json:"apiServer"`
	Etcd      EtcdConfig    `json:"etcd"`
	Debugging Debugging     `json:"debugging"`
	Manifests Manifests     `json:"manifests"`
	Ingress   IngressConfig `json:"ingress"`
	Storage   Storage       `json:"storage"`
	Telemetry Telemetry     `json:"telemetry"`

	// Settings specified in this section are transferred as-is into the Kubelet config.
	// +kubebuilder:validation:Schemaless
	Kubelet map[string]any `json:"kubelet"`

	// Internal-only fields
	userSettings *Config `json:"-"` // the values read from the config file

	MultiNode MultiNodeConfig `json:"-"` // the value read from commond line

	Warnings []string `json:"-"` // Warnings that should not prevent the service from starting.
}

// NewDefault creates a new Config struct populated with the
// default values and with any computed values updated based on those
// defaults.
func NewDefault() *Config {
	c := &Config{}
	if err := c.fillDefaults(); err != nil {
		klog.Fatalf("Failed to initialize config: %v", err)
	}
	if err := c.updateComputedValues(); err != nil {
		klog.Fatalf("Failed to initialize config: %v", err)
	}
	return c
}

// fillDefaults forcibly sets the configuration to the default
// values. We do not use a static struct for the defaults because some
// of them are computed from the environment. If any error occurs
// probing the environment, the values in the Config instance are not
// changed.
func (c *Config) fillDefaults() error {
	// Look up any values that may generate an error
	subjectAltNames, err := getAllHostnames()
	if err != nil {
		return fmt.Errorf("failed to get all hostnames: %v", err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname %v", err)
	}
	nodeIP, err := util.GetHostIP("")
	if err != nil {
		return fmt.Errorf("failed to get host IP: %v", err)
	}

	c.Debugging = Debugging{
		LogLevel: "Normal",
	}
	c.ApiServer = ApiServer{
		SubjectAltNames: subjectAltNames,
		URL:             "https://localhost:6443",
		Port:            6443,
	}
	c.ApiServer.AuditLog = AuditLog{
		MaxFileAge:  0,
		MaxFiles:    10,
		MaxFileSize: 200,
		Profile:     "Default",
	}
	c.Node = Node{
		HostnameOverride: hostname,
		NodeIP:           nodeIP,
	}
	c.DNS = DNS{
		BaseDomain: "example.com",
	}
	c.Network = Network{
		ServiceNodePortRange: "30000-32767",
	}
	c.Etcd = EtcdConfig{
		MemoryLimitMB:           0,
		QuotaBackendBytes:       8 * 1024 * 1024 * 1024,
		MinDefragBytes:          100 * 1024 * 1024,
		MaxFragmentedPercentage: 45,
		DefragCheckFreq:         5 * time.Minute,
	}
	c.Manifests = Manifests{
		KustomizePaths: []string{
			defaultManifestDirLib,
			defaultManifestDirLibGlob,
			defaultManifestDirEtc,
			defaultManifestDirEtcGlob,
		},
	}
	c.Ingress = IngressConfig{
		Status: StatusManaged,
		AdmissionPolicy: RouteAdmissionPolicy{
			NamespaceOwnership: NamespaceOwnershipAllowed,
			WildcardPolicy:     WildcardPolicyDisallowed,
		},
		Ports: IngressPortsConfig{
			Http:  ptr.To[int](80),
			Https: ptr.To[int](443),
		},
		TuningOptions: IngressControllerTuningOptions{
			HeaderBufferBytes:           32768,
			HeaderBufferMaxRewriteBytes: 8192,
			HealthCheckInterval:         &metav1.Duration{Duration: 5 * time.Second},
			ClientTimeout:               &metav1.Duration{Duration: 30 * time.Second},
			ClientFinTimeout:            &metav1.Duration{Duration: 1 * time.Second},
			ServerTimeout:               &metav1.Duration{Duration: 30 * time.Second},
			ServerFinTimeout:            &metav1.Duration{Duration: 1 * time.Second},
			TunnelTimeout:               &metav1.Duration{Duration: 1 * time.Hour},
			TLSInspectDelay:             &metav1.Duration{Duration: 5 * time.Second},
			ThreadCount:                 4,
			MaxConnections:              50000,
		},
		DefaultHttpVersionPolicy: 1,
		LogEmptyRequests:         "Log",
		ForwardedHeaderPolicy:    "Append",
		HTTPEmptyRequestsPolicy:  "Respond",
		ServingCertificateSecret: "router-certs-default",
		TLSSecurityProfile: &configv1.TLSSecurityProfile{
			Type: configv1.TLSProfileIntermediateType,
		},
	}
	c.MultiNode.Enabled = false
	c.Kubelet = nil
	c.Telemetry = telemetryDefaults()

	return nil
}

// incorporateUserSettings merges any values read from the
// configuration file provided by the user with the existing settings
// (usually the defaults).
func (c *Config) incorporateUserSettings(u *Config) {
	c.userSettings = u

	if u.DNS.BaseDomain != "" {
		c.DNS.BaseDomain = u.DNS.BaseDomain
	}

	if u.Network.CNIPlugin != "" {
		c.Network.CNIPlugin = u.Network.CNIPlugin
	}

	if len(u.Network.ClusterNetwork) != 0 {
		c.Network.ClusterNetwork = u.Network.ClusterNetwork
	}
	if len(u.Network.ServiceNetwork) != 0 {
		c.Network.ServiceNetwork = u.Network.ServiceNetwork
		// The default for the API server address is computed from the
		// service network. If the user provides a network without
		// also overriding the computed address, we need to clear the
		// address here so it is recomputed later. If they provide
		// both the network and the address, the address will be
		// copied into place below with the other API server settings.
		if u.ApiServer.AdvertiseAddress == "" {
			c.ApiServer.AdvertiseAddress = ""
		}
	}
	if u.Network.ServiceNodePortRange != "" {
		c.Network.ServiceNodePortRange = u.Network.ServiceNodePortRange
	}
	if u.Network.DNS != "" {
		c.Network.DNS = u.Network.DNS
	}

	if u.Network.Multus.Status != "" {
		c.Network.Multus.Status = u.Network.Multus.Status
	}

	if u.Etcd.MemoryLimitMB != 0 {
		c.Etcd.MemoryLimitMB = u.Etcd.MemoryLimitMB
	}

	if u.Node.HostnameOverride != "" {
		c.Node.HostnameOverride = u.Node.HostnameOverride
	}
	if u.Node.NodeIP != "" {
		c.Node.NodeIP = u.Node.NodeIP
	}
	if u.Node.NodeIPV6 != "" {
		c.Node.NodeIPV6 = u.Node.NodeIPV6
	}
	if len(u.ApiServer.SubjectAltNames) != 0 {
		c.ApiServer.SubjectAltNames = u.ApiServer.SubjectAltNames
	}
	if u.ApiServer.AdvertiseAddress != "" {
		c.ApiServer.AdvertiseAddress = u.ApiServer.AdvertiseAddress
	}
	if u.ApiServer.URL != "" {
		c.ApiServer.URL = u.ApiServer.URL
	}
	if u.ApiServer.AuditLog.Profile != "" {
		c.ApiServer.AuditLog.Profile = u.ApiServer.AuditLog.Profile
	}
	if u.ApiServer.AuditLog.MaxFiles != 0 {
		c.ApiServer.AuditLog.MaxFiles = u.ApiServer.AuditLog.MaxFiles
	}
	if u.ApiServer.AuditLog.MaxFileAge != 0 {
		c.ApiServer.AuditLog.MaxFileAge = u.ApiServer.AuditLog.MaxFileAge
	}
	if u.ApiServer.AuditLog.MaxFileSize != 0 {
		c.ApiServer.AuditLog.MaxFileSize = u.ApiServer.AuditLog.MaxFileSize
	}
	if len(u.ApiServer.TLS.CipherSuites) != 0 {
		c.ApiServer.TLS.CipherSuites = u.ApiServer.TLS.CipherSuites
	}
	if u.ApiServer.TLS.MinVersion != "" {
		c.ApiServer.TLS.MinVersion = u.ApiServer.TLS.MinVersion
	}

	if u.Debugging.LogLevel != "" {
		c.Debugging.LogLevel = u.Debugging.LogLevel
	}

	// Check for nil instead of an empty list because if a user
	// provides a list but it is empty we want to treat that as
	// disabling the manifest loader.
	if u.Manifests.KustomizePaths != nil {
		c.Manifests.KustomizePaths = u.Manifests.KustomizePaths
	}

	if len(u.Ingress.Status) != 0 {
		c.Ingress.Status = u.Ingress.Status
	}

	if len(u.Ingress.AdmissionPolicy.NamespaceOwnership) != 0 {
		c.Ingress.AdmissionPolicy.NamespaceOwnership = u.Ingress.AdmissionPolicy.NamespaceOwnership
	}

	if u.Ingress.Ports.Http != nil {
		c.Ingress.Ports.Http = ptr.To[int](*u.Ingress.Ports.Http)
	}

	if u.Ingress.Ports.Https != nil {
		c.Ingress.Ports.Https = ptr.To[int](*u.Ingress.Ports.Https)
	}

	if len(u.Ingress.ListenAddress) != 0 {
		c.Ingress.ListenAddress = slices.DeleteFunc(u.Ingress.ListenAddress, func(s string) bool {
			return len(s) == 0
		})
	}

	if len(u.ApiServer.NamedCertificates) != 0 {
		c.ApiServer.NamedCertificates = u.ApiServer.NamedCertificates
	}

	if u.Storage.Driver != "" {
		c.Storage.Driver = u.Storage.Driver
	}
	if len(u.Storage.OptionalCSIComponents) > 0 {
		c.Storage.OptionalCSIComponents = u.Storage.OptionalCSIComponents
	}
	if u.Kubelet != nil {
		c.Kubelet = u.Kubelet
	}

	if u.Ingress.TuningOptions.HeaderBufferBytes > 0 {
		c.Ingress.TuningOptions.HeaderBufferBytes = u.Ingress.TuningOptions.HeaderBufferBytes
	}
	if u.Ingress.TuningOptions.HeaderBufferMaxRewriteBytes > 0 {
		c.Ingress.TuningOptions.HeaderBufferMaxRewriteBytes = u.Ingress.TuningOptions.HeaderBufferMaxRewriteBytes
	}
	if u.Ingress.TuningOptions.HealthCheckInterval != nil && u.Ingress.TuningOptions.HealthCheckInterval.Duration >= 1*time.Second {
		c.Ingress.TuningOptions.HealthCheckInterval = u.Ingress.TuningOptions.HealthCheckInterval
	}
	if u.Ingress.TuningOptions.ClientTimeout != nil && u.Ingress.TuningOptions.ClientTimeout.Duration > 0*time.Second {
		c.Ingress.TuningOptions.ClientTimeout = u.Ingress.TuningOptions.ClientTimeout
	}
	if u.Ingress.TuningOptions.ClientFinTimeout != nil && u.Ingress.TuningOptions.ClientFinTimeout.Duration > 0*time.Second {
		c.Ingress.TuningOptions.ClientFinTimeout = u.Ingress.TuningOptions.ClientFinTimeout
	}
	if u.Ingress.TuningOptions.ServerTimeout != nil && u.Ingress.TuningOptions.ServerTimeout.Duration > 0*time.Second {
		c.Ingress.TuningOptions.ServerTimeout = u.Ingress.TuningOptions.ServerTimeout
	}
	if u.Ingress.TuningOptions.ServerFinTimeout != nil && u.Ingress.TuningOptions.ServerFinTimeout.Duration > 0*time.Second {
		c.Ingress.TuningOptions.ServerFinTimeout = u.Ingress.TuningOptions.ServerFinTimeout
	}
	if u.Ingress.TuningOptions.TunnelTimeout != nil && u.Ingress.TuningOptions.TunnelTimeout.Duration > 0*time.Second {
		c.Ingress.TuningOptions.TunnelTimeout = u.Ingress.TuningOptions.TunnelTimeout
	}
	if u.Ingress.TuningOptions.TLSInspectDelay != nil && u.Ingress.TuningOptions.TLSInspectDelay.Duration > 0*time.Second {
		c.Ingress.TuningOptions.TLSInspectDelay = u.Ingress.TuningOptions.TLSInspectDelay
	}
	if u.Ingress.TuningOptions.ThreadCount > 0 {
		c.Ingress.TuningOptions.ThreadCount = u.Ingress.TuningOptions.ThreadCount
	}
	if u.Ingress.TuningOptions.MaxConnections > 0 {
		c.Ingress.TuningOptions.MaxConnections = u.Ingress.TuningOptions.MaxConnections
	}
	if len(u.Ingress.LogEmptyRequests) > 0 {
		c.Ingress.LogEmptyRequests = u.Ingress.LogEmptyRequests
	}
	if len(u.Ingress.ForwardedHeaderPolicy) > 0 {
		c.Ingress.ForwardedHeaderPolicy = u.Ingress.ForwardedHeaderPolicy
	}
	if len(u.Ingress.HTTPEmptyRequestsPolicy) > 0 {
		c.Ingress.HTTPEmptyRequestsPolicy = u.Ingress.HTTPEmptyRequestsPolicy
	}
	if len(u.Ingress.HTTPCompressionPolicy.MimeTypes) > 0 {
		c.Ingress.HTTPCompressionPolicy = u.Ingress.HTTPCompressionPolicy
	}
	if u.Ingress.DefaultHttpVersionPolicy > 0 {
		c.Ingress.DefaultHttpVersionPolicy = u.Ingress.DefaultHttpVersionPolicy
	}

	if u.Telemetry.Status != "" {
		c.Telemetry.Status = u.Telemetry.Status
	}
	if u.Telemetry.Endpoint != "" {
		c.Telemetry.Endpoint = u.Telemetry.Endpoint
	}
	if len(u.Ingress.ServingCertificateSecret) != 0 {
		c.Ingress.ServingCertificateSecret = u.Ingress.ServingCertificateSecret
	}
	if u.Ingress.TLSSecurityProfile != nil {
		c.Ingress.TLSSecurityProfile = u.Ingress.TLSSecurityProfile
	}

	if len(u.Ingress.AdmissionPolicy.WildcardPolicy) != 0 {
		c.Ingress.AdmissionPolicy.WildcardPolicy = u.Ingress.AdmissionPolicy.WildcardPolicy
	}
	if len(u.Ingress.ClientTLS.ClientCertificatePolicy) != 0 {
		c.Ingress.ClientTLS = u.Ingress.ClientTLS
	}
}

// updateComputedValues examins the existing settings and converts any
// inputs to more easily consumable units or fills in any defaults
// computed based on the values of other settings.
func (c *Config) updateComputedValues() error {
	if len(c.Network.ClusterNetwork) == 0 {
		defaultClusterNetwork := "10.42.0.0/16"
		ip := net.ParseIP(c.Node.NodeIP)
		if ip.To4() == nil {
			defaultClusterNetwork = "fd01::/48"
		}
		c.Network.ClusterNetwork = []string{defaultClusterNetwork}
	}
	if len(c.Network.ServiceNetwork) == 0 {
		defaultServiceNetwork := "10.43.0.0/16"
		ip := net.ParseIP(c.Node.NodeIP)
		if ip.To4() == nil {
			defaultServiceNetwork = "fd02::/112"
		}
		c.Network.ServiceNetwork = []string{defaultServiceNetwork}
	}

	if c.IsIPv4() && c.IsIPv6() && len(c.Node.NodeIPV6) == 0 {
		// NodeIPv6 is a dual-stack only parameter that needs to be configured.
		// When the user does not provide a value, MicroShift needs to take
		// one from the host. The function to take the default IP (like node.NodeIP)
		// is not valid in this case, because it relies on net.ChooseHostInterface
		// which gives preference to IPv4 addresses. Instead, a simple helper
		// is used.
		ip, err := util.GetHostIPv6("")
		if err != nil {
			return fmt.Errorf("unable to determine ipv6 host address: %v", err)
		}
		c.Node.NodeIPV6 = ip
	}

	clusterDNS, err := c.computeClusterDNS()
	if err != nil {
		return err
	}
	c.Network.DNS = clusterDNS

	// If KAS advertise address configured, we do not want to apply
	// the IP to the internal interface.
	if c.userSettings != nil && len(c.userSettings.ApiServer.AdvertiseAddress) != 0 {
		c.ApiServer.SkipInterface = true
	}

	// If we have no advertise address, pick one.
	if len(c.ApiServer.AdvertiseAddress) == 0 {
		// Since the KAS advertise address was not provided we will default to the
		// next immediate subnet after the service CIDR. This is due to the fact
		// that using the actual apiserver service IP as an endpoint slice breaks
		// host network pods trying to reach apiserver, as the VIP 10.43.0.1:443 is
		// not translated to 10.43.0.1:6443. It remains unchanged and therefore
		// connects to the ingress router instead, triggering all sorts of errors.
		ip, err := firstIPFromNextSubnet(c.Network.ServiceNetwork[0])
		if err != nil {
			return fmt.Errorf("unable to compute AdvertiseAddress: %s", err)
		}
		c.ApiServer.AdvertiseAddress = ip
	}

	// Use this variable instead, as we may be in dual stack ip an need to
	// configure one extra IP address in the ovn gateway interface. Pick
	// the IP family that was not used for the advertise address and add
	// the first valid IP for the next subnet.
	c.ApiServer.AdvertiseAddresses = []string{c.ApiServer.AdvertiseAddress}
	if c.IsIPv4() && c.IsIPv6() {
		ip, err := firstIPFromNextSubnet(c.Network.ServiceNetwork[1])
		if err != nil {
			return fmt.Errorf("unable to compute secondary address for br-ex: %s", err)
		}
		c.ApiServer.AdvertiseAddresses = append(c.ApiServer.AdvertiseAddresses, ip)
	}

	c.ApiServer.TLS.UpdateValues()

	c.computeLoggingSetting()

	return nil
}

func (c *Config) validate() error {
	if !isValidIPAddress(c.ApiServer.AdvertiseAddress) {
		return fmt.Errorf("error validating apiServer.advertiseAddress (%q)", c.ApiServer.AdvertiseAddress)
	}
	if c.ApiServer.SkipInterface {
		err := checkAdvertiseAddressConfigured(c.ApiServer.AdvertiseAddress)
		if err != nil {
			return err
		}
	}

	if !isValidIPAddress(c.Node.NodeIP) {
		return fmt.Errorf("error validating node.nodeIP (%q)", c.Node.NodeIP)
	}

	if err := validateNetworkStack(c); err != nil {
		return fmt.Errorf("error validating networks: %w", err)
	}

	if !c.Network.validCNIPlugin() {
		return fmt.Errorf("invalid cni plugin for network configuration  %q", c.Network.CNIPlugin)
	}

	//nolint:nestif // extracting the nested ifs will just increase the complexity of the if expressions as validation expands
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
				return fmt.Errorf("cluster URL host %q must be included in subjectAltNames or nodeName", u.String())
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
		) {
			return fmt.Errorf("subjectAltNames must not contain kubernetes service names")
		}
		if stringSliceContains(
			c.ApiServer.SubjectAltNames,
			c.ApiServer.AdvertiseAddresses...,
		) {
			return fmt.Errorf("subjectAltNames must not contain apiserver advertise address IPs")
		}
	}

	if c.Etcd.MemoryLimitMB > 0 && c.Etcd.MemoryLimitMB < EtcdMinimumMemoryLimit {
		return fmt.Errorf("etcd.memoryLimitMB value %d is below the minimum allowed %d",
			c.Etcd.MemoryLimitMB, EtcdMinimumMemoryLimit,
		)
	}

	if c.ApiServer.SkipInterface {
		err := checkAdvertiseAddressConfigured(c.ApiServer.AdvertiseAddresses[0])
		if err != nil {
			return err
		}
	}

	switch c.Ingress.Status {
	case StatusManaged, StatusRemoved:
	default:
		return fmt.Errorf("unsupported ingress.status value %v", c.Ingress.Status)
	}

	switch c.Ingress.AdmissionPolicy.NamespaceOwnership {
	case NamespaceOwnershipAllowed, NamespaceOwnershipStrict:
	default:
		return fmt.Errorf("unsupported namespaceOwnership value %v", c.Ingress.AdmissionPolicy.NamespaceOwnership)
	}

	if c.Ingress.Ports.Http != nil && (*c.Ingress.Ports.Http < 1 || *c.Ingress.Ports.Http > math.MaxUint16) {
		return fmt.Errorf("unsupported value %v for ingress.ports.http", *c.Ingress.Ports.Http)
	}
	if c.Ingress.Ports.Https != nil && (*c.Ingress.Ports.Https < 1 || *c.Ingress.Ports.Https > math.MaxUint16) {
		return fmt.Errorf("unsupported value %v for ingress.ports.https", *c.Ingress.Ports.Https)
	}

	if len(c.Ingress.ListenAddress) != 0 {
		if err := validateRouterListenAddress(c.Ingress.ListenAddress, c.ApiServer.AdvertiseAddresses, c.ApiServer.SkipInterface, c.IsIPv4(), c.IsIPv6()); err != nil {
			return fmt.Errorf("error validating ingress.listenAddress: %w", err)
		}
	}
	if err := validateAuditLogConfig(c.ApiServer.AuditLog); err != nil {
		return fmt.Errorf("error validating apiserver.auditLog:\n%w", err)
	}

	if err := validateNodeIPv6Address(c.Node.NodeIPV6, c.IsIPv4() && c.IsIPv6()); err != nil {
		return fmt.Errorf("error validating node.nodeIPv6: %w", err)
	}

	if errs := c.Storage.IsValid(); c.Storage.IsEnabled() && len(errs) > 0 {
		return fmt.Errorf("error validating storage: %w", errors.Join(errs...))
	}

	if err := c.ApiServer.TLS.Validate(); err != nil {
		return fmt.Errorf("error validating apiServer.tls: %v", err)
	}

	if err := c.Telemetry.validate(); err != nil {
		return fmt.Errorf("error validating telemetry: %v", err)
	}

	return nil
}

// AddWarning saves a warning message to be reported later.
func (c *Config) AddWarning(message string) {
	c.Warnings = append(c.Warnings, message)
}

// UserNodeIP return the user configured NodeIP, or "" if it's unset.
func (c Config) UserNodeIP() string {
	if c.userSettings != nil {
		return c.userSettings.Node.NodeIP
	}
	return ""
}

// UserNodeIPv6 return the user configured NodeIPv6, or "" if it's unset.
func (c Config) UserNodeIPv6() string {
	if c.userSettings != nil {
		return c.userSettings.Node.NodeIPV6
	}
	return ""
}

func (c Config) IsIPv4() bool {
	for _, network := range c.Network.ClusterNetwork {
		ip, _, _ := net.ParseCIDR(network)
		if ip.To4() != nil {
			return true
		}
	}
	return false
}

func (c Config) IsIPv6() bool {
	for _, network := range c.Network.ClusterNetwork {
		ip, _, _ := net.ParseCIDR(network)
		if ip.To4() == nil {
			return true
		}
	}
	return false
}

var allHostnames []string

func getAllHostnames() ([]string, error) {
	if len(allHostnames) != 0 {
		return allHostnames, nil
	}
	cmd := exec.Command("/bin/hostname", "-A")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error when executing 'hostname -A': %v", err)
	}
	outString := out.String()
	outString = strings.Trim(outString[:len(outString)-1], " ")
	// Remove duplicates to avoid having them in the certificates.
	names := strings.Split(outString, " ")
	set := sets.NewString(names...)
	allHostnames = set.List()
	return allHostnames, nil
}

func checkAdvertiseAddressConfigured(advertiseAddress string) error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		// interface addresses come with the mask at the end.
		addrStr := addr.String()
		if idx := strings.Index(addrStr, "/"); idx != -1 {
			addrStr = addrStr[:idx]
		}
		if addrStr == advertiseAddress {
			return nil
		}
	}
	return fmt.Errorf("Advertise address: %s not present in any interface", advertiseAddress)
}

func validateRouterListenAddress(ingressListenAddresses []string, advertiseAddresses []string, skipInterface bool, ipv4, ipv6 bool) error {
	addresses, err := AllowedListeningIPAddresses(ipv4, ipv6)
	if err != nil {
		return err
	}
	nicNames, err := AllowedNICNames()
	if err != nil {
		return err
	}
	for _, entry := range ingressListenAddresses {
		if slices.Contains(advertiseAddresses, entry) && !skipInterface {
			continue
		}
		ip := net.ParseIP(entry)
		if ip == nil {
			if slices.Contains(nicNames, entry) {
				continue
			}
			return fmt.Errorf("interface %v not present in the host", entry)
		}
		if (ip.To4() != nil && !ipv4) || (ip.To4() == nil && !ipv6) {
			return fmt.Errorf("IP %v does not match family of service/cluster network", entry)
		}
		if !slices.Contains(addresses, entry) {
			return fmt.Errorf("IP %v not present in any of the host's interfaces", entry)
		}
	}
	return nil
}

func getForbiddenIPs() ([]*net.IPNet, error) {
	banned := make([]*net.IPNet, 0)
	for _, entry := range defaultRouterForbiddenCIDRs {
		_, netIP, err := net.ParseCIDR(entry)
		if err != nil {
			return nil, err
		}
		banned = append(banned, netIP)
	}
	return banned, nil
}

func getHostAddresses(ipv4, ipv6 bool) ([]net.IP, error) {
	handle, err := netlink.NewHandle()
	if err != nil {
		return nil, err
	}
	links, err := handle.LinkList()
	if err != nil {
		return nil, err
	}
	family := netlink.FAMILY_V4
	if ipv6 {
		family = netlink.FAMILY_ALL
		if !ipv4 {
			family = netlink.FAMILY_V6
		}
	}
	addresses := make([]net.IP, 0, len(links)*2)
	for _, link := range links {
		// Filter out slave NICs. These include ovs/ovn created interfaces, in case of a restart.
		if link.Attrs().MasterIndex != 0 {
			continue
		}
		addressList, err := handle.AddrList(link, family)
		if err != nil {
			return nil, err
		}
		for _, addr := range addressList {
			addresses = append(addresses, addr.IP)
		}
	}
	return addresses, nil
}

func AllowedListeningIPAddresses(ipv4, ipv6 bool) ([]string, error) {
	bannedAddresses, err := getForbiddenIPs()
	if err != nil {
		return nil, err
	}
	hostAddresses, err := getHostAddresses(ipv4, ipv6)
	if err != nil {
		return nil, err
	}
	addressList := make([]string, 0, len(hostAddresses))
	for _, addr := range hostAddresses {
		skip := false
		for _, banned := range bannedAddresses {
			if banned.Contains(addr) {
				skip = true
			}
		}
		if skip {
			continue
		}
		addressList = append(addressList, addr.String())
	}
	return addressList, nil
}

func AllowedNICNames() ([]string, error) {
	handle, err := netlink.NewHandle()
	if err != nil {
		return nil, err
	}
	links, err := handle.LinkList()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(links))
	for _, link := range links {
		names = append(names, link.Attrs().Name)
	}
	return names, nil
}

func validateAuditLogConfig(cfg AuditLog) error {
	// compose a list of errors so that multiple executions are not required to detect each invalid value individually
	errs := make([]error, 0)
	if cfg.Profile != "" {
		if _, err := apiserver.GetPolicy(cfg.Profile); err != nil {
			errs = append(errs, fmt.Errorf("invalid value for apiserver.auditlog.profile: %v", err))
		}
	}
	if cfg.MaxFiles < 0 {
		errs = append(errs, fmt.Errorf("invalid value for apiserver.auditlog.maxFiles, expected value >=0"))
	}
	if cfg.MaxFileAge < 0 {
		errs = append(errs, fmt.Errorf("invalid value for apiserver.auditlog.maxFileAge, expected value >=0"))
	}
	if cfg.MaxFileSize < 0 {
		errs = append(errs, fmt.Errorf("invalid value for apiserver.auditlog.maxFileSize, expected value >=0"))
	}
	return errors.Join(errs...) // Join returns nil if len(errs) == 0
}

func validateNetworkStack(cfg *Config) error {
	if len(cfg.Network.ClusterNetwork) != len(cfg.Network.ServiceNetwork) {
		return fmt.Errorf("network.serviceNetwork and network.clusterNetwork have different cardinality")
	}
	if len(cfg.Network.ServiceNetwork) > 2 {
		return fmt.Errorf("network.serviceNetwork can not have more than 2 entries")
	}
	ipv4Entries := 0
	ipv6Entries := 0
	for i := 0; i < len(cfg.Network.ClusterNetwork); i++ {
		_, _, err := net.ParseCIDR(cfg.Network.ServiceNetwork[i])
		if err != nil {
			return fmt.Errorf("invalid format in network.ServiceNetwork[%d]: %w", i, err)
		}
		_, _, err = net.ParseCIDR(cfg.Network.ClusterNetwork[i])
		if err != nil {
			return fmt.Errorf("invalid format in network.ClusterNetwork[%d]: %w", i, err)
		}
		if netutils.IPFamilyOfCIDRString(cfg.Network.ServiceNetwork[i]) != netutils.IPFamilyOfCIDRString(cfg.Network.ClusterNetwork[i]) {
			return fmt.Errorf("mismatched IP families in network.ServiceNetwork[%d] and network.ClusterNetwork[%d]", i, i)
		}
		if netutils.IPFamilyOfCIDRString(cfg.Network.ServiceNetwork[i]) == netutils.IPv4 {
			ipv4Entries++
		} else {
			ipv6Entries++
		}
	}
	if ipv4Entries > 1 || ipv6Entries > 1 {
		return fmt.Errorf("invalid number of entries of the same IP family in network.serviceNetwork and network.clusterNetwork")
	}
	if netutils.IPFamilyOfString(cfg.ApiServer.AdvertiseAddress) != netutils.IPFamilyOfCIDRString(cfg.Network.ServiceNetwork[0]) {
		return fmt.Errorf("invalid IP family in apiServer.AdvertiseAddress: does not match first network.ServiceNetwork IP family")
	}
	return nil
}

func firstIPFromNextSubnet(subnet string) (string, error) {
	_, svcNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", err
	}
	prefix := 32
	if svcNet.IP.To4() == nil {
		prefix = 128
	}
	nextSubnet, exceed := cidr.NextSubnet(svcNet, prefix)
	if exceed {
		return "", fmt.Errorf("unable to compute next subnet from service CIDR")
	}
	firstValidIP, _ := cidr.AddressRange(nextSubnet)
	return firstValidIP.String(), nil
}

func validateNodeIPv6Address(addr string, dualStack bool) error {
	if len(addr) == 0 {
		if dualStack {
			return fmt.Errorf("dual stack detected, address must be configured")
		}
		return nil
	}
	if !dualStack {
		return fmt.Errorf("dual stack not detected, address must not be configured")
	}
	if !isValidIPAddress(addr) {
		return fmt.Errorf("address %v is not valid", addr)
	}
	if ip := net.ParseIP(addr); ip.To4() != nil {
		return fmt.Errorf("address %v must be ipv6", addr)
	}
	return nil
}
