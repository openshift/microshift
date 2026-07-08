package config

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"k8s.io/utils/ptr"
)

func TestC2CC_IsEnabled(t *testing.T) {
	t.Run("empty remote clusters", func(t *testing.T) {
		c := C2CC{}
		assert.False(t, c.IsEnabled())
	})

	t.Run("with remote clusters", func(t *testing.T) {
		c := C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
			}},
		}
		assert.True(t, c.IsEnabled())
	})
}

func TestC2CC_StripEmptyRemoteClusters(t *testing.T) {
	t.Run("strips zero-value entries", func(t *testing.T) {
		c := C2CC{
			RemoteClusters: []RemoteCluster{{}},
		}
		c.stripEmptyRemoteClusters()
		assert.Empty(t, c.RemoteClusters)
		assert.False(t, c.IsEnabled())
	})

	t.Run("keeps non-empty entries", func(t *testing.T) {
		c := C2CC{
			RemoteClusters: []RemoteCluster{
				{},
				{NextHop: []string{"10.0.0.1"}, ClusterNetwork: []string{"10.45.0.0/16"}, ServiceNetwork: []string{"10.46.0.0/16"}},
				{},
			},
		}
		c.stripEmptyRemoteClusters()
		assert.Len(t, c.RemoteClusters, 1)
		assert.Equal(t, []string{"10.0.0.1"}, c.RemoteClusters[0].NextHop)
	})

	t.Run("no-op on empty list", func(t *testing.T) {
		c := C2CC{}
		c.stripEmptyRemoteClusters()
		assert.Empty(t, c.RemoteClusters)
	})
}

func withDNSDefaults(c2cc C2CC) C2CC {
	if c2cc.DNS.CacheTTL == nil {
		c2cc.DNS.CacheTTL = ptr.To(10)
	}
	if c2cc.DNS.CacheNegativeTTL == nil {
		c2cc.DNS.CacheNegativeTTL = ptr.To(10)
	}
	return c2cc
}

func withRoutingDefaults(c2cc C2CC) C2CC {
	if c2cc.Routing.RouteTableID == nil {
		c2cc.Routing.RouteTableID = ptr.To(200)
	}
	if c2cc.Routing.ServiceRouteTableID == nil {
		c2cc.Routing.ServiceRouteTableID = ptr.To(201)
	}
	return c2cc
}

func mkC2CCConfig(c2cc C2CC) *Config {
	if c2cc.ProbeInterval == "" {
		c2cc.ProbeInterval = "10s"
	}
	return &Config{
		Network: Network{
			CNIPlugin:      CniPluginOVNK,
			ClusterNetwork: []string{"10.42.0.0/16"},
			ServiceNetwork: []string{"10.43.0.0/16"},
		},
		Node: Node{
			NodeIP: "10.100.0.1",
		},
		C2CC: withRoutingDefaults(withDNSDefaults(c2cc)),
	}
}

func mkDualStackC2CCConfig(c2cc C2CC) *Config {
	if c2cc.ProbeInterval == "" {
		c2cc.ProbeInterval = "10s"
	}
	return &Config{
		Network: Network{
			CNIPlugin:      CniPluginOVNK,
			ClusterNetwork: []string{"10.42.0.0/16", "fd01::/48"},
			ServiceNetwork: []string{"10.43.0.0/16", "fd02::/112"},
		},
		Node: Node{
			NodeIP:   "10.100.0.1",
			NodeIPV6: "fd00::1",
		},
		C2CC: withRoutingDefaults(withDNSDefaults(c2cc)),
	}
}

func mkIPv6OnlyC2CCConfig(c2cc C2CC) *Config {
	if c2cc.ProbeInterval == "" {
		c2cc.ProbeInterval = "10s"
	}
	return &Config{
		Network: Network{
			CNIPlugin:      CniPluginOVNK,
			ClusterNetwork: []string{"fd01::/48"},
			ServiceNetwork: []string{"fd02::/112"},
		},
		Node: Node{
			NodeIP: "fd00::1",
		},
		C2CC: withRoutingDefaults(withDNSDefaults(c2cc)),
	}
}

func stubHostIPs(t *testing.T, ips []net.IP) {
	t.Helper()
	orig := getHostIPs
	getHostIPs = func() ([]net.IP, error) { return ips, nil }
	t.Cleanup(func() { getHostIPs = orig })
}

