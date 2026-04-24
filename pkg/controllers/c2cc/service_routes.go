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
	policyRouteTable
	remoteCIDRs   []*net.IPNet
	localSvcCIDRs []*net.IPNet
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
		policyRouteTable: policyRouteTable{
			table:    c2ccSvcRouteTable,
			proto:    c2ccSvcRouteProto,
			priority: c2ccSvcRulePriority,
		},
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

	var desired []netlink.Route
	for _, svcCIDR := range m.localSvcCIDRs {
		desired = append(desired, netlink.Route{
			Dst:       svcCIDR,
			Gw:        gwIP,
			Table:     m.table,
			Protocol:  netlink.RouteProtocol(m.proto),
			LinkIndex: linkIdx,
		})
	}

	if err := m.reconcileRoutes(desired); err != nil {
		return fmt.Errorf("service routes: %w", err)
	}
	if err := m.reconcileRules(); err != nil {
		return fmt.Errorf("service rules: %w", err)
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
			rule.Table = m.table
			rule.Priority = m.priority
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
		}
	}

	return nil
}

func (m *serviceRouteManager) cleanup(ctx context.Context) error {
	_ = m.cleanupRoutes()
	_ = m.cleanupRules()
	return nil
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
