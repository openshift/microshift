package config

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/vishvananda/netlink"
	"k8s.io/apimachinery/pkg/util/validation"
	netutils "k8s.io/utils/net"
)

type C2CCDNS struct {
	// Maximum TTL (seconds) for positive DNS cache entries in CoreDNS server blocks
	// generated for remote clusters. Must be >= 0. Setting to 0 disables positive caching.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=10
	CacheTTL *int `json:"cacheTTL,omitempty"`
	// Maximum TTL (seconds) for denial (NXDOMAIN/NODATA) DNS cache entries in CoreDNS
	// server blocks generated for remote clusters. Must be >= 0. Setting to 0 disables denial caching.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=10
	CacheNegativeTTL *int `json:"cacheNegativeTTL,omitempty"`
}

type C2CCRouting struct {
	// Linux policy routing table ID for direct routes to remote cluster CIDRs.
	// The route protocol number is set to the same value.
	// Must be between 1 and 252 (253-255 are reserved by the kernel).
	// Must differ from serviceRouteTableID.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=252
	// +kubebuilder:default=200
	RouteTableID *int `json:"routeTableID,omitempty"`
	// Linux policy routing table ID for service routes via the OVN management port.
	// The route protocol number is set to the same value.
	// Must be between 1 and 252 (253-255 are reserved by the kernel).
	// Must differ from routeTableID.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=252
	// +kubebuilder:default=201
	ServiceRouteTableID *int `json:"serviceRouteTableID,omitempty"`
}

type C2CC struct {
	// DNS cache settings for CoreDNS server blocks generated for remote clusters.
	DNS C2CCDNS `json:"dns"`
	// Linux policy routing table settings for C2CC routes.
	Routing C2CCRouting `json:"routing"`
	// List of remote clusters to establish connectivity with.
	// C2CC is disabled when this list is empty.
	RemoteClusters []RemoteCluster `json:"remoteClusters,omitempty"`

	// Interval between healthcheck probe attempts to each remote cluster.
	// Parsed as a Go duration string (e.g. "10s", "1m"). Must be between 1s and 5m.
	// +kubebuilder:default="10s"
	ProbeInterval string `json:"probeInterval,omitempty"`

	// Populated during validation with parsed network objects.
	Resolved                    []ResolvedRemoteCluster `json:"-"`
	ResolvedAllCIDRs            []*net.IPNet            `json:"-"`
	ResolvedProbeInterval       time.Duration           `json:"-"`
	ResolvedRouteTableID        int                     `json:"-"`
	ResolvedServiceRouteTableID int                     `json:"-"`
}

type RemoteCluster struct {
	// IP addresses of the remote cluster's node, used as next-hop for routing.
	// At most one IPv4 and one IPv6 address. Dual-stack clusters need both.
	NextHop []string `json:"nextHop"`
	// Pod CIDRs of the remote cluster. Must not overlap with local cluster or other remotes.
	ClusterNetwork []string `json:"clusterNetwork"`
	// Service CIDRs of the remote cluster. Must not overlap with local cluster or other remotes.
	ServiceNetwork []string `json:"serviceNetwork"`
	// DNS domain suffix for the remote cluster (e.g., "cluster-b.remote").
	// Services are reachable as <svc>.<ns>.svc.<domain>.
	// Optional — if empty, no DNS forwarding is configured for this remote.
	Domain string `json:"domain,omitempty"`
}

type ResolvedRemoteCluster struct {
	NextHops       map[int]net.IP // key: netlink.FAMILY_V4 or netlink.FAMILY_V6
	ClusterNetwork []*net.IPNet
	ServiceNetwork []*net.IPNet
	Domain         string
	DNSIPs         []string       // 10th IP of each ServiceNetwork entry (one per address family)
	ProbeIPs       map[int]string // 11th IP of each ServiceNetwork entry, keyed by family
}

func (rc *ResolvedRemoteCluster) NextHopForFamily(family int) (net.IP, bool) {
	ip, ok := rc.NextHops[family]
	return ip, ok
}

func (rc *ResolvedRemoteCluster) PrimaryNextHop() net.IP {
	// Prefer IPv4 when available (matching OVN-K's IPv4-first ordering convention),
	// fallback to IPv6 for IPv6-only clusters.
	if ip, ok := rc.NextHops[netlink.FAMILY_V4]; ok {
		return ip
	}
	if ip, ok := rc.NextHops[netlink.FAMILY_V6]; ok {
		return ip
	}
	return nil
}

