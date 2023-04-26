package ovn

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"

	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	ovnConfigFileName           = "ovn.yaml"
	OVNGatewayInterface         = "br-ex"
	OVNExternalGatewayInterface = "br-ex1"
	defaultMTU                  = 1500
	OVNKubernetesV4MasqueradeIP = "169.254.169.2"
	OVNKubernetesV6MasqueradeIP = "fd69::2"

	// used for multinode ovn database transport
	OVN_NB_PORT = "9641"
	OVN_SB_PORT = "9642"

	// geneve header length for IPv4
	GeneveHeaderLengthIPv4 = 58
	// geneve header length for IPv6
	GeneveHeaderLengthIPv6 = GeneveHeaderLengthIPv4 + 20
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
	// Uplink interface for OVS bridge "br-ex1"
	ExternalGatewayInterface string `json:"externalGatewayInterface,omitempty"`
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
	if o.OVSInit.ExternalGatewayInterface != "" {
		_, err := net.InterfaceByName(o.OVSInit.ExternalGatewayInterface)
		if err != nil {
			return fmt.Errorf("external gateway interface %s not found", o.OVSInit.ExternalGatewayInterface)
		}
		_, err = net.InterfaceByName(OVNExternalGatewayInterface)
		if err != nil {
			return fmt.Errorf("external gateway interface %s is configured, but external gateway bridge %s not found",
				o.OVSInit.ExternalGatewayInterface, OVNExternalGatewayInterface)
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

// getClusterMTU retrieves MTU from ovn-kubernetes gateway interface "br-ex",
// and falls back to use 1500 when "br-ex" mtu is unable to get or less than 0.
func (o *OVNKubernetesConfig) getClusterMTU(multinode bool) {
	link, err := net.InterfaceByName(OVNGatewayInterface)
	if err == nil && link.MTU > 0 {
		o.MTU = link.MTU
	} else {
		o.MTU = defaultMTU
	}

	if multinode {
		o.MTU = o.MTU - GeneveHeaderLengthIPv6
	}
}

// withDefaults returns the default values when ovn.yaml is not provided
func (o *OVNKubernetesConfig) withDefaults(multinode bool) *OVNKubernetesConfig {
	o.OVSInit.DisableOVSInit = false
	o.getClusterMTU(multinode)
	return o
}

func newOVNKubernetesConfigFromFile(path string, multinode bool) (*OVNKubernetesConfig, error) {
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
		o.getClusterMTU(multinode)
	}
	klog.Infof("parsed OVNKubernetes config from file %q: %+v", path, o)

	return o, nil
}

func NewOVNKubernetesConfigFromFileOrDefault(dir string, multinode bool) (*OVNKubernetesConfig, error) {
	path := filepath.Join(dir, ovnConfigFileName)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			klog.Infof("OVNKubernetes config file not found, assuming default values")
			return new(OVNKubernetesConfig).withDefaults(multinode), nil
		}
		return nil, fmt.Errorf("failed to get OVNKubernetes config file: %v", err)
	}

	o, err := newOVNKubernetesConfigFromFile(path, multinode)
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

func HasDefaultGateway(family int) (bool, error) {
	// filter the default route to obtain the gateway
	filter := &netlink.Route{Dst: nil}
	mask := netlink.RT_FILTER_DST

	routeList, err := netlink.RouteListFiltered(family, filter, mask)
	if err != nil {
		return false, fmt.Errorf("failed to get routing table in node: %v", err)
	}

	link, err := netlink.LinkByName(OVNGatewayInterface)
	if err != nil {
		return false, fmt.Errorf("error looking up gw interface %s: %v", OVNGatewayInterface, err)
	}
	if link.Attrs() == nil {
		return false, fmt.Errorf("no attributes found for link: %#v", link)
	}
	// filter routes pass br-ex
	routes := filterRoutesByIfIndex(routeList, link.Attrs().Index)
	for _, r := range routes {
		// no multipath
		if len(r.MultiPath) == 0 {
			if r.Gw == nil {
				klog.Infof("Failed to get gateway for route %v : %v", r, err)
				continue
			}
			klog.Infof("Found gateway for route %v", r)
			return true, nil
		}

		// multipath, use the first valid entry
		for _, nh := range r.MultiPath {
			if nh.Gw == nil {
				klog.Infof("Failed to get gateway for multipath route %v : %v", nh, err)
				continue
			}
			klog.Infof("Found gateway for multipath route %v", r)
			return true, nil
		}
	}
	return false, nil
}

func filterRoutesByIfIndex(routesUnfiltered []netlink.Route, ifIdx int) []netlink.Route {
	if ifIdx <= 0 {
		return routesUnfiltered
	}
	var routes []netlink.Route
	for _, r := range routesUnfiltered {
		if r.LinkIndex == ifIdx ||
			multipathRouteMatchesIfIndex(r, ifIdx) {
			routes = append(routes, r)
		}
	}
	return routes
}

func multipathRouteMatchesIfIndex(r netlink.Route, ifIdx int) bool {
	if r.LinkIndex != 0 {
		return false
	}
	if len(r.MultiPath) == 0 {
		return false
	}
	for _, mr := range r.MultiPath {
		if mr.LinkIndex != ifIdx {
			return false
		}
	}
	return true
}
