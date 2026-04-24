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
	c2ccRouteTable   = 200
	c2ccRouteProto   = 200
	c2ccRulePriority = 100
)

type linuxRouteManager struct {
	nodeIP      net.IP
	desiredDsts []*net.IPNet
	desiredGWs  map[string]net.IP // cidr string -> nexthop
}

func newLinuxRouteManager(cfg *config.Config) *linuxRouteManager {
	nodeIP := net.ParseIP(cfg.Node.NodeIP)

	m := &linuxRouteManager{
		nodeIP:     nodeIP,
		desiredGWs: make(map[string]net.IP),
	}

	for _, rc := range cfg.C2CC.Resolved {
		allCIDRs := append([]*net.IPNet{}, rc.ClusterNetwork...)
		allCIDRs = append(allCIDRs, rc.ServiceNetwork...)
		for _, cidr := range allCIDRs {
			m.desiredDsts = append(m.desiredDsts, cidr)
			m.desiredGWs[cidr.String()] = rc.NextHop
		}
	}

	return m
}

func (m *linuxRouteManager) reconcile(ctx context.Context) error {
	if err := m.reconcileRoutes(); err != nil {
		return fmt.Errorf("linux routes: %w", err)
	}
	if err := m.reconcileRules(); err != nil {
		return fmt.Errorf("ip rules: %w", err)
	}
	return nil
}

func (m *linuxRouteManager) reconcileRoutes() error {
	linkIdx, err := m.getOutgoingLinkIndex()
	if err != nil {
		return fmt.Errorf("get outgoing link: %w", err)
	}

	actual, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{
		Table:    c2ccRouteTable,
		Protocol: c2ccRouteProto,
	}, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_PROTOCOL)
	if err != nil {
		return fmt.Errorf("listing table %d routes: %w", c2ccRouteTable, err)
	}

	actualByDst := make(map[string]netlink.Route, len(actual))
	for _, r := range actual {
		if r.Dst != nil {
			actualByDst[r.Dst.String()] = r
		}
	}

	for _, cidr := range m.desiredDsts {
		dst := cidr.String()
		if _, exists := actualByDst[dst]; exists {
			delete(actualByDst, dst)
			continue
		}
		route := netlink.Route{
			Dst:       cidr,
			Gw:        m.desiredGWs[dst],
			Table:     c2ccRouteTable,
			Protocol:  c2ccRouteProto,
			LinkIndex: linkIdx,
		}
		if err := netlink.RouteReplace(&route); err != nil {
			klog.Errorf("Failed to add route to %s via %s: %v", dst, route.Gw, err)
			continue
		}
		klog.V(2).Infof("Route add: %s via %s table %d", dst, route.Gw, c2ccRouteTable)
	}

	for dst, r := range actualByDst {
		route := r
		if err := netlink.RouteDel(&route); err != nil {
			klog.Errorf("Failed to delete stale route %s: %v", dst, err)
			continue
		}
		klog.V(2).Infof("Route del: %s table %d (stale)", dst, c2ccRouteTable)
	}

	return nil
}

func (m *linuxRouteManager) reconcileRules() error {
	allRules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("listing ip rules: %w", err)
	}

	actualByDst := make(map[string]netlink.Rule)
	for _, r := range allRules {
		if r.Priority == c2ccRulePriority && r.Table == c2ccRouteTable && r.Dst != nil {
			actualByDst[r.Dst.String()] = r
		}
	}

	for _, cidr := range m.desiredDsts {
		dst := cidr.String()
		if _, exists := actualByDst[dst]; exists {
			delete(actualByDst, dst)
			continue
		}
		rule := netlink.NewRule()
		rule.Dst = cidr
		rule.Table = c2ccRouteTable
		rule.Priority = c2ccRulePriority
		rule.Family = ipFamilyOf(cidr)
		if err := netlink.RuleAdd(rule); err != nil {
			if !errors.Is(err, syscall.EEXIST) {
				klog.Errorf("Failed to add ip rule for %s: %v", dst, err)
			}
			continue
		}
		klog.V(2).Infof("IP rule add: to %s lookup %d priority %d", dst, c2ccRouteTable, c2ccRulePriority)
	}

	for dst, r := range actualByDst {
		rule := r
		if err := netlink.RuleDel(&rule); err != nil {
			klog.Errorf("Failed to delete stale ip rule for %s: %v", dst, err)
			continue
		}
		klog.V(2).Infof("IP rule del: to %s (stale)", dst)
	}

	return nil
}

func (m *linuxRouteManager) cleanup(ctx context.Context) error {
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{
		Table:    c2ccRouteTable,
		Protocol: c2ccRouteProto,
	}, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_PROTOCOL)
	if err == nil {
		for _, r := range routes {
			route := r
			if err := netlink.RouteDel(&route); err != nil {
				klog.Errorf("Failed to cleanup route %v: %v", r.Dst, err)
			}
		}
	}

	allRules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err == nil {
		for _, r := range allRules {
			if r.Priority == c2ccRulePriority && r.Table == c2ccRouteTable {
				rule := r
				if err := netlink.RuleDel(&rule); err != nil {
					klog.Errorf("Failed to cleanup ip rule: %v", err)
				}
			}
		}
	}

	return nil
}

func (m *linuxRouteManager) getOutgoingLinkIndex() (int, error) {
	routes, err := netlink.RouteGet(m.nodeIP)
	if err != nil {
		return 0, fmt.Errorf("route get for %s: %w", m.nodeIP, err)
	}
	if len(routes) == 0 {
		return 0, fmt.Errorf("no route to node IP %s", m.nodeIP)
	}
	return routes[0].LinkIndex, nil
}

func ipFamilyOf(cidr *net.IPNet) int {
	if cidr.IP.To4() != nil {
		return netlink.FAMILY_V4
	}
	return netlink.FAMILY_V6
}
