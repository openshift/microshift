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
