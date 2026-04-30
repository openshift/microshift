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
	policyRouteTable
	desiredDsts []*net.IPNet
	desiredGWs  map[string]net.IP
}

func newLinuxRouteManager(cfg *config.Config) *linuxRouteManager {
	m := &linuxRouteManager{
		policyRouteTable: policyRouteTable{
			table:    c2ccRouteTable,
			proto:    c2ccRouteProto,
			priority: c2ccRulePriority,
		},
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
	var desired []netlink.Route
	for _, cidr := range m.desiredDsts {
		desired = append(desired, netlink.Route{
			Dst:      cidr,
			Gw:       m.desiredGWs[cidr.String()],
			Table:    m.table,
			Protocol: netlink.RouteProtocol(m.proto),
		})
	}

	if err := m.reconcileRoutes(desired); err != nil {
		return fmt.Errorf("linux routes: %w", err)
	}
	if err := m.reconcileRules(); err != nil {
		return fmt.Errorf("ip rules: %w", err)
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
		if r.Priority == m.priority && r.Table == m.table && r.Dst != nil {
			actualByDst[r.Dst.String()] = r
		}
	}

	var errs []error
	for _, cidr := range m.desiredDsts {
		dst := cidr.String()
		if _, exists := actualByDst[dst]; exists {
			delete(actualByDst, dst)
			continue
		}
		rule := netlink.NewRule()
		rule.Dst = cidr
		rule.Table = m.table
		rule.Priority = m.priority
		rule.Family = ipFamilyOf(cidr)
		if err := netlink.RuleAdd(rule); err != nil {
			if !errors.Is(err, syscall.EEXIST) {
				klog.Errorf("Failed to add ip rule for %s: %v", dst, err)
				errs = append(errs, fmt.Errorf("add rule %s: %w", dst, err))
			}
			continue
		}
		klog.V(2).Infof("IP rule add: to %s lookup %d priority %d", dst, m.table, m.priority)
	}

	for dst, r := range actualByDst {
		rule := r
		if err := netlink.RuleDel(&rule); err != nil {
			klog.Errorf("Failed to delete stale ip rule for %s: %v", dst, err)
			errs = append(errs, fmt.Errorf("delete rule %s: %w", dst, err))
			continue
		}
		klog.V(2).Infof("IP rule del: to %s (stale)", dst)
	}

	return errors.Join(errs...)
}

func (m *linuxRouteManager) cleanup(ctx context.Context) error {
	_ = m.cleanupRoutes()
	_ = m.cleanupRules()
	return nil
}

