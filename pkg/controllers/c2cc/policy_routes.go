package c2cc

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
)

type policyRouteTable struct {
	table    int
	proto    int
	priority int
}

func (t *policyRouteTable) reconcileRoutes(desired []netlink.Route) error {
	actual, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{
		Table:    t.table,
		Protocol: netlink.RouteProtocol(t.proto),
	}, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_PROTOCOL)
	if err != nil {
		return fmt.Errorf("listing table %d routes: %w", t.table, err)
	}

	actualByDst := make(map[string]netlink.Route, len(actual))
	for _, r := range actual {
		if r.Dst != nil {
			actualByDst[r.Dst.String()] = r
		}
	}

	desiredByDst := make(map[string]bool, len(desired))
	for _, r := range desired {
		dst := r.Dst.String()
		desiredByDst[dst] = true
		route := r
		if actual, exists := actualByDst[dst]; exists && actual.Gw.Equal(route.Gw) {
			continue
		}
		if err := netlink.RouteReplace(&route); err != nil {
			klog.Errorf("Failed to add route to %s via %s: %v", dst, route.Gw, err)
			continue
		}
		klog.V(2).Infof("Route add: %s via %s table %d", dst, route.Gw, t.table)
	}

	for dst, r := range actualByDst {
		if desiredByDst[dst] {
			continue
		}
		route := r
		if err := netlink.RouteDel(&route); err != nil {
			klog.Errorf("Failed to delete stale route %s: %v", dst, err)
			continue
		}
		klog.V(2).Infof("Route del: %s table %d (stale)", dst, t.table)
	}

	return nil
}

func (t *policyRouteTable) cleanupRoutes() error {
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{
		Table:    t.table,
		Protocol: netlink.RouteProtocol(t.proto),
	}, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_PROTOCOL)
	if err != nil {
		return fmt.Errorf("listing table %d routes for cleanup: %w", t.table, err)
	}
	for _, r := range routes {
		route := r
		if err := netlink.RouteDel(&route); err != nil {
			klog.Errorf("Failed to cleanup route %v: %v", r.Dst, err)
		}
	}
	return nil
}

func (t *policyRouteTable) cleanupRules() error {
	allRules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("listing ip rules for cleanup: %w", err)
	}
	for _, r := range allRules {
		if r.Priority == t.priority && r.Table == t.table {
			rule := r
			if err := netlink.RuleDel(&rule); err != nil {
				klog.Errorf("Failed to cleanup ip rule: %v", err)
			}
		}
	}
	return nil
}

func (t *policyRouteTable) subscribe(reconcileCh chan<- string, reason string) (chan struct{}, error) {
	routeUpdates := make(chan netlink.RouteUpdate, 100)
	done := make(chan struct{})

	if err := netlink.RouteSubscribe(routeUpdates, done); err != nil {
		return nil, fmt.Errorf("subscribe to route events: %w", err)
	}

	go func() {
		for update := range routeUpdates {
			if update.Table != t.table {
				continue
			}
			select {
			case reconcileCh <- reason:
			default:
			}
		}
	}()

	klog.V(2).Infof("Subscribed to netlink route events for table %d", t.table)
	return done, nil
}

func ipFamilyOf(cidr *net.IPNet) int {
	if cidr.IP.To4() != nil {
		return netlink.FAMILY_V4
	}
	return netlink.FAMILY_V6
}