func ipFamilyOfIPNet(ipNet *net.IPNet) int {
	if ipNet.IP.To4() != nil {
		return netlink.FAMILY_V4
	}
	return netlink.FAMILY_V6
}

func familyName(family int) string {
	if family == netlink.FAMILY_V4 {
		return "IPv4"
	}
	return "IPv6"
}

func (rc *ResolvedRemoteCluster) AllCIDRs() []*net.IPNet {
	all := make([]*net.IPNet, 0, len(rc.ClusterNetwork)+len(rc.ServiceNetwork))
	all = append(all, rc.ClusterNetwork...)
	all = append(all, rc.ServiceNetwork...)
	return all
}

type labeledCIDR struct {
	net *net.IPNet
	str string
}

func (c *C2CC) IsEnabled() bool {
	return len(c.RemoteClusters) > 0
}

func (c *C2CC) AllRemoteCIDRStrings() []string {
	strs := make([]string, len(c.ResolvedAllCIDRs))
	for i, cidr := range c.ResolvedAllCIDRs {
		strs[i] = cidr.String()
	}
	return strs
}

func (rc *RemoteCluster) isEmpty() bool {
	return len(rc.NextHop) == 0 && len(rc.ClusterNetwork) == 0 && len(rc.ServiceNetwork) == 0 && rc.Domain == ""
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

func (c *C2CC) parseRemoteClusters() ([]ResolvedRemoteCluster, []error) {
	resolved := make([]ResolvedRemoteCluster, len(c.RemoteClusters))
	var errs []error

	for i := range c.RemoteClusters {
		rc := &c.RemoteClusters[i]
		label := fmt.Sprintf("remoteClusters[%d]", i)

		if len(rc.NextHop) == 0 {
			errs = append(errs, fmt.Errorf("%s.nextHop must not be empty", label))
		}
		resolved[i].NextHops = make(map[int]net.IP, len(rc.NextHop))
		for j, hopStr := range rc.NextHop {
			ip := net.ParseIP(hopStr)
			if ip == nil {
				errs = append(errs, fmt.Errorf("%s.nextHop[%d] %q is not a valid IP address", label, j, hopStr))
				continue
			}
			family := netlink.FAMILY_V4
			if ip.To4() == nil {
				family = netlink.FAMILY_V6
			}
			if _, dup := resolved[i].NextHops[family]; dup {
				errs = append(errs, fmt.Errorf("%s.nextHop has multiple %s addresses (max 1 per family)", label, familyName(family)))
				continue
			}
			resolved[i].NextHops[family] = ip
		}

		if len(rc.ClusterNetwork) == 0 {
			errs = append(errs, fmt.Errorf("%s.clusterNetwork must not be empty", label))
		}
		if len(rc.ServiceNetwork) == 0 {
			errs = append(errs, fmt.Errorf("%s.serviceNetwork must not be empty", label))
		}

		for j, cidr := range rc.ClusterNetwork {
			if ipNet := parseAndValidateCIDR(cidr, fmt.Sprintf("%s.clusterNetwork[%d]", label, j), &errs); ipNet != nil {
				resolved[i].ClusterNetwork = append(resolved[i].ClusterNetwork, ipNet)
			}
		}
		for j, cidr := range rc.ServiceNetwork {
			if ipNet := parseAndValidateCIDR(cidr, fmt.Sprintf("%s.serviceNetwork[%d]", label, j), &errs); ipNet != nil {
				resolved[i].ServiceNetwork = append(resolved[i].ServiceNetwork, ipNet)
			}
		}

		resolved[i].Domain = rc.Domain

		// Compute probe IP for each service network (one per address family)
		resolved[i].ProbeIPs = make(map[int]string, len(resolved[i].ServiceNetwork))
		for j, svcNet := range resolved[i].ServiceNetwork {
			probeIP, err := cidr.Host(svcNet, 11)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: failed to compute probe IP from serviceNetwork[%d]: %w", label, j, err))
				continue
			}
			family := ipFamilyOfIPNet(svcNet)
			resolved[i].ProbeIPs[family] = probeIP.String()
		}
	}

	return resolved, errs
}

func parseAndValidateCIDR(cidr, field string, errs *[]error) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		*errs = append(*errs, fmt.Errorf("%s %q is not a valid CIDR: %w", field, cidr, err))
		return nil
	}

	ones, _ := ipNet.Mask.Size()
	if ipNet.IP.To4() != nil {
		if ones < 8 {
			*errs = append(*errs, fmt.Errorf("%s %q has mask /%d shorter than minimum /8", field, cidr, ones))
			return nil
		}
	} else {
		if ones < 32 {
			*errs = append(*errs, fmt.Errorf("%s %q has mask /%d shorter than minimum /32", field, cidr, ones))
			return nil
		}
	}
	return ipNet
}

