package c2cc

import (
	"net"
	"testing"

	"github.com/openshift/microshift/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOVNRouteManager_DesiredRoutes(t *testing.T) {
	resolved := []config.ResolvedRemoteCluster{
		{
			NextHops:       map[int]net.IP{2: net.ParseIP("192.168.1.10")},
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.45.0.0/16")},
			ServiceNetwork: []*net.IPNet{parseCIDR(t, "10.46.0.0/16")},
		},
	}

	mgr := newOVNRouteManager(nil, "test-node", resolved)

	assert.Equal(t, "GR_test-node", mgr.gwRouter)
	require.Len(t, mgr.desired, 2)

	assert.Equal(t, "10.45.0.0/16", mgr.desired[0].IPPrefix)
	assert.Equal(t, "192.168.1.10", mgr.desired[0].Nexthop)
	assert.Equal(t, c2ccOwnerController, mgr.desired[0].ExternalIDs[ownerControllerKey])

	assert.Equal(t, "10.46.0.0/16", mgr.desired[1].IPPrefix)
	assert.Equal(t, "192.168.1.10", mgr.desired[1].Nexthop)
}

func TestNewOVNRouteManager_MultipleRemotes(t *testing.T) {
	resolved := []config.ResolvedRemoteCluster{
		{
			NextHops:       map[int]net.IP{2: net.ParseIP("192.168.1.10")},
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.45.0.0/16")},
			ServiceNetwork: []*net.IPNet{parseCIDR(t, "10.46.0.0/16")},
		},
		{
			NextHops:       map[int]net.IP{2: net.ParseIP("192.168.1.20")},
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.55.0.0/16")},
			ServiceNetwork: []*net.IPNet{parseCIDR(t, "10.56.0.0/16")},
		},
	}

	mgr := newOVNRouteManager(nil, "node-a", resolved)

	assert.Equal(t, "GR_node-a", mgr.gwRouter)
	require.Len(t, mgr.desired, 4)

	nexthops := make(map[string]string)
	for _, d := range mgr.desired {
		nexthops[d.IPPrefix] = d.Nexthop
	}
	assert.Equal(t, "192.168.1.10", nexthops["10.45.0.0/16"])
	assert.Equal(t, "192.168.1.10", nexthops["10.46.0.0/16"])
	assert.Equal(t, "192.168.1.20", nexthops["10.55.0.0/16"])
	assert.Equal(t, "192.168.1.20", nexthops["10.56.0.0/16"])
}

func TestNewOVNRouteManager_EmptyResolved(t *testing.T) {
	mgr := newOVNRouteManager(nil, "test-node", nil)
	assert.Empty(t, mgr.desired)
	assert.Equal(t, "GR_test-node", mgr.gwRouter)
}

func TestOVNRouteManager_OwnerTag(t *testing.T) {
	resolved := []config.ResolvedRemoteCluster{
		{
			NextHops:       map[int]net.IP{2: net.ParseIP("192.168.1.10")},
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.45.0.0/16")},
		},
	}
	mgr := newOVNRouteManager(nil, "test-node", resolved)

	for _, d := range mgr.desired {
		assert.Equal(t, "microshift-c2cc", d.ExternalIDs[ownerControllerKey],
			"all routes must have the owner controller tag")
	}
}

