package config

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	ctrl "k8s.io/kubernetes/pkg/controlplane"

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
		c.Node.HostnameOverride = nodeName
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

	if c.Etcd.MinDefragSize == "" {
		c.Etcd.MinDefragSize = "100Mi"
	}
	if c.Etcd.MaxFragmentedPercentage == 0 {
		c.Etcd.MaxFragmentedPercentage = 45
	}
	if c.Etcd.DefragCheckFreq == "" {
		c.Etcd.DefragCheckFreq = "5m"
	}
	// DoStartupDefrag:         true,
	if c.Etcd.QuotaBackendSize == "" {
		c.Etcd.QuotaBackendSize = "2Gi"
	}

	return c.updateComputedValues()
}

func (c *Config) updateComputedValues() error {
	clusterDNS, err := c.computeClusterDNS()
	if err != nil {
		return err
	}
	c.Network.DNS = clusterDNS

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

	return nil
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
