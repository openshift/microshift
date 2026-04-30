package c2cc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceRouteManager_DesiredState(t *testing.T) {
	cfg := testConfigWithRemotes(t,
		testRemote("192.168.1.10", []string{"10.45.0.0/16"}, []string{"10.46.0.0/16"}),
	)

	mgr := newServiceRouteManager(cfg)

	require.Len(t, mgr.remoteCIDRs, 2)
	assert.Equal(t, "10.45.0.0/16", mgr.remoteCIDRs[0].String())
	assert.Equal(t, "10.46.0.0/16", mgr.remoteCIDRs[1].String())

	require.Len(t, mgr.localSvcCIDRs, 1)
	assert.Equal(t, "10.43.0.0/16", mgr.localSvcCIDRs[0].String())
}

func TestNewServiceRouteManager_MultipleRemotes(t *testing.T) {
	cfg := testConfigWithRemotes(t,
		testRemote("192.168.1.10", []string{"10.45.0.0/16"}, []string{"10.46.0.0/16"}),
		testRemote("192.168.1.20", []string{"10.55.0.0/16"}, []string{"10.56.0.0/16"}),
	)

	mgr := newServiceRouteManager(cfg)
	assert.Len(t, mgr.remoteCIDRs, 4)
	assert.Len(t, mgr.localSvcCIDRs, 1)
}

func TestNewServiceRouteManager_EmptyConfig(t *testing.T) {
	cfg := testConfigWithRemotes(t)
	mgr := newServiceRouteManager(cfg)
	assert.Empty(t, mgr.remoteCIDRs)
}
