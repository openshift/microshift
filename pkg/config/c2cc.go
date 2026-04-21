package config

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/vishvananda/netlink"
	"k8s.io/apimachinery/pkg/util/validation"
	netutils "k8s.io/utils/net"
)

type C2CC struct {
	// List of remote clusters to establish connectivity with.
	// C2CC is disabled when this list is empty.
	RemoteClusters []RemoteCluster `json:"remoteClusters,omitempty"`
}

type RemoteCluster struct {
	// IP address of the remote cluster's node, used as next-hop for routing.
	NextHop string `json:"nextHop"`
	// Pod CIDRs of the remote cluster. Must not overlap with local cluster or other remotes.
	ClusterNetwork []string `json:"clusterNetwork"`
	// Service CIDRs of the remote cluster. Must not overlap with local cluster or other remotes.
	ServiceNetwork []string `json:"serviceNetwork"`
	// DNS domain suffix for the remote cluster (e.g., "cluster-b.remote").
	// Services are reachable as <svc>.<ns>.svc.<domain>.
	// Optional — if empty, no DNS forwarding is configured for this remote.
	Domain string `json:"domain,omitempty"`
}

func (c *C2CC) IsEnabled() bool {
	return len(c.RemoteClusters) > 0
}

func (rc *RemoteCluster) isEmpty() bool {
	return rc.NextHop == "" && len(rc.ClusterNetwork) == 0 && len(rc.ServiceNetwork) == 0 && rc.Domain == ""
}

func (c *C2CC) stripEmptyRemoteClusters() {
	filtered := c.RemoteClusters[:0]
	for i := range c.RemoteClusters {
		if !c.RemoteClusters[i].isEmpty() {
			filtered = append(filtered, c.RemoteClusters[i])
		}
	}
	c.RemoteClusters = filtered
}

var getHostIPs = defaultGetHostIPs

func defaultGetHostIPs() ([]net.IP, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("listing network interfaces: %w", err)
	}
	var ips []net.IP
	for _, link := range links {
		addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return nil, fmt.Errorf("listing addresses for interface %q: %w", link.Attrs().Name, err)
		}
		for _, addr := range addrs {
			ips = append(ips, addr.IP)
		}
	}
	return ips, nil
}

func (c *C2CC) validate(cfg *Config) error {
	if cfg.Network.CNIPlugin != CniPluginUnset && cfg.Network.CNIPlugin != CniPluginOVNK {
		return fmt.Errorf("c2cc requires OVN-Kubernetes CNI (network.cniPlugin must be \"\" or \"ovnk\", got %q)", cfg.Network.CNIPlugin)
	}

	hostIPs, err := getHostIPs()
	if err != nil {
		return fmt.Errorf("failed to get host IPs: %w", err)
	}

	nodeIP := net.ParseIP(cfg.Node.NodeIP)
	if nodeIP == nil {
		return fmt.Errorf("failed to parse cfg.Node.NodeIP (%q)", cfg.Node.NodeIP)
	}

	var nodeIPv6 net.IP
	if cfg.Node.NodeIPV6 != "" {
		nodeIPv6 = net.ParseIP(cfg.Node.NodeIPV6)
		if nodeIPv6 == nil {
			return fmt.Errorf("failed to parse cfg.Node.NodeIPV6 (%q)", cfg.Node.NodeIPV6)
		}
	}

	localV4 := cfg.IsIPv4()
	localV6 := cfg.IsIPv6()

	var errs []error

	seenNextHops := make(map[string]int, len(c.RemoteClusters))
	seenRemoteDomains := make(map[string]int, len(c.RemoteClusters))

	// Overlap check: start with local CIDRs, then accumulate each remote's CIDRs.
	seenCIDRs := make([]string, 0, len(cfg.Network.ClusterNetwork)+len(cfg.Network.ServiceNetwork)+len(c.RemoteClusters)*4)
	seenCIDRs = append(seenCIDRs, cfg.Network.ClusterNetwork...)
	seenCIDRs = append(seenCIDRs, cfg.Network.ServiceNetwork...)

	for i := range c.RemoteClusters {
		rc := &c.RemoteClusters[i]
		label := fmt.Sprintf("remoteClusters[%d]", i)

		ip := net.ParseIP(rc.NextHop)
		if ip == nil {
			errs = append(errs, fmt.Errorf("%s.nextHop %q is not a valid IP address", label, rc.NextHop))
		} else {
			normalizedNextHop := ip.String()
			if ip.Equal(nodeIP) || (nodeIPv6 != nil && ip.Equal(nodeIPv6)) {
				errs = append(errs, fmt.Errorf("%s.nextHop %q must not equal the local node IP (routing loop)", label, normalizedNextHop))
			}
			if prev, ok := seenNextHops[normalizedNextHop]; ok {
				errs = append(errs, fmt.Errorf("%s.nextHop %q duplicates remoteClusters[%d]", label, normalizedNextHop, prev))
			} else {
				seenNextHops[normalizedNextHop] = i
			}
		}

		if len(rc.ClusterNetwork) == 0 {
			errs = append(errs, fmt.Errorf("%s.clusterNetwork must not be empty", label))
		}
		if len(rc.ServiceNetwork) == 0 {
			errs = append(errs, fmt.Errorf("%s.serviceNetwork must not be empty", label))
		}

		for j, cidr := range rc.ClusterNetwork {
			errs = append(errs, validateRemoteCIDR(cidr, fmt.Sprintf("%s.clusterNetwork[%d]", label, j)))
			errs = append(errs, checkCIDRConflicts(cidr, label, seenCIDRs, hostIPs)...)
			seenCIDRs = append(seenCIDRs, cidr)
		}
		for j, cidr := range rc.ServiceNetwork {
			errs = append(errs, validateRemoteCIDR(cidr, fmt.Sprintf("%s.serviceNetwork[%d]", label, j)))
			errs = append(errs, checkCIDRConflicts(cidr, label, seenCIDRs, hostIPs)...)
			seenCIDRs = append(seenCIDRs, cidr)
		}

		if rc.Domain != "" {
			if domainErrs := validation.IsDNS1123Subdomain(rc.Domain); len(domainErrs) > 0 {
				errs = append(errs, fmt.Errorf("%s.domain %q is not a valid DNS name: %s", label, rc.Domain, strings.Join(domainErrs, ", ")))
			}
			if prev, ok := seenRemoteDomains[rc.Domain]; ok {
				errs = append(errs, fmt.Errorf("%s.domain %q duplicates remoteClusters[%d]", label, rc.Domain, prev))
			} else {
				seenRemoteDomains[rc.Domain] = i
			}
		}

		errs = append(errs, validateIPFamilyConsistency(rc.ClusterNetwork, label+".clusterNetwork")...)
		errs = append(errs, validateIPFamilyConsistency(rc.ServiceNetwork, label+".serviceNetwork")...)
		errs = append(errs, validateNetworkShape(rc.ClusterNetwork, rc.ServiceNetwork, label)...)
		errs = append(errs, validateRemoteIPFamilyCompatibility(localV4, localV6, rc.ClusterNetwork, label)...)
		errs = append(errs, validateRemoteIPFamilyCompatibility(localV4, localV6, rc.ServiceNetwork, label)...)
	}

	return errors.Join(errs...)
}

