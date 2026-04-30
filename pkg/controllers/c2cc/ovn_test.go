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
			NextHop:        net.ParseIP("192.168.1.10"),
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
			NextHop:        net.ParseIP("192.168.1.10"),
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.45.0.0/16")},
			ServiceNetwork: []*net.IPNet{parseCIDR(t, "10.46.0.0/16")},
		},
		{
			NextHop:        net.ParseIP("192.168.1.20"),
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
			NextHop:        net.ParseIP("192.168.1.10"),
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.45.0.0/16")},
		},
	}
	mgr := newOVNRouteManager(nil, "test-node", resolved)

	for _, d := range mgr.desired {
		assert.Equal(t, "microshift-c2cc", d.ExternalIDs[ownerControllerKey],
			"all routes must have the owner controller tag")
	}
}