func (c *C2CC) resolveRoutingDefaults() {
	routeTable := 200
	if c.Routing.RouteTableID != nil {
		routeTable = *c.Routing.RouteTableID
	}
	svcRouteTable := 201
	if c.Routing.ServiceRouteTableID != nil {
		svcRouteTable = *c.Routing.ServiceRouteTableID
	}
	c.ResolvedRouteTableID = routeTable
	c.ResolvedServiceRouteTableID = svcRouteTable
}

func (c *C2CC) validateRouting() error {
	c.resolveRoutingDefaults()

	var errs []error
	if c.ResolvedRouteTableID < 1 || c.ResolvedRouteTableID > 252 {
		errs = append(errs, fmt.Errorf("routing.routeTableID must be between 1 and 252, got %d", c.ResolvedRouteTableID))
	}
	if c.ResolvedServiceRouteTableID < 1 || c.ResolvedServiceRouteTableID > 252 {
		errs = append(errs, fmt.Errorf("routing.serviceRouteTableID must be between 1 and 252, got %d", c.ResolvedServiceRouteTableID))
	}
	if c.ResolvedRouteTableID == c.ResolvedServiceRouteTableID {
		errs = append(errs, fmt.Errorf("routing.routeTableID (%d) and routing.serviceRouteTableID (%d) must differ",
			c.ResolvedRouteTableID, c.ResolvedServiceRouteTableID))
	}
	return errors.Join(errs...)
}

func (d *C2CCDNS) validate() error {
	var errs []error
	if d.CacheTTL != nil && *d.CacheTTL < 0 {
		errs = append(errs, fmt.Errorf("dns.cacheTTL must be >= 0, got %d", *d.CacheTTL))
	}
	if d.CacheNegativeTTL != nil && *d.CacheNegativeTTL < 0 {
		errs = append(errs, fmt.Errorf("dns.cacheNegativeTTL must be >= 0, got %d", *d.CacheNegativeTTL))
	}
	return errors.Join(errs...)
}

