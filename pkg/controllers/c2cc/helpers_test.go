package c2cc

import (
	"net"
	"testing"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/openshift/microshift/pkg/config"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
)

type testRemoteConfig struct {
	nextHops       []string
	clusterNetwork []string
	serviceNetwork []string
	domain         string
}

func testRemote(nextHop string, clusterNetwork, serviceNetwork []string) testRemoteConfig {
	return testRemoteConfig{
		nextHops:       []string{nextHop},
		clusterNetwork: clusterNetwork,
		serviceNetwork: serviceNetwork,
	}
}

func testRemoteWithDomain(nextHop string, clusterNetwork, serviceNetwork []string, domain string) testRemoteConfig {
	return testRemoteConfig{
		nextHops:       []string{nextHop},
		clusterNetwork: clusterNetwork,
		serviceNetwork: serviceNetwork,
		domain:         domain,
	}
}

func testDualStackRemote(nextHops []string, clusterNetwork, serviceNetwork []string) testRemoteConfig {
	return testRemoteConfig{
		nextHops:       nextHops,
		clusterNetwork: clusterNetwork,
		serviceNetwork: serviceNetwork,
	}
}

func parseNextHops(t *testing.T, hops []string) map[int]net.IP {
	t.Helper()
	m := make(map[int]net.IP, len(hops))
	for _, h := range hops {
		ip := net.ParseIP(h)
		require.NotNil(t, ip, "invalid nextHop: %s", h)
		family := netlink.FAMILY_V4
		if ip.To4() == nil {
			family = netlink.FAMILY_V6
		}
		m[family] = ip
	}
	return m
}

func testConfigWithRemotes(t *testing.T, remotes ...testRemoteConfig) *config.Config {
	t.Helper()

	cfg := &config.Config{}
	cfg.Node.NodeIP = "192.168.1.1"
	cfg.Network.ServiceNetwork = []string{"10.43.0.0/16"}

	for _, r := range remotes {
		resolved := config.ResolvedRemoteCluster{
			NextHops: parseNextHops(t, r.nextHops),
			Domain:   r.domain,
		}

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

		// Compute ProbeIPs from ServiceNetwork (like the real parser does)
		resolved.ProbeIPs = make(map[int]string, len(resolved.ServiceNetwork))
		for _, svcNet := range resolved.ServiceNetwork {
			probeIP, err := cidr.Host(svcNet, 11)
			require.NoError(t, err)
			family := netlink.FAMILY_V4
			if svcNet.IP.To4() == nil {
				family = netlink.FAMILY_V6
			}
			resolved.ProbeIPs[family] = probeIP.String()
		}

		require.NotNil(t, resolved.PrimaryNextHop(), "no valid nextHops in: %v", r.nextHops)
		cfg.C2CC.Resolved = append(cfg.C2CC.Resolved, resolved)
		cfg.C2CC.ResolvedAllCIDRs = append(cfg.C2CC.ResolvedAllCIDRs, resolved.AllCIDRs()...)
	}

	cfg.C2CC.ResolvedRouteTableID = 200
	cfg.C2CC.ResolvedServiceRouteTableID = 201

	return cfg
}
