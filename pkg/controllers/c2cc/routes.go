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
	c2ccRulePriority = 100
)

type routeTarget struct {
	dst *net.IPNet
	key string // dst.String(), pre-computed
	gw  net.IP
}

type linuxRouteManager struct {
	policyRouteTable

	desired []routeTarget
}

func newLinuxRouteManager(cfg *config.Config) *linuxRouteManager {
	m := &linuxRouteManager{
		policyRouteTable: policyRouteTable{
			table:    cfg.C2CC.ResolvedRouteTableID,
			proto:    cfg.C2CC.ResolvedRouteTableID,
			priority: c2ccRulePriority,
		},
	}

	for i := range cfg.C2CC.Resolved {
		rc := &cfg.C2CC.Resolved[i]
		for _, cidr := range rc.AllCIDRs() {
			gw, ok := rc.NextHopForFamily(ipFamilyOf(cidr))
			if !ok {
				continue
			}
			m.desired = append(m.desired, routeTarget{
				dst: cidr,
				key: cidr.String(),
				gw:  gw,
			})
		}
	}

	return m
}

func (m *linuxRouteManager) reconcile(ctx context.Context) error {
	desired := make([]netlink.Route, 0, len(m.desired))
	for _, rt := range m.desired {
		desired = append(desired, netlink.Route{
			Dst:      rt.dst,
			Gw:       rt.gw,
			Table:    m.table,
			Protocol: netlink.RouteProtocol(m.proto),
		})
	}

	if err := m.reconcileRoutes(desired); err != nil {
		return fmt.Errorf("failed to reconcile linux routes: %w", err)
	}
	if err := m.reconcileRules(); err != nil {
		return fmt.Errorf("failed to reconcile ip rules: %w", err)
	}
	return nil
}

func (m *linuxRouteManager) reconcileRules() error {
	allRules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("failed to list ip rules: %w", err)
	}

	actualByDst := make(map[string]netlink.Rule)
	for _, r := range allRules {
		if r.Priority == m.priority && r.Table == m.table && r.Dst != nil {
			actualByDst[r.Dst.String()] = r
		}
	}

	var errs []error
	for _, rt := range m.desired {
		if _, exists := actualByDst[rt.key]; exists {
			delete(actualByDst, rt.key)
			continue
		}
		rule := netlink.NewRule()
		rule.Dst = rt.dst
		rule.Table = m.table
		rule.Priority = m.priority
		rule.Family = ipFamilyOf(rt.dst)
		if err := netlink.RuleAdd(rule); err != nil {
			if !errors.Is(err, syscall.EEXIST) {
				klog.Errorf("Failed to add ip rule for %s: %v", rt.key, err)
				errs = append(errs, fmt.Errorf("failed to add rule %s: %w", rt.key, err))
			}
			continue
		}
		klog.V(2).Infof("IP rule add: to %s lookup %d priority %d", rt.key, m.table, m.priority)
	}

	for dst, r := range actualByDst {
		rule := r
		if err := netlink.RuleDel(&rule); err != nil {
			klog.Errorf("Failed to delete stale ip rule for %s: %v", dst, err)
			errs = append(errs, fmt.Errorf("failed to delete rule %s: %w", dst, err))
			continue
		}
		klog.V(2).Infof("IP rule del: to %s (stale)", dst)
	}

	return errors.Join(errs...)
}

func (m *linuxRouteManager) cleanup(ctx context.Context) error {
	return errors.Join(m.cleanupRoutes(), m.cleanupRules())
}