func (c *C2CC) validate(cfg *Config) error {
	if cfg.Network.CNIPlugin != CniPluginUnset && cfg.Network.CNIPlugin != CniPluginOVNK {
		return fmt.Errorf("cluster to cluster requires OVN-Kubernetes CNI (network.cniPlugin must be \"\" or \"ovnk\", got %q)", cfg.Network.CNIPlugin)
	}

	if err := c.DNS.validate(); err != nil {
		return err
	}

	if err := c.validateRouting(); err != nil {
		return err
	}

	resolved, parseErrs := c.parseRemoteClusters()
	if len(parseErrs) > 0 {
		return errors.Join(parseErrs...)
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

	seenCIDRs := make([]labeledCIDR, 0, len(cfg.Network.ClusterNetwork)+len(cfg.Network.ServiceNetwork)+len(c.RemoteClusters)*4)
	for _, s := range cfg.Network.ClusterNetwork {
		if _, ipNet, err := net.ParseCIDR(s); err == nil {
			seenCIDRs = append(seenCIDRs, labeledCIDR{net: ipNet, str: s})
		}
	}
	for _, s := range cfg.Network.ServiceNetwork {
		if _, ipNet, err := net.ParseCIDR(s); err == nil {
			seenCIDRs = append(seenCIDRs, labeledCIDR{net: ipNet, str: s})
		}
	}

	for i := range c.RemoteClusters {
		errs = append(errs, validateRemoteCluster(i, &c.RemoteClusters[i], &resolved[i],
			nodeIP, nodeIPv6, localV4, localV6, hostIPs,
			seenNextHops, seenRemoteDomains, &seenCIDRs)...)
	}

	probeInterval, err := c.validateProbeInterval()
	if err != nil {
		errs = append(errs, err)
	}

	if err := errors.Join(errs...); err != nil {
		return err
	}

	c.Resolved = resolved
	c.ResolvedProbeInterval = probeInterval

	var allCIDRs []*net.IPNet
	for i := range resolved {
		allCIDRs = append(allCIDRs, resolved[i].AllCIDRs()...)
	}
	c.ResolvedAllCIDRs = allCIDRs

	return nil
}

func (c *C2CC) validateProbeInterval() (time.Duration, error) {
	d, err := time.ParseDuration(c.ProbeInterval)
	if err != nil {
		return 0, fmt.Errorf("probeInterval %q is not a valid duration: %w", c.ProbeInterval, err)
	}
	if d < 1*time.Second || d > 5*time.Minute {
		return 0, fmt.Errorf("probeInterval must be between 1s and 5m, got %s", d)
	}
	return d, nil
}

func validateRemoteCluster(
	i int, rc *RemoteCluster, res *ResolvedRemoteCluster,
	nodeIP, nodeIPv6 net.IP, localV4, localV6 bool, hostIPs []net.IP,
	seenNextHops map[string]int, seenRemoteDomains map[string]int, seenCIDRs *[]labeledCIDR,
) []error {
	label := fmt.Sprintf("remoteClusters[%d]", i)
	var errs []error

	for _, hop := range res.NextHops {
		normalized := hop.String()
		if hop.Equal(nodeIP) || (nodeIPv6 != nil && hop.Equal(nodeIPv6)) {
			errs = append(errs, fmt.Errorf("%s.nextHop %q must not equal the local node IP (routing loop)", label, normalized))
		}
		if prev, ok := seenNextHops[normalized]; ok {
			errs = append(errs, fmt.Errorf("%s.nextHop %q duplicates remoteClusters[%d]", label, normalized, prev))
		} else {
			seenNextHops[normalized] = i
		}
	}

	for j, cidrNet := range res.ClusterNetwork {
		errs = append(errs, checkCIDRConflicts(cidrNet, rc.ClusterNetwork[j], label, *seenCIDRs, hostIPs)...)
		*seenCIDRs = append(*seenCIDRs, labeledCIDR{net: cidrNet, str: rc.ClusterNetwork[j]})
	}
	for j, cidrNet := range res.ServiceNetwork {
		errs = append(errs, checkCIDRConflicts(cidrNet, rc.ServiceNetwork[j], label, *seenCIDRs, hostIPs)...)
		*seenCIDRs = append(*seenCIDRs, labeledCIDR{net: cidrNet, str: rc.ServiceNetwork[j]})
	}

	if rc.Domain != "" {
		if domainErrs := validation.IsDNS1123Subdomain(rc.Domain); len(domainErrs) > 0 {
			errs = append(errs, fmt.Errorf("%s.domain %q is not a valid DNS name: %s", label, rc.Domain, strings.Join(domainErrs, ", ")))
		}
		if rc.Domain == "cluster.local" {
			errs = append(errs, fmt.Errorf("%s.domain cannot be cluster.local", label))
		}
		if prev, ok := seenRemoteDomains[rc.Domain]; ok {
			errs = append(errs, fmt.Errorf("%s.domain %q duplicates remoteClusters[%d]", label, rc.Domain, prev))
		} else {
			seenRemoteDomains[rc.Domain] = i
		}
	}

	if rc.Domain != "" && len(rc.ServiceNetwork) > 0 {
		for i, svcNetStr := range rc.ServiceNetwork {
			dnsIP, err := getClusterDNS(svcNetStr)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: failed to compute DNS IP from serviceNetwork[%d] %q: %w", label, i, svcNetStr, err))
			} else {
				res.DNSIPs = append(res.DNSIPs, dnsIP)
			}
		}
	}

	errs = append(errs, validateIPFamilyConsistencyNets(res.ClusterNetwork, label+".clusterNetwork")...)
	errs = append(errs, validateIPFamilyConsistencyNets(res.ServiceNetwork, label+".serviceNetwork")...)
	errs = append(errs, validateNetworkShapeNets(res.ClusterNetwork, res.ServiceNetwork, label)...)
	errs = append(errs, validateRemoteIPFamilyCompatibility(localV4, localV6, res.ClusterNetwork, label)...)
	errs = append(errs, validateRemoteIPFamilyCompatibility(localV4, localV6, res.ServiceNetwork, label)...)
	errs = append(errs, validateNextHopCoverage(res.NextHops, res.ClusterNetwork, res.ServiceNetwork, label)...)

	return errs
}

func validateIPFamilyConsistencyNets(cidrs []*net.IPNet, field string) []error {
	var v4, v6 int
	var errs []error
	for _, c := range cidrs {
		switch netutils.IPFamilyOfCIDR(c) {
		case netutils.IPv4:
			v4++
		case netutils.IPv6:
			v6++
		case netutils.IPFamilyUnknown:
			errs = append(errs, fmt.Errorf("%s has unrecognized IP family: %v", field, c))
		}
	}
	if v4 > 1 {
		errs = append(errs, fmt.Errorf("%s has multiple IPv4 entries (max 1)", field))
	}
	if v6 > 1 {
		errs = append(errs, fmt.Errorf("%s has multiple IPv6 entries (max 1)", field))
	}
	return errs
}

