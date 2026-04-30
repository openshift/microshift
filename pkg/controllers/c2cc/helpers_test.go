package c2cc

import (
	"net"
	"testing"

	"github.com/openshift/microshift/pkg/config"
	"github.com/stretchr/testify/require"
)

type testRemoteConfig struct {
	nextHop        string
	clusterNetwork []string
	serviceNetwork []string
}

func testRemote(nextHop string, clusterNetwork, serviceNetwork []string) testRemoteConfig {
	return testRemoteConfig{
		nextHop:        nextHop,
		clusterNetwork: clusterNetwork,
		serviceNetwork: serviceNetwork,
	}
}

func testConfigWithRemotes(t *testing.T, remotes ...testRemoteConfig) *config.Config {
	t.Helper()

	cfg := &config.Config{}
	cfg.Node.NodeIP = "192.168.1.1"
	cfg.Network.ServiceNetwork = []string{"10.43.0.0/16"}

	for _, r := range remotes {
		resolved := config.ResolvedRemoteCluster{
			NextHop: net.ParseIP(r.nextHop),
		}
		require.NotNil(t, resolved.NextHop, "invalid nextHop: %s", r.nextHop)

		for _, cn := range r.clusterNetwork {
			_, ipNet, err := net.ParseCIDR(cn)
			require.NoError(t, err)
			resolved.ClusterNetwork = append(resolved.ClusterNetwork, ipNet)
		}
		for _, sn := range r.serviceNetwork {
			_, ipNet, err := net.ParseCIDR(sn)
			require.NoError(t, err)
			resolved.ServiceNetwork = append(resolved.ServiceNetwork, ipNet)
		}
		cfg.C2CC.Resolved = append(cfg.C2CC.Resolved, resolved)
	}

	return cfg
}
