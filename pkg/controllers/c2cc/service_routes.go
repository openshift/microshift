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
	c2ccSvcRouteTable   = 201
	c2ccSvcRouteProto   = 201
	c2ccSvcRulePriority = 99
	mgmtPortInterface   = "ovn-k8s-mp0"
)

type serviceRouteManager struct {
	remoteCIDRs     []*net.IPNet
	localSvcCIDRs   []*net.IPNet
}

func newServiceRouteManager(cfg *config.Config) *serviceRouteManager {
	var remoteCIDRs []*net.IPNet
	for _, rc := range cfg.C2CC.Resolved {
		remoteCIDRs = append(remoteCIDRs, rc.ClusterNetwork...)
		remoteCIDRs = append(remoteCIDRs, rc.ServiceNetwork...)
	}

	var localSvcCIDRs []*net.IPNet
	for _, s := range cfg.Network.ServiceNetwork {
		_, ipNet, err := net.ParseCIDR(s)
		if err == nil {
			localSvcCIDRs = append(localSvcCIDRs, ipNet)
		}
	}

	return &serviceRouteManager{
		remoteCIDRs:   remoteCIDRs,
		localSvcCIDRs: localSvcCIDRs,
	}
}

func (m *serviceRouteManager) reconcile(ctx context.Context) error {
	gwIP, linkIdx, err := getMgmtPortGateway()
	if err != nil {
		klog.V(4).Infof("Management port not ready: %v, will retry", err)
		return nil
	}

	if err := m.reconcileRoutes(gwIP, linkIdx); err != nil {
		return fmt.Errorf("service routes: %w", err)
	}
	if err := m.reconcileRules(); err != nil {
		return fmt.Errorf("service rules: %w", err)
	}
	return nil
}

func (m *serviceRouteManager) reconcileRoutes(gwIP net.IP, linkIdx int) error {
	var desired []netlink.Route
	for _, svcCIDR := range m.localSvcCIDRs {
		desired = append(desired, netlink.Route{
			Dst:       svcCIDR,
			Gw:        gwIP,
			Table:     c2ccSvcRouteTable,
			Protocol:  c2ccSvcRouteProto,
			LinkIndex: linkIdx,
		})
	}

	actual, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{
		Table:    c2ccSvcRouteTable,
		Protocol: c2ccSvcRouteProto,
	}, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_PROTOCOL)
	if err != nil {
		return fmt.Errorf("listing table %d routes: %w", c2ccSvcRouteTable, err)
	}

	actualByDst := make(map[string]netlink.Route, len(actual))
	for _, r := range actual {
		if r.Dst != nil {
			actualByDst[r.Dst.String()] = r
		}
	}

	desiredByDst := make(map[string]bool, len(desired))
	for _, r := range desired {
		desiredByDst[r.Dst.String()] = true
	}

	for _, r := range desired {
		if _, exists := actualByDst[r.Dst.String()]; exists {
			continue
		}
		route := r
		if err := netlink.RouteReplace(&route); err != nil {
			klog.Errorf("Failed to add service route to %s: %v", route.Dst, err)
			continue
		}
		klog.V(2).Infof("Service route add: %s via %s table %d", route.Dst, route.Gw, c2ccSvcRouteTable)
	}

	for dst, r := range actualByDst {
		if desiredByDst[dst] {
			continue
		}
		route := r
		if err := netlink.RouteDel(&route); err != nil {
			klog.Errorf("Failed to delete stale service route %s: %v", dst, err)
		}
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
	for _, remoteCIDR := range m.remoteCIDRs {
		for _, svcCIDR := range m.localSvcCIDRs {
			if ipFamilyOf(remoteCIDR) != ipFamilyOf(svcCIDR) {
				continue
			}
			rule := netlink.NewRule()
			rule.Src = remoteCIDR
			rule.Dst = svcCIDR
			rule.Table = c2ccSvcRouteTable
			rule.Priority = c2ccSvcRulePriority
			rule.Family = ipFamilyOf(remoteCIDR)
			desired = append(desired, *rule)
			desiredKeys[ruleKey{src: remoteCIDR.String(), dst: svcCIDR.String()}] = true
		}
	}

	allRules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("listing ip rules: %w", err)
	}

	actualKeys := make(map[ruleKey]netlink.Rule)
	for _, r := range allRules {
		if r.Priority == c2ccSvcRulePriority && r.Table == c2ccSvcRouteTable && r.Src != nil && r.Dst != nil {
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
			}
			continue
		}
		klog.V(2).Infof("Service rule add: from %s to %s lookup %d", rule.Src, rule.Dst, c2ccSvcRouteTable)
	}

	for k, r := range actualKeys {
		if desiredKeys[k] {
			continue
		}
		rule := r
		if err := netlink.RuleDel(&rule); err != nil {
			klog.Errorf("Failed to delete stale service rule from %s to %s: %v", k.src, k.dst, err)
		}
	}

	return nil
}

func (m *serviceRouteManager) cleanup(ctx context.Context) error {
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{
		Table:    c2ccSvcRouteTable,
		Protocol: c2ccSvcRouteProto,
	}, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_PROTOCOL)
	if err == nil {
		for _, r := range routes {
			route := r
			if err := netlink.RouteDel(&route); err != nil {
				klog.Errorf("Failed to cleanup service route %v: %v", r.Dst, err)
			}
		}
	}

	allRules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err == nil {
		for _, r := range allRules {
			if r.Priority == c2ccSvcRulePriority && r.Table == c2ccSvcRouteTable {
				rule := r
				if err := netlink.RuleDel(&rule); err != nil {
					klog.Errorf("Failed to cleanup service rule: %v", err)
				}
			}
		}
	}

	return nil
}

func (m *serviceRouteManager) subscribe(reconcileCh chan<- string) (chan struct{}, error) {
	routeUpdates := make(chan netlink.RouteUpdate, 100)
	done := make(chan struct{})

	if err := netlink.RouteSubscribe(routeUpdates, done); err != nil {
		return nil, fmt.Errorf("subscribe to route events: %w", err)
	}

	go func() {
		for update := range routeUpdates {
			if update.Table != c2ccSvcRouteTable {
				continue
			}
			select {
			case reconcileCh <- "service-route-change":
			default:
			}
		}
	}()

	klog.V(2).Infof("Subscribed to netlink route events for table %d", c2ccSvcRouteTable)
	return done, nil
}

func getMgmtPortGateway() (net.IP, int, error) {
	link, err := netlink.LinkByName(mgmtPortInterface)
	if err != nil {
		return nil, 0, fmt.Errorf("get %s: %w", mgmtPortInterface, err)
	}

	addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return nil, 0, fmt.Errorf("list addresses on %s: %w", mgmtPortInterface, err)
	}

	for _, addr := range addrs {
		if addr.IP.To4() != nil {
			gwIP := make(net.IP, len(addr.IP.To4()))
			copy(gwIP, addr.IP.To4())
			gwIP = gwIP.Mask(addr.Mask)
			gwIP[len(gwIP)-1] = 1
			return gwIP, link.Attrs().Index, nil
		}
	}

	return nil, 0, fmt.Errorf("no IPv4 address found on %s", mgmtPortInterface)
}