func validateNetworkShapeNets(clusterNetwork, serviceNetwork []*net.IPNet, label string) []error {
	if len(clusterNetwork) != len(serviceNetwork) {
		return []error{fmt.Errorf("%s: clusterNetwork and serviceNetwork have different cardinality (%d vs %d)",
			label, len(clusterNetwork), len(serviceNetwork))}
	}
	var errs []error
	for i := 0; i < len(clusterNetwork); i++ {
		cFamily := netutils.IPFamilyOfCIDR(clusterNetwork[i])
		sFamily := netutils.IPFamilyOfCIDR(serviceNetwork[i])
		if cFamily != netutils.IPFamilyUnknown && sFamily != netutils.IPFamilyUnknown && cFamily != sFamily {
			errs = append(errs, fmt.Errorf("%s: clusterNetwork[%d] and serviceNetwork[%d] have mismatched IP families", label, i, i))
		}
	}
	return errs
}

func validateRemoteIPFamilyCompatibility(localV4, localV6 bool, remoteCIDRs []*net.IPNet, label string) []error {
	var errs []error
	for _, c := range remoteCIDRs {
		family := netutils.IPFamilyOfCIDR(c)
		if family == netutils.IPv4 && !localV4 {
			errs = append(errs, fmt.Errorf("%s: %q is IPv4 but local cluster has no IPv4 network", label, c))
		}
		if family == netutils.IPv6 && !localV6 {
			errs = append(errs, fmt.Errorf("%s: %q is IPv6 but local cluster has no IPv6 network", label, c))
		}
	}
	return errs
}

func validateNextHopCoverage(nextHops map[int]net.IP, clusterNetwork, serviceNetwork []*net.IPNet, label string) []error {
	needed := make(map[int]bool)
	for _, cidr := range clusterNetwork {
		needed[ipFamilyOfIPNet(cidr)] = true
	}
	for _, cidr := range serviceNetwork {
		needed[ipFamilyOfIPNet(cidr)] = true
	}
	var errs []error
	for family := range needed {
		if _, ok := nextHops[family]; !ok {
			errs = append(errs, fmt.Errorf("%s has %s CIDRs but no %s nextHop", label, familyName(family), familyName(family)))
		}
	}
	return errs
}

func checkCIDRConflicts(cidr *net.IPNet, cidrStr, label string, seenCIDRs []labeledCIDR, hostIPs []net.IP) []error {
	var errs []error
	for _, existing := range seenCIDRs {
		if cidrsOverlap(cidr, existing.net) {
			errs = append(errs, fmt.Errorf("%s: CIDR %q overlaps with %q", label, cidrStr, existing.str))
		}
	}
	for _, hostIP := range hostIPs {
		if cidr.Contains(hostIP) {
			errs = append(errs, fmt.Errorf("remote CIDR %q contains host interface IP %s — this would disrupt management traffic", cidrStr, hostIP))
		}
	}
	return errs
}

func cidrsOverlap(a, b *net.IPNet) bool {
	return a.Contains(b.IP) || b.Contains(a.IP)
}

// RenderC2CCDNSBlocks generates CoreDNS server blocks for cross-cluster DNS.
func RenderC2CCDNSBlocks(resolved []ResolvedRemoteCluster, cacheTTL, cacheNegativeTTL int) string {
	var blocks []string
	for _, rc := range resolved {
		if rc.Domain == "" {
			continue
		}
		blocks = append(blocks, formatDNSBlock(rc.Domain, rc.DNSIPs, cacheTTL, cacheNegativeTTL))
	}
	if len(blocks) == 0 {
		return ""
	}
	return "\n" + strings.Join(blocks, "\n")
}

func formatDNSBlock(domain string, dnsIPs []string, cacheTTL, cacheNegativeTTL int) string {
	return fmt.Sprintf(`    %s:5353 {
        bufsize 1232
        errors
        log . {
            class error
        }
        rewrite stop name suffix .%s .cluster.local answer auto
        forward . %s
        cache %d {
            denial 9984 %d
        }
    }`, domain, domain, strings.Join(dnsIPs, " "), cacheTTL, cacheNegativeTTL)
}
