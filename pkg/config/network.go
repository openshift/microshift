package config

import (
	"fmt"
	"net"
	"slices"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/openshift/microshift/pkg/config/ovn"
	"github.com/vishvananda/netlink"
)

type Network struct {
	// IP address pool to use for pod IPs.
	// This field is immutable after installation.
	// +kubebuilder:default={"10.42.0.0/16"}
	ClusterNetwork []string `json:"clusterNetwork"`

	// IP address pool for services.
	// Currently, we only support a single entry here.
	// This field is immutable after installation.
	// +kubebuilder:default={"10.43.0.0/16"}
	ServiceNetwork []string `json:"serviceNetwork"`

	// The port range allowed for Services of type NodePort.
	// If not specified, the default of 30000-32767 will be used.
	// Such Services without a NodePort specified will have one
	// automatically allocated from this range.
	// This parameter can be updated after the cluster is
	// installed.
	// +kubebuilder:validation:Pattern=`^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])-([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`
	// +kubebuilder:default="30000-32767"
	ServiceNodePortRange string `json:"serviceNodePortRange"`

	// The DNS server to use
	DNS string `json:"-"`
}

func (c *Config) computeClusterDNS() (string, error) {
	if len(c.Network.ServiceNetwork) == 0 {
		return "", fmt.Errorf("network.serviceNetwork not filled in")
	}

	clusterDNS, err := getClusterDNS(c.Network.ServiceNetwork[0])
	if err != nil {
		return "", fmt.Errorf("failed to get DNS IP: %v", err)
	}
	return clusterDNS, nil
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

func (c *Config) EnsureNetworksDontOverlap() error {
	linkList, err := netlink.LinkList()
	if err != nil {
		return fmt.Errorf("unable to load NIC list: %v", err)
	}
	for _, link := range linkList {
		if link.Attrs().Name == ovn.OVNGatewayInterface || ovn.IsOVNKubernetesInternalInterface(link.Attrs().Name) {
			continue
		}
		routes, err := netlink.RouteList(link, netlink.FAMILY_ALL)
		if err != nil {
			return fmt.Errorf("unable to determine routes for NIC %q: %v", link.Attrs().Name, err)
		}
		for _, route := range routes {
			if slices.Contains(c.Network.ClusterNetwork, route.Dst.String()) {
				return fmt.Errorf("route %v in the system overlaps with clusterNetwork entries", route.Dst.String())
			}
			if slices.Contains(c.Network.ServiceNetwork, route.Dst.String()) {
				return fmt.Errorf("route %v in the system overlaps with serviceNetwork entries", route.Dst.String())
			}
		}
	}
	return nil
}