func TestC2CC_Validate(t *testing.T) {
	ttests := []struct {
		name      string
		cfg       *Config
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid single remote IPv4",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
		},
		{
			name: "valid multiple remotes",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{
					{
						NextHop:        []string{"10.100.0.2"},
						ClusterNetwork: []string{"10.45.0.0/16"},
						ServiceNetwork: []string{"10.46.0.0/16"},
					},
					{
						NextHop:        []string{"10.100.0.3"},
						ClusterNetwork: []string{"10.55.0.0/16"},
						ServiceNetwork: []string{"10.56.0.0/16"},
					},
				},
			}),
		},
		{
			name: "valid dual-stack remote with dual-stack local",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2", "fd00::2"},
					ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
					ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
				}},
			}),
		},
		{
			name: "dual-stack CIDRs with only IPv4 nextHop",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
					ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
				}},
			}),
			expectErr: true,
			errMsg:    "has IPv6 CIDRs but no IPv6 nextHop",
		},
		{
			name: "invalid NextHop",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"not-an-ip"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "not a valid IP",
		},
		{
			name: "invalid CIDR format",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"not-a-cidr"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "not a valid CIDR",
		},
		{
			name: "local cluster network overlap",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.42.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "overlaps with",
		},
		{
			name: "local service network overlap",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.43.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "overlaps with",
		},
		{
			name: "same remote clusterNetwork overlaps own serviceNetwork",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.45.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "overlaps with",
		},
		{
			name: "remote A cluster overlaps remote B service",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{
					{
						NextHop:        []string{"10.100.0.2"},
						ClusterNetwork: []string{"10.45.0.0/16"},
						ServiceNetwork: []string{"10.46.0.0/16"},
					},
					{
						NextHop:        []string{"10.100.0.3"},
						ClusterNetwork: []string{"10.55.0.0/16"},
						ServiceNetwork: []string{"10.45.0.0/16"},
					},
				},
			}),
			expectErr: true,
			errMsg:    "overlaps with",
		},
		{
			name: "duplicate NextHop across remotes",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{
					{
						NextHop:        []string{"10.100.0.2"},
						ClusterNetwork: []string{"10.45.0.0/16"},
						ServiceNetwork: []string{"10.46.0.0/16"},
					},
					{
						NextHop:        []string{"10.100.0.2"},
						ClusterNetwork: []string{"10.55.0.0/16"},
						ServiceNetwork: []string{"10.56.0.0/16"},
					},
				},
			}),
			expectErr: true,
			errMsg:    "duplicates remoteClusters[0]",
		},
		{
			name: "NextHop equals local node IP",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.1"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "routing loop",
		},
		{
			name: "CIDR mask too short IPv4",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.0.0.0/4"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "shorter than minimum /8",
		},
		{
			name: "CIDR mask too short IPv6",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2", "fd00::2"},
					ClusterNetwork: []string{"10.45.0.0/16", "fd00::/16"},
					ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
				}},
			}),
			expectErr: true,
			errMsg:    "shorter than minimum /32",
		},
		{
			name: "OVN-K disabled",
			cfg: func() *Config {
				cfg := mkC2CCConfig(C2CC{
					RemoteClusters: []RemoteCluster{{
						NextHop:        []string{"10.100.0.2"},
						ClusterNetwork: []string{"10.45.0.0/16"},
						ServiceNetwork: []string{"10.46.0.0/16"},
					}},
				})
				cfg.Network.CNIPlugin = CniPluginNone
				return cfg
			}(),
			expectErr: true,
			errMsg:    "requires OVN-Kubernetes",
		},
		{
			name: "remote CIDR contains host interface IP",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.100.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "contains host interface IP",
		},
		{
			name: "multiple IPv4 entries in clusterNetwork",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16", "10.47.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "multiple IPv4 entries",
		},
		{
			name: "multiple IPv6 entries in clusterNetwork",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"fd03::/48", "fd05::/48"},
					ServiceNetwork: []string{"fd04::/112"},
				}},
			}),
			expectErr: true,
			errMsg:    "multiple IPv6 entries",
		},
		{
			name: "IPv6 remote with IPv4-only local",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"fd00::2"},
					ClusterNetwork: []string{"fd03::/48"},
					ServiceNetwork: []string{"fd04::/112"},
				}},
			}),
			expectErr: true,
			errMsg:    "is IPv6 but local cluster has no IPv6 network",
		},
		{
			name: "IPv4 remote with IPv6-only local",
			cfg: mkIPv6OnlyC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "is IPv4 but local cluster has no IPv4 network",
		},
		{
			name: "NextHop equals local node IP non-canonical form",
			cfg: mkIPv6OnlyC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"fd00:0:0:0:0:0:0:1"},
					ClusterNetwork: []string{"fd03::/48"},
					ServiceNetwork: []string{"fd04::/112"},
				}},
			}),
			expectErr: true,
			errMsg:    "routing loop",
		},
		{
			name: "empty clusterNetwork",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "clusterNetwork must not be empty",
		},
		{
			name: "empty serviceNetwork",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{},
				}},
			}),
			expectErr: true,
			errMsg:    "serviceNetwork must not be empty",
		},
		{
			name: "invalid domain",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
					Domain:         "not a valid domain!",
				}},
			}),
			expectErr: true,
			errMsg:    "is not a valid DNS name",
		},
		{
			name: "valid domain",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
					Domain:         "cluster-b.remote",
				}},
			}),
		},
		{
			name: "duplicate domain across remotes",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{
					{
						NextHop:        []string{"10.100.0.2"},
						ClusterNetwork: []string{"10.45.0.0/16"},
						ServiceNetwork: []string{"10.46.0.0/16"},
						Domain:         "cluster-b.remote",
					},
					{
						NextHop:        []string{"10.100.0.3"},
						ClusterNetwork: []string{"10.55.0.0/16"},
						ServiceNetwork: []string{"10.56.0.0/16"},
						Domain:         "cluster-b.remote",
					},
				},
			}),
			expectErr: true,
			errMsg:    "domain \"cluster-b.remote\" duplicates remoteClusters[0]",
		},
		{
			name: "NextHop equals local NodeIPV6 in dual-stack",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"fd00::1"},
					ClusterNetwork: []string{"fd03::/48"},
					ServiceNetwork: []string{"fd04::/112"},
				}},
			}),
			expectErr: true,
			errMsg:    "routing loop",
		},
		{
			name: "single-stack IPv4 remote with dual-stack local",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
		},
		{
			name: "clusterNetwork and serviceNetwork cardinality mismatch",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2", "fd00::2"},
					ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "different cardinality",
		},
		{
			name: "clusterNetwork and serviceNetwork IP family mismatch at same index",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2", "fd00::2"},
					ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
					ServiceNetwork: []string{"fd04::/112", "10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "mismatched IP families",
		},
		{
			name: "negative cacheTTL",
			cfg: mkC2CCConfig(C2CC{
				DNS: C2CCDNS{CacheTTL: ptr.To(-1), CacheNegativeTTL: ptr.To(10)},
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "dns.cacheTTL must be >= 0",
		},
		{
			name: "negative cacheNegativeTTL",
			cfg: mkC2CCConfig(C2CC{
				DNS: C2CCDNS{CacheTTL: ptr.To(10), CacheNegativeTTL: ptr.To(-5)},
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "dns.cacheNegativeTTL must be >= 0",
		},
		{
			name: "zero cacheTTL is valid",
			cfg: mkC2CCConfig(C2CC{
				DNS: C2CCDNS{CacheTTL: ptr.To(0), CacheNegativeTTL: ptr.To(0)},
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
		},
		{
			name: "multiple IPv4 nextHops - should fail",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2", "10.100.0.3"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "has multiple IPv4 addresses (max 1 per family)",
		},
		{
			name: "multiple IPv6 nextHops - should fail",
			cfg: mkIPv6OnlyC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"fd00::2", "fd00::3"},
					ClusterNetwork: []string{"fd03::/48"},
					ServiceNetwork: []string{"fd04::/112"},
				}},
			}),
			expectErr: true,
			errMsg:    "has multiple IPv6 addresses (max 1 per family)",
		},
		{
			name: "dual-stack nextHop with dual-stack CIDRs - valid",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2", "fd00::2"},
					ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
					ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
				}},
			}),
		},
		{
			name: "dual-stack nextHop order reversed - valid",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"fd00::2", "10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
					ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
				}},
			}),
		},
		{
			name: "multiple IPv4 nextHops in dual-stack config - should fail",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2", "10.100.0.3", "fd00::2"},
					ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
					ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
				}},
			}),
			expectErr: true,
			errMsg:    "has multiple IPv4 addresses (max 1 per family)",
		},
		{
			name: "multiple IPv6 nextHops in dual-stack config - should fail",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2", "fd00::2", "fd00::3"},
					ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
					ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
				}},
			}),
			expectErr: true,
			errMsg:    "has multiple IPv6 addresses (max 1 per family)",
		},
		{
			name: "empty nextHop array - should fail",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
			expectErr: true,
			errMsg:    "nextHop must not be empty",
		},
	}

	for _, tt := range ttests {
		t.Run(tt.name, func(t *testing.T) {
			stubHostIPs(t, []net.IP{net.ParseIP("10.100.0.1")})

			err := tt.cfg.C2CC.validate(tt.cfg)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestC2CC_ValidateDualStack(t *testing.T) {
	stubHostIPs(t, nil)

	t.Run("valid IPv6-only remote with IPv6-only local", func(t *testing.T) {
		cfg := mkIPv6OnlyC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"fd00::2"},
				ClusterNetwork: []string{"fd03::/48"},
				ServiceNetwork: []string{"fd04::/112"},
			}},
		})
		assert.NoError(t, cfg.C2CC.validate(cfg))
	})

	t.Run("dual-stack remote with dual-stack local", func(t *testing.T) {
		cfg := mkDualStackC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2", "fd00::2"},
				ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
				ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
			}},
		})
		assert.NoError(t, cfg.C2CC.validate(cfg))
	})

	t.Run("single-stack IPv6 remote with dual-stack local", func(t *testing.T) {
		cfg := mkDualStackC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"fd00::2"},
				ClusterNetwork: []string{"fd03::/48"},
				ServiceNetwork: []string{"fd04::/112"},
			}},
		})
		assert.NoError(t, cfg.C2CC.validate(cfg))
	})
}

