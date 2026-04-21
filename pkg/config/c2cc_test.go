package config

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestC2CC_IsEnabled(t *testing.T) {
	t.Run("empty remote clusters", func(t *testing.T) {
		c := C2CC{}
		assert.False(t, c.IsEnabled())
	})

	t.Run("with remote clusters", func(t *testing.T) {
		c := C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        "10.100.0.2",
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
				{NextHop: "10.0.0.1", ClusterNetwork: []string{"10.45.0.0/16"}, ServiceNetwork: []string{"10.46.0.0/16"}},
				{},
			},
		}
		c.stripEmptyRemoteClusters()
		assert.Len(t, c.RemoteClusters, 1)
		assert.Equal(t, "10.0.0.1", c.RemoteClusters[0].NextHop)
	})

	t.Run("no-op on empty list", func(t *testing.T) {
		c := C2CC{}
		c.stripEmptyRemoteClusters()
		assert.Empty(t, c.RemoteClusters)
	})
}

func mkC2CCConfig(c2cc C2CC) *Config {
	return &Config{
		Network: Network{
			CNIPlugin:      CniPluginOVNK,
			ClusterNetwork: []string{"10.42.0.0/16"},
			ServiceNetwork: []string{"10.43.0.0/16"},
		},
		Node: Node{
			NodeIP: "10.100.0.1",
		},
		C2CC: c2cc,
	}
}

func mkDualStackC2CCConfig(c2cc C2CC) *Config {
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
		C2CC: c2cc,
	}
}

func mkIPv6OnlyC2CCConfig(c2cc C2CC) *Config {
	return &Config{
		Network: Network{
			CNIPlugin:      CniPluginOVNK,
			ClusterNetwork: []string{"fd01::/48"},
			ServiceNetwork: []string{"fd02::/112"},
		},
		Node: Node{
			NodeIP: "fd00::1",
		},
		C2CC: c2cc,
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
					NextHop:        "10.100.0.2",
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
						NextHop:        "10.100.0.2",
						ClusterNetwork: []string{"10.45.0.0/16"},
						ServiceNetwork: []string{"10.46.0.0/16"},
					},
					{
						NextHop:        "10.100.0.3",
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
					NextHop:        "10.100.0.2",
					ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
					ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
				}},
			}),
		},
		{
			name: "invalid NextHop",
			cfg: mkC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        "not-an-ip",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.2",
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
						NextHop:        "10.100.0.2",
						ClusterNetwork: []string{"10.45.0.0/16"},
						ServiceNetwork: []string{"10.46.0.0/16"},
					},
					{
						NextHop:        "10.100.0.3",
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
						NextHop:        "10.100.0.2",
						ClusterNetwork: []string{"10.45.0.0/16"},
						ServiceNetwork: []string{"10.46.0.0/16"},
					},
					{
						NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.1",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.2",
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
						NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "fd00::2",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "fd00:0:0:0:0:0:0:1",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.2",
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
					NextHop:        "10.100.0.2",
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
						NextHop:        "10.100.0.2",
						ClusterNetwork: []string{"10.45.0.0/16"},
						ServiceNetwork: []string{"10.46.0.0/16"},
						Domain:         "cluster-b.remote",
					},
					{
						NextHop:        "10.100.0.3",
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
			name: "single-stack IPv4 remote with dual-stack local",
			cfg: mkDualStackC2CCConfig(C2CC{
				RemoteClusters: []RemoteCluster{{
					NextHop:        "10.100.0.2",
					ClusterNetwork: []string{"10.45.0.0/16"},
					ServiceNetwork: []string{"10.46.0.0/16"},
				}},
			}),
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
				NextHop:        "fd00::2",
				ClusterNetwork: []string{"fd03::/48"},
				ServiceNetwork: []string{"fd04::/112"},
			}},
		})
		assert.NoError(t, cfg.C2CC.validate(cfg))
	})

	t.Run("dual-stack remote with dual-stack local", func(t *testing.T) {
		cfg := mkDualStackC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        "10.100.0.2",
				ClusterNetwork: []string{"10.45.0.0/16", "fd03::/48"},
				ServiceNetwork: []string{"10.46.0.0/16", "fd04::/112"},
			}},
		})
		assert.NoError(t, cfg.C2CC.validate(cfg))
	})

	t.Run("single-stack IPv6 remote with dual-stack local", func(t *testing.T) {
		cfg := mkDualStackC2CCConfig(C2CC{
			RemoteClusters: []RemoteCluster{{
				NextHop:        "fd00::2",
				ClusterNetwork: []string{"fd03::/48"},
				ServiceNetwork: []string{"fd04::/112"},
			}},
		})
		assert.NoError(t, cfg.C2CC.validate(cfg))
	})
}
