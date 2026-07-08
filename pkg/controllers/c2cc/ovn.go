package c2cc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/ovn-kubernetes/libovsdb/cache"
	"github.com/ovn-kubernetes/libovsdb/client"
	"github.com/ovn-kubernetes/libovsdb/model"
	"github.com/ovn-kubernetes/libovsdb/ovsdb"
	"k8s.io/klog/v2"
)

const (
	c2ccOwnerController = "microshift-c2cc"
	ownerControllerKey  = "k8s.ovn.org/owner-controller"
	gwRouterPrefix      = "GR_"
)

func buildNamedUUID(prefix, suffix string) string {
	r := strings.NewReplacer(".", "_", "/", "_", ":", "_", "-", "_")
	return r.Replace(prefix + suffix)
}

type routeKey struct {
	prefix  string
	nexthop string
}

type ovnRouteManager struct {
	nbClient client.Client
	gwRouter string
	desired  []LogicalRouterStaticRoute
}

func newOVNRouteManager(nbClient client.Client, nodeName string, resolved []config.ResolvedRemoteCluster) *ovnRouteManager {
	gwRouter := gwRouterPrefix + nodeName

	var desired []LogicalRouterStaticRoute
	for i := range resolved {
		rc := &resolved[i]
		for _, cidr := range rc.AllCIDRs() {
			gw, ok := rc.NextHopForFamily(ipFamilyOf(cidr))
			if !ok {
				continue
			}
			desired = append(desired, LogicalRouterStaticRoute{
				IPPrefix:    cidr.String(),
				Nexthop:     gw.String(),
				ExternalIDs: map[string]string{ownerControllerKey: c2ccOwnerController},
			})
		}
	}

	return &ovnRouteManager{
		nbClient: nbClient,
		gwRouter: gwRouter,
		desired:  desired,
	}
}

func (m *ovnRouteManager) reconcile(ctx context.Context) error {
	actual, err := m.listC2CCRoutes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list OVN routes: %w", err)
	}

	actualByKey := make(map[routeKey]*LogicalRouterStaticRoute, len(actual))
	for i := range actual {
		k := routeKey{prefix: actual[i].IPPrefix, nexthop: actual[i].Nexthop}
		actualByKey[k] = &actual[i]
	}

	desiredKeys := make(map[routeKey]bool, len(m.desired))
	for _, d := range m.desired {
		desiredKeys[routeKey{prefix: d.IPPrefix, nexthop: d.Nexthop}] = true
	}

	var ops []ovsdb.Operation

	for _, d := range m.desired {
		k := routeKey{prefix: d.IPPrefix, nexthop: d.Nexthop}
		if _, exists := actualByKey[k]; exists {
			continue
		}

		route := d
		route.UUID = buildNamedUUID("c2cc_route_", route.IPPrefix)
		createOps, err := m.nbClient.Create(&route)
		if err != nil {
			return fmt.Errorf("failed to create route %s via %s: %w", route.IPPrefix, route.Nexthop, err)
		}
		ops = append(ops, createOps...)

		router := &LogicalRouter{Name: m.gwRouter}
		mutateOps, err := m.nbClient.Where(router).Mutate(router, model.Mutation{
			Field:   &router.StaticRoutes,
			Mutator: ovsdb.MutateOperationInsert,
			Value:   []string{route.UUID},
		})
		if err != nil {
			return fmt.Errorf("failed to mutate router for route %s: %w", route.IPPrefix, err)
		}
		ops = append(ops, mutateOps...)

		klog.V(2).Infof("OVN route add: %s via %s on %s", route.IPPrefix, route.Nexthop, m.gwRouter)
	}

	for k, existing := range actualByKey {
		if desiredKeys[k] {
			continue
		}

		router := &LogicalRouter{Name: m.gwRouter}
		mutateOps, err := m.nbClient.Where(router).Mutate(router, model.Mutation{
			Field:   &router.StaticRoutes,
			Mutator: ovsdb.MutateOperationDelete,
			Value:   []string{existing.UUID},
		})
		if err != nil {
			return fmt.Errorf("failed to mutate router to remove route %s: %w", existing.UUID, err)
		}
		ops = append(ops, mutateOps...)

		deleteOps, err := m.nbClient.Where(existing).Delete()
		if err != nil {
			return fmt.Errorf("failed to delete route %s: %w", existing.UUID, err)
		}
		ops = append(ops, deleteOps...)

		klog.V(2).Infof("OVN route remove: %s via %s from %s", existing.IPPrefix, existing.Nexthop, m.gwRouter)
	}

	if len(ops) == 0 {
		return nil
	}

	results, err := m.nbClient.Transact(ctx, ops...)
	if err != nil {
		return fmt.Errorf("failed to transact OVN: %w", err)
	}
	for _, r := range results {
		if r.Error != "" {
			return fmt.Errorf("failed to transact OVN: %s (%s)", r.Error, r.Details)
		}
	}

	return nil
}