func TestC2CC_ProbeIntervalValidation(t *testing.T) {
	stubHostIPs(t, nil)

	t.Run("too low", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			ProbeInterval: "500ms",
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
			}},
		})
		err := cfg.C2CC.validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "probeInterval must be between 1s and 5m")
	})

	t.Run("too high", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			ProbeInterval: "6m",
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
			}},
		})
		err := cfg.C2CC.validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "probeInterval must be between 1s and 5m")
	})

	t.Run("invalid duration string", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			ProbeInterval: "not-a-duration",
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
			}},
		})
		err := cfg.C2CC.validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a valid duration")
	})

	t.Run("minimum boundary", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			ProbeInterval: "1s",
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
			}},
		})
		assert.NoError(t, cfg.C2CC.validate(cfg))
	})

	t.Run("maximum boundary", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			ProbeInterval: "5m",
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
			}},
		})
		assert.NoError(t, cfg.C2CC.validate(cfg))
	})

	t.Run("valid mid-range value", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			ProbeInterval: "30s",
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		assert.Equal(t, 30*time.Second, cfg.C2CC.ResolvedProbeInterval)
	})
}

func TestC2CC_ProbeIPs(t *testing.T) {
	stubHostIPs(t, nil)

	t.Run("IPv4 service network", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		require.Len(t, cfg.C2CC.Resolved, 1)
		assert.Equal(t, "10.46.0.11", cfg.C2CC.Resolved[0].ProbeIPs[2]) // netlink.FAMILY_V4 = 2
	})

	t.Run("IPv6 service network", func(t *testing.T) {
		cfg := mkIPv6OnlyC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"fd00::2"},
				ClusterNetwork: []string{"fd03::/48"},
				ServiceNetwork: []string{"fd04::/112"},
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		require.Len(t, cfg.C2CC.Resolved, 1)
		assert.Equal(t, "fd04::b", cfg.C2CC.Resolved[0].ProbeIPs[10]) // netlink.FAMILY_V6 = 10
	})

	t.Run("dual-stack populates both families", func(t *testing.T) {
		cfg := mkDualStackC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2", "fd00::2"},
				ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
				ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		require.Len(t, cfg.C2CC.Resolved, 1)
		assert.Equal(t, "10.46.0.11", cfg.C2CC.Resolved[0].ProbeIPs[2]) // IPv4
		assert.Equal(t, "fd04::b", cfg.C2CC.Resolved[0].ProbeIPs[10])   // IPv6
	})
}