func validateRemoteCIDR(cidr, field string) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("%s %q is not a valid CIDR: %w", field, cidr, err)
	}

	ones, _ := ipNet.Mask.Size()
	if ipNet.IP.To4() != nil {
		if ones < 8 {
			return fmt.Errorf("%s %q has mask /%d shorter than minimum /8", field, cidr, ones)
		}
	} else {
		if ones < 32 {
			return fmt.Errorf("%s %q has mask /%d shorter than minimum /32", field, cidr, ones)
		}
	}
	return nil
}

func validateIPFamilyConsistency(cidrs []string, field string) []error {
	var v4, v6 int
	for _, c := range cidrs {
		switch netutils.IPFamilyOfCIDRString(c) {
		case netutils.IPv4:
			v4++
		case netutils.IPv6:
			v6++
		case netutils.IPFamilyUnknown:
		}
	}
	var errs []error
	if v4 > 1 {
		errs = append(errs, fmt.Errorf("%s has multiple IPv4 entries (max 1)", field))
	}
	if v6 > 1 {
		errs = append(errs, fmt.Errorf("%s has multiple IPv6 entries (max 1)", field))
	}
	return errs
}

func validateNetworkShape(clusterNetwork, serviceNetwork []string, label string) []error {
	if len(clusterNetwork) != len(serviceNetwork) {
		return []error{fmt.Errorf("%s: clusterNetwork and serviceNetwork have different cardinality (%d vs %d)",
			label, len(clusterNetwork), len(serviceNetwork))}
	}
	var errs []error
	for i := 0; i < len(clusterNetwork); i++ {
		cFamily := netutils.IPFamilyOfCIDRString(clusterNetwork[i])
		sFamily := netutils.IPFamilyOfCIDRString(serviceNetwork[i])
		if cFamily != netutils.IPFamilyUnknown && sFamily != netutils.IPFamilyUnknown && cFamily != sFamily {
			errs = append(errs, fmt.Errorf("%s: clusterNetwork[%d] and serviceNetwork[%d] have mismatched IP families", label, i, i))
		}
	}
	return errs
}

func validateRemoteIPFamilyCompatibility(localV4, localV6 bool, remoteCIDRs []string, label string) []error {
	var errs []error
	for _, c := range remoteCIDRs {
		family := netutils.IPFamilyOfCIDRString(c)
		if family == netutils.IPv4 && !localV4 {
			errs = append(errs, fmt.Errorf("%s: %q is IPv4 but local cluster has no IPv4 network", label, c))
		}
		if family == netutils.IPv6 && !localV6 {
			errs = append(errs, fmt.Errorf("%s: %q is IPv6 but local cluster has no IPv6 network", label, c))
		}
	}
	return errs
}

func checkCIDRConflicts(cidr, label string, seenCIDRs []string, hostIPs []net.IP) []error {
	var errs []error
	for _, existing := range seenCIDRs {
		if cidrsOverlap(cidr, existing) {
			errs = append(errs, fmt.Errorf("%s: CIDR %q overlaps with %q", label, cidr, existing))
		}
	}
	if _, ipNet, err := net.ParseCIDR(cidr); err == nil {
		for _, hostIP := range hostIPs {
			if ipNet.Contains(hostIP) {
				errs = append(errs, fmt.Errorf("remote CIDR %q contains host interface IP %s — this would disrupt management traffic", cidr, hostIP))
			}
		}
	}
	return errs
}

func cidrsOverlap(a, b string) bool {
	_, netA, errA := net.ParseCIDR(a)
	_, netB, errB := net.ParseCIDR(b)
	if errA != nil || errB != nil {
		return false
	}
	return netA.Contains(netB.IP) || netB.Contains(netA.IP)
}
