package config

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/openshift/microshift/pkg/util"
)

const (
	// default DNS resolve file when systemd-resolved is used
	DefaultSystemdResolvedFile = "/run/systemd/resolve/resolv.conf"
)

type Config struct {
	DNS       DNS        `json:"dns"`
	Network   Network    `json:"network"`
	Node      Node       `json:"node"`
	ApiServer ApiServer  `json:"apiServer"`
	Etcd      EtcdConfig `json:"etcd"`
	Debugging Debugging  `json:"debugging"`

	// Internal-only fields
	Ingress IngressConfig `json:"-"`
}

func NewMicroshiftConfig() *Config {
	c := &Config{}
	err := c.fillDefaults()
	if err != nil {
		klog.Fatalf("Failed to initialize config: %v", err)
	}
	return c
}

func (c *Config) fillDefaults() error {
	if c.Debugging.LogLevel == "" {
		c.Debugging.LogLevel = "Normal"
	}

	if len(c.ApiServer.SubjectAltNames) == 0 {
		subjectAltNames, err := getAllHostnames()
		if err != nil {
			return fmt.Errorf("failed to get all hostnames: %v", err)
		}
		c.ApiServer.SubjectAltNames = subjectAltNames
	}
	if c.ApiServer.URL == "" {
		c.ApiServer.URL = "https://localhost:6443"
	}

	if c.Node.HostnameOverride == "" {
		nodeName, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("Failed to get hostname %v", err)
		}
		c.Node.HostnameOverride = strings.ToLower(nodeName)
	}
	if c.Node.NodeIP == "" {
		nodeIP, err := util.GetHostIP()
		if err != nil {
			return fmt.Errorf("failed to get host IP: %v", err)
		}
		c.Node.NodeIP = nodeIP
	}

	if c.DNS.BaseDomain == "" {
		c.DNS.BaseDomain = "example.com"
	}

	if len(c.Network.ClusterNetwork) == 0 {
		c.Network.ClusterNetwork = []ClusterNetworkEntry{
			{
				CIDR: "10.42.0.0/16",
			},
		}
	}
	if len(c.Network.ServiceNetwork) == 0 {
		c.Network.ServiceNetwork = []string{
			"10.43.0.0/16",
		}
	}
	if c.Network.ServiceNodePortRange == "" {
		c.Network.ServiceNodePortRange = "30000-32767"
	}
	if c.Network.DNS == "" {
		c.Network.DNS = "10.43.0.10"
	}

	return c.updateComputedValues()
}

func (c *Config) updateComputedValues() error {
	clusterDNS, err := c.computeClusterDNS()
	if err != nil {
		return err
	}
	c.Network.DNS = clusterDNS

	// If KAS advertise address is not configured then compute it from the service
	// CIDR automatically.
	if len(c.ApiServer.AdvertiseAddress) == 0 {
		// unchecked error because this was done when getting cluster DNS
		_, svcNet, _ := net.ParseCIDR(c.Network.ServiceNetwork[0])
		// Since the KAS advertise address was not provided we will default to the
		// next immediate subnet after the service CIDR. This is due to the fact
		// that using the actual apiserver service IP as an endpoint slice breaks
		// host network pods trying to reach apiserver, as the VIP 10.43.0.1:443 is
		// not translated to 10.43.0.1:6443. It remains unchanged and therefore
		// connects to the ingress router instead, triggering all sorts of errors.
		nextSubnet, exceed := cidr.NextSubnet(svcNet, 32)
		if exceed {
			return fmt.Errorf("unable to compute next subnet from service CIDR")
		}
		// First and last are the same because of the /32 netmask.
		firstValidIP, _ := cidr.AddressRange(nextSubnet)
		c.ApiServer.AdvertiseAddress = firstValidIP.String()
		c.ApiServer.SkipInterface = false
	} else {
		c.ApiServer.SkipInterface = true
	}

	return nil
}

func (c *Config) validate() error {
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

	if c.Etcd.MemoryLimitMB > 0 && c.Etcd.MemoryLimitMB < EtcdMinimumMemoryLimit {
		return fmt.Errorf("etcd.memoryLimitMB value %d is below the minimum allowed %d",
			c.Etcd.MemoryLimitMB, EtcdMinimumMemoryLimit,
		)
	}

	return nil
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
		return nil, fmt.Errorf("Error when executing 'hostname -A': %v", err)
	}
	outString := out.String()
	outString = strings.Trim(outString[:len(outString)-1], " ")
	// Remove duplicates to avoid having them in the certificates.
	names := strings.Split(outString, " ")
	set := sets.NewString(names...)
	allHostnames = set.List()
	return allHostnames, nil
}