func TestC2CC_PrimaryNextHop(t *testing.T) {
	stubHostIPs(t, nil)

	t.Run("IPv4-only returns IPv4", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		require.Len(t, cfg.C2CC.Resolved, 1)
		assert.Equal(t, "10.100.0.2", cfg.C2CC.Resolved[0].PrimaryNextHop().String())
	})

	t.Run("IPv6-only returns IPv6", func(t *testing.T) {
		cfg := mkIPv6OnlyC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"fd00::2"},
				ClusterNetwork: []string{"fd03::/48"},
				ServiceNetwork: []string{"fd04::/112"},
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		require.Len(t, cfg.C2CC.Resolved, 1)
		assert.Equal(t, "fd00::2", cfg.C2CC.Resolved[0].PrimaryNextHop().String())
	})

	t.Run("dual-stack prefers IPv4 even when IPv6 is first", func(t *testing.T) {
		cfg := mkDualStackC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2", "fd00::2"},
				ClusterNetwork: []string{"fd03::/48", "10.45.0.0/16"}, // IPv6 first
				ServiceNetwork: []string{"fd04::/112", "10.46.0.0/16"},
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		require.Len(t, cfg.C2CC.Resolved, 1)
		// Should still return IPv4 because PrimaryNextHop prefers IPv4
		assert.Equal(t, "10.100.0.2", cfg.C2CC.Resolved[0].PrimaryNextHop().String())
	})
}