func TestNewOVNRouteManager_DualStack(t *testing.T) {
	resolved := []config.ResolvedRemoteCluster{
		{
			NextHops: map[int]net.IP{
				2:  net.ParseIP("192.168.1.10"),
				10: net.ParseIP("fd01::10"),
			},
			ClusterNetwork: []*net.IPNet{
				parseCIDR(t, "10.45.0.0/16"),
				parseCIDR(t, "fd02::/64"),
			},
			ServiceNetwork: []*net.IPNet{
				parseCIDR(t, "10.46.0.0/16"),
				parseCIDR(t, "fd03::/112"),
			},
		},
	}

	mgr := newOVNRouteManager(nil, "test-node", resolved)

	assert.Equal(t, "GR_test-node", mgr.gwRouter)
	require.Len(t, mgr.desired, 4, "should have 2 IPv4 + 2 IPv6 routes")

	// Build map for easier assertion
	routesByPrefix := make(map[string]string)
	for _, d := range mgr.desired {
		routesByPrefix[d.IPPrefix] = d.Nexthop
		assert.Equal(t, c2ccOwnerController, d.ExternalIDs[ownerControllerKey])
	}

	// Verify IPv4 routes use IPv4 gateway
	assert.Equal(t, "192.168.1.10", routesByPrefix["10.45.0.0/16"])
	assert.Equal(t, "192.168.1.10", routesByPrefix["10.46.0.0/16"])

	// Verify IPv6 routes use IPv6 gateway
	assert.Equal(t, "fd01::10", routesByPrefix["fd02::/64"])
	assert.Equal(t, "fd01::10", routesByPrefix["fd03::/112"])
}

func TestNewOVNRouteManager_DualStackMultipleRemotes(t *testing.T) {
	resolved := []config.ResolvedRemoteCluster{
		{
			NextHops: map[int]net.IP{
				2:  net.ParseIP("192.168.1.10"),
				10: net.ParseIP("fd01::10"),
			},
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.45.0.0/16"), parseCIDR(t, "fd02::/64")},
			ServiceNetwork: []*net.IPNet{parseCIDR(t, "10.46.0.0/16"), parseCIDR(t, "fd03::/112")},
		},
		{
			NextHops: map[int]net.IP{
				2:  net.ParseIP("192.168.1.20"),
				10: net.ParseIP("fd01::20"),
			},
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.55.0.0/16"), parseCIDR(t, "fd12::/64")},
			ServiceNetwork: []*net.IPNet{parseCIDR(t, "10.56.0.0/16"), parseCIDR(t, "fd13::/112")},
		},
	}

	mgr := newOVNRouteManager(nil, "node-a", resolved)

	assert.Equal(t, "GR_node-a", mgr.gwRouter)
	require.Len(t, mgr.desired, 8, "should have 4 IPv4 + 4 IPv6 routes")

	nexthops := make(map[string]string)
	for _, d := range mgr.desired {
		nexthops[d.IPPrefix] = d.Nexthop
	}

	// Remote cluster 1 - IPv4
	assert.Equal(t, "192.168.1.10", nexthops["10.45.0.0/16"])
	assert.Equal(t, "192.168.1.10", nexthops["10.46.0.0/16"])
	// Remote cluster 1 - IPv6
	assert.Equal(t, "fd01::10", nexthops["fd02::/64"])
	assert.Equal(t, "fd01::10", nexthops["fd03::/112"])

	// Remote cluster 2 - IPv4
	assert.Equal(t, "192.168.1.20", nexthops["10.55.0.0/16"])
	assert.Equal(t, "192.168.1.20", nexthops["10.56.0.0/16"])
	// Remote cluster 2 - IPv6
	assert.Equal(t, "fd01::20", nexthops["fd12::/64"])
	assert.Equal(t, "fd01::20", nexthops["fd13::/112"])
}

func TestNewOVNRouteManager_DualStackMissingGateway(t *testing.T) {
	// Scenario: IPv6 CIDRs provided but no IPv6 gateway
	resolved := []config.ResolvedRemoteCluster{
		{
			NextHops:       map[int]net.IP{2: net.ParseIP("192.168.1.10")}, // only IPv4 gateway
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.45.0.0/16"), parseCIDR(t, "fd02::/64")},
			ServiceNetwork: []*net.IPNet{parseCIDR(t, "10.46.0.0/16"), parseCIDR(t, "fd03::/112")},
		},
	}

	mgr := newOVNRouteManager(nil, "test-node", resolved)

	require.Len(t, mgr.desired, 2, "should only have IPv4 routes (IPv6 skipped due to missing gateway)")

	for _, d := range mgr.desired {
		assert.Contains(t, []string{"10.45.0.0/16", "10.46.0.0/16"}, d.IPPrefix)
		assert.Equal(t, "192.168.1.10", d.Nexthop)
	}
}