func (m *ovnRouteManager) cleanup(ctx context.Context) error {
	routes, err := m.listC2CCRoutes(ctx)
	if err != nil {
		return err
	}

	var errs []error
	for _, r := range routes {
		route := r
		router := &LogicalRouter{Name: m.gwRouter}

		ops := make([]ovsdb.Operation, 0, 2)
		mutateOps, err := m.nbClient.Where(router).Mutate(router, model.Mutation{
			Field:   &router.StaticRoutes,
			Mutator: ovsdb.MutateOperationDelete,
			Value:   []string{route.UUID},
		})
		if err != nil {
			klog.Errorf("Failed to build mutate for route %s: %v", route.UUID, err)
			errs = append(errs, err)
			continue
		}
		ops = append(ops, mutateOps...)

		delOps, err := m.nbClient.Where(&route).Delete()
		if err != nil {
			klog.Errorf("Failed to build delete for route %s: %v", route.UUID, err)
			errs = append(errs, err)
			continue
		}
		ops = append(ops, delOps...)

		results, err := m.nbClient.Transact(ctx, ops...)
		if err != nil {
			klog.Errorf("Failed to remove OVN route %s: %v", route.UUID, err)
			errs = append(errs, err)
			continue
		}
		for _, res := range results {
			if res.Error != "" {
				opErr := fmt.Errorf("failed to remove OVN route %s: %s (%s)", route.UUID, res.Error, res.Details)
				klog.Errorf("%v", opErr)
				errs = append(errs, opErr)
			}
		}
	}
	return errors.Join(errs...)
}

func (m *ovnRouteManager) subscribe(ctx context.Context, reconcileCh chan<- string) {
	m.nbClient.Cache().AddEventHandler(&cache.EventHandlerFuncs{
		UpdateFunc: func(table string, old, updated model.Model) {
			if table != "Logical_Router_Static_Route" {
				return
			}
			if route, ok := updated.(*LogicalRouterStaticRoute); ok && route.ExternalIDs[ownerControllerKey] == c2ccOwnerController {
				select {
				case reconcileCh <- "ovn-route-updated":
				default:
				}
			}
		},
		DeleteFunc: func(table string, old model.Model) {
			if table != "Logical_Router_Static_Route" {
				return
			}
			if route, ok := old.(*LogicalRouterStaticRoute); ok && route.ExternalIDs[ownerControllerKey] == c2ccOwnerController {
				select {
				case reconcileCh <- "ovn-route-deleted":
				default:
				}
			}
		},
	})
	klog.V(2).Infof("Subscribed to OVN NB cache events for C2CC routes")

	go func() {
		for {
			disconnectCh := m.nbClient.DisconnectNotify()
			select {
			case <-ctx.Done():
				return
			case <-disconnectCh:
				klog.Warningf("OVN NB connection lost, waiting for reconnect...")
				for !m.nbClient.Connected() {
					select {
					case <-ctx.Done():
						return
					case <-time.After(1 * time.Second):
					}
				}
				klog.Infof("OVN NB reconnected, triggering full reconcile")
				select {
				case reconcileCh <- "ovn-reconnected":
				default:
				}
			}
		}
	}()
	klog.V(2).Infof("Subscribed to OVN NB disconnect notifications")
}

func (m *ovnRouteManager) listC2CCRoutes(ctx context.Context) ([]LogicalRouterStaticRoute, error) {
	var routes []LogicalRouterStaticRoute
	err := m.nbClient.WhereCache(func(r *LogicalRouterStaticRoute) bool {
		return r.ExternalIDs[ownerControllerKey] == c2ccOwnerController
	}).List(ctx, &routes)
	if err != nil {
		return nil, err
	}
	return routes, nil
}