func TestC2CC_DNSIPs(t *testing.T) {
	stubHostIPs(t, nil)

	t.Run("DNSIPs populated when domain is set", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
				Domain:         "cluster-b.remote",
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		assert.Equal(t, []string{"10.46.0.10"}, cfg.C2CC.Resolved[0].DNSIPs)
	})

	t.Run("DNSIPs empty when domain is not set", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2"},
				ClusterNetwork: []string{"10.45.0.0/16"},
				ServiceNetwork: []string{"10.46.0.0/16"},
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		assert.Empty(t, cfg.C2CC.Resolved[0].DNSIPs)
	})

	t.Run("DNSIPs for IPv6 service network", func(t *testing.T) {
		cfg := mkIPv6OnlyC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"fd00::2"},
				ClusterNetwork: []string{"fd03::/48"},
				ServiceNetwork: []string{"fd04::/112"},
				Domain:         "cluster-b.remote",
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		assert.Equal(t, []string{"fd04::a"}, cfg.C2CC.Resolved[0].DNSIPs)
	})

	t.Run("DNSIPs for dual-stack service network", func(t *testing.T) {
		cfg := mkDualStackC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        []string{"10.100.0.2", "fd00::2"},
				ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
				ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
				Domain:         "cluster-b.remote",
			}},
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		assert.Equal(t, []string{"10.46.0.10", "fd04::a"}, cfg.C2CC.Resolved[0].DNSIPs)
	})
}

func parseCIDR(t *testing.T, s string) *net.IPNet {
	t.Helper()
	_, ipNet, err := net.ParseCIDR(s)
	require.NoError(t, err)
	return ipNet
}

