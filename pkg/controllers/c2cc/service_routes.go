package c2cc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/openshift/microshift/pkg/config"
	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
)

const (
	c2ccSvcRulePriority = 99
	mgmtPortInterface   = "ovn-k8s-mp0"
)

type serviceRouteManager struct {
	policyRouteTable

	remoteCIDRs   []*net.IPNet
	localSvcCIDRs []*net.IPNet
}

func newServiceRouteManager(cfg *config.Config) *serviceRouteManager {
	remoteCIDRs := cfg.C2CC.ResolvedAllCIDRs

	var localSvcCIDRs []*net.IPNet
	for _, s := range cfg.Network.ServiceNetwork {
		_, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			klog.Warningf("Invalid service network CIDR %q: %v", s, err)
			continue
		}
		localSvcCIDRs = append(localSvcCIDRs, ipNet)
	}

	return &serviceRouteManager{
		policyRouteTable: policyRouteTable{
			table:    cfg.C2CC.ResolvedServiceRouteTableID,
			proto:    cfg.C2CC.ResolvedServiceRouteTableID,
			priority: c2ccSvcRulePriority,
		},
		remoteCIDRs:   remoteCIDRs,
		localSvcCIDRs: localSvcCIDRs,
	}
}

func (m *serviceRouteManager) reconcile(ctx context.Context) error {
	gateways, err := getMgmtPortGateways()
	if err != nil {
		klog.V(4).Infof("Management port not ready: %v, will retry", err)
		return nil
	}

	var desired []netlink.Route
	for _, svcCIDR := range m.localSvcCIDRs {
		family := ipFamilyOf(svcCIDR)
		gw, ok := gateways[family]
		if !ok {
			klog.V(4).Infof("No %s gateway on %s for service CIDR %s, skipping", familyName(family), mgmtPortInterface, svcCIDR)
			continue
		}
		desired = append(desired, netlink.Route{
			Dst:       svcCIDR,
			Gw:        gw.ip,
			Table:     m.table,
			Protocol:  netlink.RouteProtocol(m.proto),
			LinkIndex: gw.linkIdx,
		})
	}

	if err := m.reconcileRoutes(desired); err != nil {
		return fmt.Errorf("failed to reconcile service routes: %w", err)
	}
	if err := m.reconcileRules(); err != nil {
		return fmt.Errorf("failed to reconcile service rules: %w", err)
	}
	return nil
}

func (m *serviceRouteManager) reconcileRules() error {
	type ruleKey struct {
		src string
		dst string
	}

	var desired []netlink.Rule
	desiredKeys := make(map[ruleKey]bool)
	var errs []error
	for _, remoteCIDR := range m.remoteCIDRs {
		for _, svcCIDR := range m.localSvcCIDRs {
			if ipFamilyOf(remoteCIDR) != ipFamilyOf(svcCIDR) {
				continue
			}
			rule := netlink.NewRule()
			rule.Src = remoteCIDR
			rule.Dst = svcCIDR
			rule.Table = m.table
			rule.Priority = m.priority
			rule.Family = ipFamilyOf(remoteCIDR)
			desired = append(desired, *rule)
			desiredKeys[ruleKey{src: remoteCIDR.String(), dst: svcCIDR.String()}] = true
		}
	}

	allRules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("failed to list ip rules: %w", err)
	}

	actualKeys := make(map[ruleKey]netlink.Rule)
	for _, r := range allRules {
		if r.Priority == m.priority && r.Table == m.table && r.Src != nil && r.Dst != nil {
			actualKeys[ruleKey{src: r.Src.String(), dst: r.Dst.String()}] = r
		}
	}

	for _, r := range desired {
		k := ruleKey{src: r.Src.String(), dst: r.Dst.String()}
		if _, exists := actualKeys[k]; exists {
			continue
		}
		rule := r
		if err := netlink.RuleAdd(&rule); err != nil {
			if !errors.Is(err, syscall.EEXIST) {
				klog.Errorf("Failed to add service ip rule from %s to %s: %v", rule.Src, rule.Dst, err)
				errs = append(errs, fmt.Errorf("failed to add service ip rule from %s to %s: %w", rule.Src, rule.Dst, err))
			}
			continue
		}
		klog.V(2).Infof("Service rule add: from %s to %s lookup %d", rule.Src, rule.Dst, m.table)
	}

	for k, r := range actualKeys {
		if desiredKeys[k] {
			continue
		}
		rule := r
		if err := netlink.RuleDel(&rule); err != nil {
			klog.Errorf("Failed to delete stale service rule from %s to %s: %v", k.src, k.dst, err)
			errs = append(errs, fmt.Errorf("failed to delete service ip rule from %s to %s: %w", k.src, k.dst, err))
		}
	}

	return errors.Join(errs...)
}

func (m *serviceRouteManager) cleanup(ctx context.Context) error {
	return errors.Join(m.cleanupRoutes(), m.cleanupRules())
}

type mgmtPortGateway struct {
	ip      net.IP
	linkIdx int
}

func getMgmtPortGateways() (map[int]mgmtPortGateway, error) {
	link, err := netlink.LinkByName(mgmtPortInterface)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", mgmtPortInterface, err)
	}

	linkIdx := link.Attrs().Index
	routes, err := netlink.RouteList(link, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes on %s: %w", mgmtPortInterface, err)
	}

	gateways := make(map[int]mgmtPortGateway)
	for _, r := range routes {
		if r.Gw == nil {
			continue
		}
		if r.Gw.To4() != nil {
			if _, exists := gateways[netlink.FAMILY_V4]; !exists {
				gateways[netlink.FAMILY_V4] = mgmtPortGateway{ip: r.Gw, linkIdx: linkIdx}
			}
		} else {
			if _, exists := gateways[netlink.FAMILY_V6]; !exists {
				gateways[netlink.FAMILY_V6] = mgmtPortGateway{ip: r.Gw, linkIdx: linkIdx}
			}
		}
	}

	if len(gateways) == 0 {
		return nil, fmt.Errorf("no routes with gateway found on %s", mgmtPortInterface)
	}

	return gateways, nil
}

func familyName(family int) string {
	if family == netlink.FAMILY_V4 {
		return "IPv4"
	}
	return "IPv6"
}
