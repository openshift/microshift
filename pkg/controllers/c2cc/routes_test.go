package c2cc

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLinuxRouteManager_DesiredState(t *testing.T) {
	cfg := testConfigWithRemotes(t,
		testRemote("192.168.1.10", []string{"10.45.0.0/16"}, []string{"10.46.0.0/16"}),
	)

	mgr := newLinuxRouteManager(cfg)

	require.Len(t, mgr.desiredDsts, 2)
	assert.Equal(t, "10.45.0.0/16", mgr.desiredDsts[0].String())
	assert.Equal(t, "10.46.0.0/16", mgr.desiredDsts[1].String())
	assert.Equal(t, net.ParseIP("192.168.1.10").To4(), mgr.desiredGWs["10.45.0.0/16"].To4())
	assert.Equal(t, net.ParseIP("192.168.1.10").To4(), mgr.desiredGWs["10.46.0.0/16"].To4())
}

func TestNewLinuxRouteManager_MultipleRemotes(t *testing.T) {
	cfg := testConfigWithRemotes(t,
		testRemote("192.168.1.10", []string{"10.45.0.0/16"}, []string{"10.46.0.0/16"}),
		testRemote("192.168.1.20", []string{"10.55.0.0/16"}, []string{"10.56.0.0/16"}),
	)

	mgr := newLinuxRouteManager(cfg)

	require.Len(t, mgr.desiredDsts, 4)
	assert.Equal(t, net.ParseIP("192.168.1.10").To4(), mgr.desiredGWs["10.45.0.0/16"].To4())
	assert.Equal(t, net.ParseIP("192.168.1.20").To4(), mgr.desiredGWs["10.55.0.0/16"].To4())
}

func TestNewLinuxRouteManager_EmptyConfig(t *testing.T) {
	cfg := testConfigWithRemotes(t)
	mgr := newLinuxRouteManager(cfg)
	assert.Empty(t, mgr.desiredDsts)
	assert.Empty(t, mgr.desiredGWs)
}

func TestNewLinuxRouteManager_DualStack(t *testing.T) {
	cfg := testConfigWithRemotes(t,
		testDualStackRemote(
			[]string{"192.168.1.10", "fd01::10"},
			[]string{"10.45.0.0/16", "fd02::/64"},
			[]string{"10.46.0.0/16", "fd03::/112"},
		),
	)

	mgr := newLinuxRouteManager(cfg)

	require.Len(t, mgr.desiredDsts, 4, "should have 2 IPv4 + 2 IPv6 routes")

	// Build map for easier assertion
	gwByDst := make(map[string]string)
	for i, dst := range mgr.desiredDsts {
		gwByDst[dst.String()] = mgr.desiredGWs[mgr.desiredDstKeys[i]].String()
	}

	// IPv4 routes use IPv4 gateway
	assert.Equal(t, "192.168.1.10", gwByDst["10.45.0.0/16"])
	assert.Equal(t, "192.168.1.10", gwByDst["10.46.0.0/16"])

	// IPv6 routes use IPv6 gateway
	assert.Equal(t, "fd01::10", gwByDst["fd02::/64"])
	assert.Equal(t, "fd01::10", gwByDst["fd03::/112"])
}

func TestNewLinuxRouteManager_DualStackMultipleRemotes(t *testing.T) {
	cfg := testConfigWithRemotes(t,
		testDualStackRemote(
			[]string{"192.168.1.10", "fd01::10"},
			[]string{"10.45.0.0/16", "fd02::/64"},
			[]string{"10.46.0.0/16", "fd03::/112"},
		),
		testDualStackRemote(
			[]string{"192.168.1.20", "fd01::20"},
			[]string{"10.55.0.0/16", "fd12::/64"},
			[]string{"10.56.0.0/16", "fd13::/112"},
		),
	)

	mgr := newLinuxRouteManager(cfg)

	require.Len(t, mgr.desiredDsts, 8, "should have 4 IPv4 + 4 IPv6 routes")

	gwByDst := make(map[string]string)
	for i, dst := range mgr.desiredDsts {
		gwByDst[dst.String()] = mgr.desiredGWs[mgr.desiredDstKeys[i]].String()
	}

	// Remote cluster 1 - IPv4
	assert.Equal(t, "192.168.1.10", gwByDst["10.45.0.0/16"])
	assert.Equal(t, "192.168.1.10", gwByDst["10.46.0.0/16"])
	// Remote cluster 1 - IPv6
	assert.Equal(t, "fd01::10", gwByDst["fd02::/64"])
	assert.Equal(t, "fd01::10", gwByDst["fd03::/112"])

	// Remote cluster 2 - IPv4
	assert.Equal(t, "192.168.1.20", gwByDst["10.55.0.0/16"])
	assert.Equal(t, "192.168.1.20", gwByDst["10.56.0.0/16"])
	// Remote cluster 2 - IPv6
	assert.Equal(t, "fd01::20", gwByDst["fd12::/64"])
	assert.Equal(t, "fd01::20", gwByDst["fd13::/112"])
}

func TestNewLinuxRouteManager_DualStackMissingGateway(t *testing.T) {
	// Scenario: IPv6 CIDRs provided but no IPv6 gateway
	cfg := testConfigWithRemotes(t,
		testRemote("192.168.1.10", // only IPv4 gateway
			[]string{"10.45.0.0/16", "fd02::/64"},
			[]string{"10.46.0.0/16", "fd03::/112"}),
	)

	mgr := newLinuxRouteManager(cfg)

	require.Len(t, mgr.desiredDsts, 2, "should only have IPv4 routes (IPv6 skipped due to missing gateway)")

	for i, dst := range mgr.desiredDsts {
		// All routes should be IPv4
		assert.Contains(t, []string{"10.45.0.0/16", "10.46.0.0/16"}, dst.String())
		assert.Equal(t, "192.168.1.10", mgr.desiredGWs[mgr.desiredDstKeys[i]].String())
	}
}