func TestRenderC2CCDNSBlocks(t *testing.T) {
	t.Run("no domains configured", func(t *testing.T) {
		resolved := []ResolvedRemoteCluster{{
			NextHops:       map[int]net.IP{netlink.FAMILY_V4: net.ParseIP("10.100.0.2")},
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.45.0.0/16")},
			ServiceNetwork: []*net.IPNet{parseCIDR(t, "10.46.0.0/16")},
		}}
		result := RenderC2CCDNSBlocks(resolved, 10, 10)
		assert.Empty(t, result)
	})

	t.Run("single domain with default TTLs", func(t *testing.T) {
		resolved := []ResolvedRemoteCluster{{
			NextHops:       map[int]net.IP{netlink.FAMILY_V4: net.ParseIP("10.100.0.2")},
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.45.0.0/16")},
			ServiceNetwork: []*net.IPNet{parseCIDR(t, "10.46.0.0/16")},
			Domain:         "cluster-b.remote",
			DNSIPs:         []string{"10.46.0.10"},
		}}
		result := RenderC2CCDNSBlocks(resolved, 10, 10)
		assert.True(t, strings.HasPrefix(result, "\n"), "result should start with newline for YAML block scalar")
		assert.Contains(t, result, "cluster-b.remote:5353")
		assert.Contains(t, result, "rewrite stop name suffix .cluster-b.remote .cluster.local answer auto")
		assert.Contains(t, result, "forward . 10.46.0.10")
		assert.Contains(t, result, "cache 10 {")
		assert.Contains(t, result, "denial 9984 10")
	})

	t.Run("custom TTLs", func(t *testing.T) {
		resolved := []ResolvedRemoteCluster{{
			Domain: "cluster-b.remote",
			DNSIPs: []string{"10.46.0.10"},
		}}
		result := RenderC2CCDNSBlocks(resolved, 30, 60)
		assert.Contains(t, result, "cache 30 {")
		assert.Contains(t, result, "denial 9984 60")
	})

	t.Run("zero TTLs", func(t *testing.T) {
		resolved := []ResolvedRemoteCluster{{
			Domain: "cluster-b.remote",
			DNSIPs: []string{"10.46.0.10"},
		}}
		result := RenderC2CCDNSBlocks(resolved, 0, 0)
		assert.Contains(t, result, "cache 0 {")
		assert.Contains(t, result, "denial 9984 0")
	})

	t.Run("multiple domains", func(t *testing.T) {
		resolved := []ResolvedRemoteCluster{
			{
				Domain: "cluster-b.remote",
				DNSIPs: []string{"10.46.0.10"},
			},
			{
				Domain: "cluster-c.remote",
				DNSIPs: []string{"10.56.0.10"},
			},
		}
		result := RenderC2CCDNSBlocks(resolved, 10, 10)
		assert.Contains(t, result, "cluster-b.remote:5353")
		assert.Contains(t, result, "forward . 10.46.0.10")
		assert.Contains(t, result, "cluster-c.remote:5353")
		assert.Contains(t, result, "forward . 10.56.0.10")
	})

	t.Run("mixed domain and no-domain", func(t *testing.T) {
		resolved := []ResolvedRemoteCluster{
			{
				Domain: "cluster-b.remote",
				DNSIPs: []string{"10.46.0.10"},
			},
			{
				Domain: "",
			},
		}
		result := RenderC2CCDNSBlocks(resolved, 10, 10)
		assert.Contains(t, result, "cluster-b.remote:5353")
		assert.NotContains(t, result, "cluster-c")
	})

	t.Run("dual-stack DNS IPs", func(t *testing.T) {
		resolved := []ResolvedRemoteCluster{{
			NextHops:       map[int]net.IP{netlink.FAMILY_V4: net.ParseIP("10.100.0.2"), netlink.FAMILY_V6: net.ParseIP("fd00::2")},
			ClusterNetwork: []*net.IPNet{parseCIDR(t, "10.45.0.0/16"), parseCIDR(t, "fd03::/48")},
			ServiceNetwork: []*net.IPNet{parseCIDR(t, "10.46.0.0/16"), parseCIDR(t, "fd04::/112")},
			Domain:         "cluster-b.remote",
			DNSIPs:         []string{"10.46.0.10", "fd04::a"},
		}}
		result := RenderC2CCDNSBlocks(resolved, 10, 10)
		assert.Contains(t, result, "cluster-b.remote:5353")
		assert.Contains(t, result, "forward . 10.46.0.10 fd04::a")
	})
}

