package ovn

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"

	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	ovnConfigFileName           = "ovn.yaml"
	OVNGatewayInterface         = "br-ex"
	defaultMTU                  = 1500
	OVNKubernetesV4MasqueradeIP = "169.254.169.2"
	OVNKubernetesV6MasqueradeIP = "fd69::2"
)

type OVNKubernetesConfig struct {
	// Configuration for microshift-ovs-init.service
	OVSInit OVSInitConfig `json:"ovsInit,omitempty"`
	// MTU to use for the pod interface. Default is 1500.
	MTU int `json:"mtu,omitempty"`
}

type OVSInitConfig struct {
	// disable microshift-ovs-init.service.
	// OVS bridge "br-ex" needs to be configured manually when disableOVSInit is true.
	DisableOVSInit bool `json:"disableOVSInit,omitempty"`
	// Uplink interface for OVS bridge "br-ex"
	GatewayInterface string `json:"gatewayInterface,omitempty"`
}

func (o *OVNKubernetesConfig) Validate() error {
	// br-ex is required to run ovn-kubernetes
	err := o.validateOVSBridge()
	if err != nil {
		return err
	}
	err = o.validateConfig()
	if err != nil {
		return err
	}
	return nil
}

// validateOVSBridge validates the existence of ovn-kubernetes br-ex bridge
func (o *OVNKubernetesConfig) validateOVSBridge() error {
	_, err := net.InterfaceByName(OVNGatewayInterface)
	return err
}

// validateConfig validates the user defined configuration in /etc/microshift/ovn.yaml
func (o *OVNKubernetesConfig) validateConfig() error {
	// validate gateway interfaces conf
	if o.OVSInit.GatewayInterface != "" {
		_, err := net.InterfaceByName(o.OVSInit.GatewayInterface)
		if err != nil {
			return fmt.Errorf("gateway interface %s not found", o.OVSInit.GatewayInterface)
		}
	}

	// validate MTU conf
	iface, err := net.InterfaceByName(OVNGatewayInterface)
	if err != nil {
		return err
	}

	if iface.MTU < o.MTU {
		return fmt.Errorf("interface MTU (%d) is too small for specified overlay (%d)", iface.MTU, o.MTU)
	}
	return nil
}

// getSystemMTU retrieves MTU from ovn-kubernetes gateway interafce "br-ex",
// and falls back to use 1500 when "br-ex" mtu is unable to get or less than 0.
func (o *OVNKubernetesConfig) getSystemMTU() {
	link, err := net.InterfaceByName(OVNGatewayInterface)
	if err == nil && link.MTU > 0 {
		o.MTU = link.MTU
	} else {
		o.MTU = defaultMTU
	}
}

// withDefaults returns the default values when ovn.yaml is not provided
func (o *OVNKubernetesConfig) withDefaults() *OVNKubernetesConfig {
	o.OVSInit.DisableOVSInit = false
	o.getSystemMTU()
	return o
}

func newOVNKubernetesConfigFromFile(path string) (*OVNKubernetesConfig, error) {
	o := new(OVNKubernetesConfig)
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buf, &o)
	if err != nil {
		return nil, fmt.Errorf("parsing OVNKubernetes config: %v", err)
	}
	// in case mtu is not defined
	if o.MTU == 0 {
		o.getSystemMTU()
	}
	klog.Infof("parsed OVNKubernetes config from file %q: %+v", path, o)

	return o, nil
}

func NewOVNKubernetesConfigFromFileOrDefault(dir string) (*OVNKubernetesConfig, error) {
	path := filepath.Join(dir, ovnConfigFileName)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			klog.Infof("OVNKubernetes config file not found, assuming default values")
			return new(OVNKubernetesConfig).withDefaults(), nil
		}
		return nil, fmt.Errorf("failed to get OVNKubernetes config file: %v", err)
	}

	o, err := newOVNKubernetesConfigFromFile(path)
	if err == nil {
		return o, nil
	}
	return nil, fmt.Errorf("getting OVNKubernetes config: %v", err)
}

func GetOVNGatewayIP() (string, error) {
	iface, err := net.InterfaceByName(OVNGatewayInterface)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ip := addr.(*net.IPNet).IP
		// return the first available addr, ipv4 takes precedence in ip.String()
		return ip.String(), nil
	}
	return "", fmt.Errorf("failed to get ovn gateway IP address")
}

func ExcludeOVNKubernetesMasqueradeIPs(addrs []net.Addr) []net.Addr {
	var netAddrs []net.Addr
	for _, a := range addrs {
		ipNet, _, _ := net.ParseCIDR(a.String())
		if ipNet.String() != OVNKubernetesV4MasqueradeIP && ipNet.String() != OVNKubernetesV6MasqueradeIP {
			netAddrs = append(netAddrs, a)
		}
	}
	return netAddrs
}

func IsOVNKubernetesInternalInterface(name string) bool {
	excludedInterfacesRegexp := regexp.MustCompile(
		"^[A-Fa-f0-9]{15}|" + // OVN pod interfaces
			"ovn.*|" + // OVN ovn-k8s-mp0 and similar interfaces
			"br-int|" + // OVN integration bridge
			"veth.*|cni.*|" + // Interfaces used in bridge-cni or flannel
			"ovs-system$") // Internal OVS interface

	return excludedInterfacesRegexp.MatchString(name)
}