func TestC2CC_RoutingTableValidation(t *testing.T) {
	stubHostIPs(t, nil)

	validRemote := []RemoteCluster{{
		NextHop:        []string{"10.100.0.2"},
		ClusterNetwork: []string{"10.45.0.0/16"},
		ServiceNetwork: []string{"10.46.0.0/16"},
	}}

	t.Run("valid custom routing table IDs", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			Routing:        C2CCRouting{RouteTableID: ptr.To(100), ServiceRouteTableID: ptr.To(101)},
			RemoteClusters: validRemote,
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		assert.Equal(t, 100, cfg.C2CC.ResolvedRouteTableID)
		assert.Equal(t, 101, cfg.C2CC.ResolvedServiceRouteTableID)
	})

	t.Run("boundary values 1 and 252", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			Routing:        C2CCRouting{RouteTableID: ptr.To(1), ServiceRouteTableID: ptr.To(252)},
			RemoteClusters: validRemote,
		})
		require.NoError(t, cfg.C2CC.validate(cfg))
		assert.Equal(t, 1, cfg.C2CC.ResolvedRouteTableID)
		assert.Equal(t, 252, cfg.C2CC.ResolvedServiceRouteTableID)
	})

	t.Run("routeTableID below range", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			Routing:        C2CCRouting{RouteTableID: ptr.To(0), ServiceRouteTableID: ptr.To(201)},
			RemoteClusters: validRemote,
		})
		err := cfg.C2CC.validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "routing.routeTableID must be between 1 and 252")
	})

	t.Run("serviceRouteTableID above range", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			Routing:        C2CCRouting{RouteTableID: ptr.To(200), ServiceRouteTableID: ptr.To(253)},
			RemoteClusters: validRemote,
		})
		err := cfg.C2CC.validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "routing.serviceRouteTableID must be between 1 and 252")
	})

	t.Run("duplicate table IDs", func(t *testing.T) {
		cfg := mkC2CCConfig(C2CC{
			Routing:        C2CCRouting{RouteTableID: ptr.To(150), ServiceRouteTableID: ptr.To(150)},
			RemoteClusters: validRemote,
		})
		err := cfg.C2CC.validate(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must differ")
	})

	t.Run("defaults are used when nil", func(t *testing.T) {
		c2cc := C2CC{RemoteClusters: validRemote, ProbeInterval: "10s"}
		cfg := &Config{
			Network: Network{
				CNIPlugin:      CniPluginOVNK,
				ClusterNetwork: []string{"10.42.0.0/16"},
				ServiceNetwork: []string{"10.43.0.0/16"},
			},
			Node: Node{NodeIP: "10.100.0.1"},
			C2CC: withDNSDefaults(c2cc),
		}
		require.NoError(t, cfg.C2CC.validate(cfg))
		assert.Equal(t, 200, cfg.C2CC.ResolvedRouteTableID)
		assert.Equal(t, 201, cfg.C2CC.ResolvedServiceRouteTableID)
	})
}

func TestC2CC_ProbeIntervalDefault(t *testing.T) {
	cfg := &Config{}
	require.NoError(t, cfg.fillDefaults())
	assert.Equal(t, "10s", cfg.C2CC.ProbeInterval)
}

func TestC2CC_IncorporateUserSettings(t *testing.T) {
	t.Run("user overrides probe interval", func(t *testing.T) {
		cfg := &Config{}
		require.NoError(t, cfg.fillDefaults())

		user := &Config{
			C2CC: C2CC{
				ProbeInterval: "30s",
			},
		}
		cfg.incorporateUserSettings(user)
		assert.Equal(t, "30s", cfg.C2CC.ProbeInterval)
	})

	t.Run("user overrides routing table IDs", func(t *testing.T) {
		cfg := &Config{}
		require.NoError(t, cfg.fillDefaults())

		user := &Config{
			C2CC: C2CC{
				Routing: C2CCRouting{
					RouteTableID:        ptr.To(100),
					ServiceRouteTableID: ptr.To(101),
				},
			},
		}
		cfg.incorporateUserSettings(user)
		assert.Equal(t, 100, *cfg.C2CC.Routing.RouteTableID)
		assert.Equal(t, 101, *cfg.C2CC.Routing.ServiceRouteTableID)
	})

	t.Run("user overrides only one routing table ID preserves other default", func(t *testing.T) {
		cfg := &Config{}
		require.NoError(t, cfg.fillDefaults())

		user := &Config{
			C2CC: C2CC{
				Routing: C2CCRouting{
					RouteTableID: ptr.To(100),
				},
			},
		}
		cfg.incorporateUserSettings(user)
		assert.Equal(t, 100, *cfg.C2CC.Routing.RouteTableID)
		assert.Equal(t, 201, *cfg.C2CC.Routing.ServiceRouteTableID)
	})

	t.Run("user sets remoteClusters without probeInterval preserves default", func(t *testing.T) {
		cfg := &Config{}
		require.NoError(t, cfg.fillDefaults())

		user := &Config{
			C2CC: C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        []string{"10.100.0.2"},
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			},
		}
		cfg.incorporateUserSettings(user)
		assert.Equal(t, "10s", cfg.C2CC.ProbeInterval)
		assert.Len(t, cfg.C2CC.RemoteClusters, 1)
	})
}
