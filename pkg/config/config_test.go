package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

const (
	IS_DEFAULT_NODENAME     = true
	IS_NOT_DEFAULT_NODENAME = false
)

func setupSuiteDataDir(t *testing.T) (string, func()) {
	tmpdir, err := os.MkdirTemp("", "microshift")
	if err != nil {
		t.Errorf("failed to create temp dir: %v", err)
	}
	return tmpdir, func() {
		os.RemoveAll(tmpdir)
	}
}

// TestGetActiveConfigFromYAML verifies that reading the config file
// correctly overrides the defaults and updates the computed values in
// the Config struct.
func TestGetActiveConfigFromYAML(t *testing.T) {
	mkDefaultConfig := func() *Config {
		c := NewDefault()
		return c
	}

	dedent := func(input string) string {
		lines := strings.Split(input, "\n")
		detectIndentFrom := lines[0]
		if detectIndentFrom == "" {
			detectIndentFrom = lines[1]
		}
		dedentedLine := strings.TrimLeft(detectIndentFrom, " \t")
		prefixLen := len(detectIndentFrom) - len(dedentedLine)
		if prefixLen == 0 {
			return input
		}
		var b strings.Builder
		for _, line := range lines {
			if len(line) >= prefixLen {
				line = line[prefixLen:]
			}
			fmt.Fprintf(&b, "%s\n", line)
		}
		return b.String()
	}

	var ttests = []struct {
		name      string
		config    string
		expected  *Config
		expectErr bool
	}{
		{
			name:     "empty",
			config:   "",
			expected: mkDefaultConfig(),
		},
		{
			name: "dns",
			config: dedent(`
            dns:
              baseDomain: test-example.com
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.DNS.BaseDomain = "test-example.com"
				return c
			}(),
		},
		{
			name: "network",
			config: dedent(`
            network:
              clusterNetwork:
                - "10.20.30.40/16"
              serviceNetwork:
                - "40.30.20.10/16"
              serviceNodePortRange: "1024-32767"
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Network.ClusterNetwork = []string{"10.20.30.40/16"}
				c.Network.ServiceNetwork = []string{"40.30.20.10/16"}
				c.Network.ServiceNodePortRange = "1024-32767"
				c.ApiServer.AdvertiseAddress = ""           // force value to be recomputed
				c.Ingress.ListenAddress = nil               // force value to be recomputed
				assert.NoError(t, c.updateComputedValues()) // recomputes DNS field
				return c
			}(),
		},
		{
			name: "node",
			config: dedent(`
            node:
              hostnameOverride: "node1"
              nodeIP: "1.2.3.4"
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Node.HostnameOverride = "node1"
				c.Node.NodeIP = "1.2.3.4"
				return c
			}(),
		},
		{
			name: "api-server-subject-alt-names",
			config: dedent(`
            apiServer:
              subjectAltNames:
              - node1
              - node2
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.ApiServer.SubjectAltNames = []string{
					"node1", "node2",
				}
				return c
			}(),
		},
		{
			name: "api-server-named-certificates",
			config: dedent(`
			        apiServer:
			          namedCertificates:
			          - certPath: /tmp/fqdn-server-1.pem
			            keyPath: /tmp/fqdn-server-1.key
			            names: 
			            - fqdn-server-1						
			          - certPath: /tmp/fqdn-server-2.pem
			            keyPath: /tmp/fqdn-server-2.key
			        `),

			expected: func() *Config {
				c := mkDefaultConfig()
				c.ApiServer.NamedCertificates = []NamedCertificateEntry{
					{
						Names:    []string{"fqdn-server-1"},
						CertPath: "/tmp/fqdn-server-1.pem",
						KeyPath:  "/tmp/fqdn-server-1.key",
					},
					{
						CertPath: "/tmp/fqdn-server-2.pem",
						KeyPath:  "/tmp/fqdn-server-2.key",
					},
				}
				return c
			}(),
		},
		{
			name: "api-server-advertise-address",
			config: dedent(`
            apiServer:
              advertiseAddress: 127.0.0.1
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.ApiServer.AdvertiseAddress = "127.0.0.1"
				c.ApiServer.SkipInterface = true
				c.Ingress.ListenAddress = nil
				assert.NoError(t, c.updateComputedValues()) // recomputes ingress listenAddress field
				return c
			}(),
		},
		{
			name: "debugging",
			config: dedent(`
            debugging:
              logLevel: Normal
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.computeLoggingSetting()
				return c
			}(),
		},
		{
			name: "debugging-invalid",
			config: dedent(`
            debugging:
              logLevel: Unknown
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.computeLoggingSetting()
				// We expect a warning, but do not want to check for
				// the message string.
				c.Warnings = []string{}
				return c
			}(),
		},
		{
			name: "etcd",
			config: dedent(`
            etcd:
              memoryLimitMB: 129
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Etcd.MemoryLimitMB = 129
				assert.NoError(t, c.updateComputedValues())
				return c
			}(),
		},
		{
			name: "manifests-default",
			config: dedent(`
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				assert.NoError(t, c.updateComputedValues())
				return c
			}(),
		},
		{
			name: "manifests-default2",
			config: dedent(`
            # the manifests struct is present, but no paths are listed
            manifests:
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				assert.NoError(t, c.updateComputedValues())
				return c
			}(),
		},
		{
			name: "manifests-default3",
			config: dedent(`
            # the manifests struct and paths keys are present but the paths
            # item is not a YAML list
            manifests:
              kustomizePaths:
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				assert.NoError(t, c.updateComputedValues())
				return c
			}(),
		},
		{
			name: "manifests-empty-list",
			config: dedent(`
            manifests:
              kustomizePaths: []
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Manifests.KustomizePaths = []string{}
				assert.NoError(t, c.updateComputedValues())
				return c
			}(),
		},
		{
			name: "manifests-override",
			config: dedent(`
            manifests:
              kustomizePaths:
                - /tmp/some/other/directory
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Manifests.KustomizePaths = []string{"/tmp/some/other/directory"}
				assert.NoError(t, c.updateComputedValues())
				return c
			}(),
		},
		{
			name: "router-managed",
			config: dedent(`
            ingress:
              status: Managed
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.Status = StatusManaged
				return c
			}(),
		},
		{
			name: "router-removed",
			config: dedent(`
            ingress:
              status: Removed
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.Status = StatusRemoved
				return c
			}(),
		},
		{
			name: "router-ports",
			config: dedent(`
			ingress:
			  ports:
			    http: 1234
			    https: 9876
			`),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.Ports.Http = ptr.To[int](1234)
				c.Ingress.Ports.Https = ptr.To[int](9876)
				return c
			}(),
		},
		{
			name: "ingress-listen-address-nic",
			config: dedent(`
            ingress:
              listenAddress:
                - lo
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.ListenAddress = []string{"lo"}
				return c
			}(),
		},
		{
			name: "kubelet",
			config: dedent(`
            kubelet:
              cpuManagerPolicy: static
              reservedMemory:
              - limits:
                  memory: 1100Mi
                numaNode: 0
              kubeReserved:
                memory: 500Mi
              evictionHard:
                imagefs.available: 15%
                memory.available: 100Mi
                nodefs.available: 10%
                nodefs.inodesFree: 5%
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Kubelet = map[string]any{
					"cpuManagerPolicy": "static",
					"reservedMemory": []any{
						map[string]any{
							"limits": map[string]any{
								"memory": "1100Mi",
							},
							"numaNode": float64(0),
						},
					},
					"kubeReserved": map[string]any{
						"memory": "500Mi",
					},
					"evictionHard": map[string]any{
						"imagefs.available": "15%",
						"memory.available":  "100Mi",
						"nodefs.available":  "10%",
						"nodefs.inodesFree": "5%",
					},
				}
				return c
			}(),
		}, {
			name: "storage",
			config: dedent(`
			storage:
			  driver: "none"
			  optionalCsiComponents:
			  - "snapshot-controller" 
			`),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.Storage = Storage{
					Driver:                CsiDriverNone,
					OptionalCSIComponents: []OptionalCsiComponent{CsiComponentSnapshot},
				}
				return c
			}(),
		},
	}

	for _, tt := range ttests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := getActiveConfigFromYAMLDropins([][]byte{[]byte(tt.config)})
			// If we have any warnings, drop them. Use an empty array
			// instead of nil so that we can differentiate between
			// unexpected warnings (where we get an array instead of
			// nil) and missing expected warnings (where we get nil
			// but expect an array).
			if config.Warnings != nil {
				config.Warnings = []string{}
			}

			if tt.expectErr && err == nil {
				t.Fatal("Expecting error and received nothing")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("Not expecting error and received: %v", err)
			}
			if !tt.expectErr {
				// blank out the user settings because the expected value
				// never has them and any computed value should be set so
				// it should be safe to ignore them
				config.userSettings = nil

				assert.Equal(t, tt.expected, config, "config input:\n---%s\n---", tt.config)
			}
		})
	}

	t.Run("multiple-drop-ins", func(t *testing.T) {
		dropins := [][]byte{
			// Individual fields should be overwritten
			[]byte(dedent(`
            ingress:
              ports:
                http: 1234
                https: 9876
            `)),
			[]byte(dedent(`
            ingress:
              ports:
                http: 2345
            `)),
			[]byte(dedent(`
            ingress:
              ports:
                https: 8765
            `)),

			// Arrays are overwritten completely (no addition)
			[]byte(dedent(`
            ingress:
              listenAddress:
                - eth1
                - eth2
            `)),
			[]byte(dedent(`
            ingress:
              listenAddress:
                - lo
            `)),

			// Even though kubelet is map[string]any, we want to merge individual settings
			[]byte(dedent(`
            kubelet:
              cpuManagerPolicy: static
              evictionHard:
                imagefs.available: 15%
                memory.available: 100Mi
            `)),
			[]byte(dedent(`
            kubelet:
              memoryManagerPolicy: Static
              evictionHard:
                nodefs.available: 10%
                nodefs.inodesFree: 5%
            `)),
		}

		expected := mkDefaultConfig()
		expected.Ingress.Ports.Http = ptr.To[int](2345)
		expected.Ingress.Ports.Https = ptr.To[int](8765)
		expected.Ingress.ListenAddress = []string{"lo"}
		expected.Kubelet = map[string]any{
			"cpuManagerPolicy":    "static",
			"memoryManagerPolicy": "Static",
			"evictionHard": map[string]any{
				"imagefs.available": "15%",
				"memory.available":  "100Mi",
				"nodefs.available":  "10%",
				"nodefs.inodesFree": "5%",
			},
		}

		config, err := getActiveConfigFromYAMLDropins(dropins)
		assert.NoError(t, err)

		config.userSettings = nil
		assert.Equal(t, expected, config)
	})
}

// Test the validation logic
func TestValidate(t *testing.T) {
	mkDefaultConfig := func() *Config {
		c := NewDefault()
		c.ApiServer.SkipInterface = false
		return c
	}

	var ttests = []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name:      "defaults-ok",
			config:    NewDefault(),
			expectErr: false,
		},
		{
			name: "subject-alt-names-with-localhost",
			config: func() *Config {
				c := mkDefaultConfig()
				c.ApiServer.SubjectAltNames = []string{"localhost"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "subject-alt-names-with-loopback-ipv4",
			config: func() *Config {
				c := mkDefaultConfig()
				c.ApiServer.SubjectAltNames = []string{"127.0.0.1"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "subject-alt-names-with-kubernetes",
			config: func() *Config {
				c := mkDefaultConfig()
				c.ApiServer.SubjectAltNames = []string{"kubernetes"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "etcd-memory-limit-low",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Etcd.MemoryLimitMB = 1
				return c
			}(),
			expectErr: true,
		},
		{
			name: "etcd-memory-zero",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Etcd.MemoryLimitMB = 0
				return c
			}(),
			expectErr: false,
		},
		{
			name: "advertise-address-not-present",
			config: func() *Config {
				c := mkDefaultConfig()
				c.ApiServer.AdvertiseAddress = "8.8.8.8"
				c.ApiServer.SkipInterface = true
				return c
			}(),
			expectErr: true,
		},
		{
			name: "router-status-invalid",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.Status = "invalid"
				return c
			}(),
			expectErr: true,
		},
		{
			name: "ingress-ports-http-invalid-value-1",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.Ports.Http = ptr.To[int](0)
				return c
			}(),
			expectErr: true,
		},
		{
			name: "ingress-listen-address-ip-forbidden",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.ListenAddress = []string{"127.0.0.1", "169.255.169.254"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "ingress-ports-http-invalid-value-2",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.Ports.Http = ptr.To[int](65536)
				return c
			}(),
			expectErr: true,
		},
		{
			name: "ingress-listen-address-ip-not-present",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.ListenAddress = []string{"1.2.3.4"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "ingress-ports-https-invalid-value-1",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.Ports.Https = ptr.To[int](0)
				return c
			}(),
			expectErr: true,
		},
		{
			name: "ingress-ports-https-invalid-value-2",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.Ports.Https = ptr.To[int](65536)
				return c
			}(),
			expectErr: true,
		},
		{
			name: "ingress-listen-address-nic-not-present",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.ListenAddress = []string{"dummyinterface"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "ingress-listen-address-bad-ip-family-1",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.ListenAddress = []string{"1.2.3.4"}
				c.Network.ClusterNetwork = []string{"fd01::/48"}
				c.Network.ServiceNetwork = []string{"fd02::/112"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "ingress-listen-address-bad-ip-family-2",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Ingress.ListenAddress = []string{"fe80::1"}
				c.Network.ClusterNetwork = []string{"10.42.0.0/16"}
				c.Network.ServiceNetwork = []string{"10.43.0.0/16"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "audit-log-flag-values-unexpected-values",
			config: func() *Config {
				c := mkDefaultConfig()
				c.ApiServer.AuditLog.MaxFiles = -1
				c.ApiServer.AuditLog.MaxFileAge = -1
				c.ApiServer.AuditLog.MaxFileSize = -1
				return c
			}(),
			expectErr: true,
		},
		{
			name: "audit-log-flag-expected-values",
			config: func() *Config {
				c := mkDefaultConfig()
				c.ApiServer.AuditLog.MaxFiles = 0
				c.ApiServer.AuditLog.MaxFileAge = 0
				c.ApiServer.AuditLog.MaxFileSize = 0
				return c
			}(),
			expectErr: false,
		},
		{
			name: "network-too-many-entries",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"1.2.3.4/24", "::1/128", "5.6.7.8/24"}
				c.Network.ClusterNetwork = []string{"9.10.11.12/24", "::2/128", "13.14.15.16/24"}
				c.ApiServer.AdvertiseAddress = "17.18.19.20"
				return c
			}(),
			expectErr: true,
		},
		{
			name: "network-same-ip-family-ipv4",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"21.22.23.24/24", "25.26.27.28/24"}
				c.Network.ClusterNetwork = []string{"29.30.31.32/24", "33.34.35.36/24"}
				c.ApiServer.AdvertiseAddress = "37.38.39.40"
				return c
			}(),
			expectErr: true,
		},
		{
			name: "network-same-ip-family-ipv6",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"fd01::/64", "fd02::/64"}
				c.Network.ClusterNetwork = []string{"fd03::/64", "fd04::/64"}
				c.ApiServer.AdvertiseAddress = "fd01::1"
				return c
			}(),
			expectErr: true,
		},
		{
			name: "network-bad-format-ipv4",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"1.2.3.300/24"}
				c.Network.ClusterNetwork = []string{"300.1.2.3/24"}
				c.ApiServer.AdvertiseAddress = "8.8.8.8"
				return c
			}(),
			expectErr: true,
		},
		{
			name: "network-bad-format-ipv6",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"fd01:::/64"}
				c.Network.ClusterNetwork = []string{"fd05::/64"}
				c.ApiServer.AdvertiseAddress = "fd01::2"
				return c
			}(),
			expectErr: true,
		},
		{
			name: "network-different-ip-family",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"fd05::/64"}
				c.Network.ClusterNetwork = []string{"4.3.2.1/24"}
				c.ApiServer.AdvertiseAddress = "fd01::3"
				return c
			}(),
			expectErr: true,
		},
		{
			name: "network-different-ip-family-advertise-address",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"fd06::/64"}
				c.Network.ClusterNetwork = []string{"fd07::/64"}
				c.ApiServer.AdvertiseAddress = "10.20.30.40"
				return c
			}(),
			expectErr: true,
		},
		{
			name: "node-ipv6-must-be-configured",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"90.80.70.60/16", "fd08::/64"}
				c.Network.ClusterNetwork = []string{"50.40.30.20/16", "fd09::/64"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "node-ipv6-must-not-be-configured",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"91.81.71.61/16"}
				c.Network.ClusterNetwork = []string{"51.41.31.21/16"}
				c.Node.NodeIPV6 = "2001:db0:ff::1"
				return c
			}(),
			expectErr: true,
		},
		{
			name: "node-ipv6-bad-format",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"92.82.72.62/16", "fd0a::/64"}
				c.Network.ClusterNetwork = []string{"52.42.32.22/16", "fd0b::/64"}
				c.Node.NodeIPV6 = "2001:db0::ff:::1"
				return c
			}(),
			expectErr: true,
		},
		{
			name: "node-ipv6-must-be-ipv6",
			config: func() *Config {
				c := mkDefaultConfig()
				c.Network.ServiceNetwork = []string{"93.83.73.63/16", "fd0c::/64"}
				c.Network.ClusterNetwork = []string{"53.43.33.23/16", "fd0d::/64"}
				c.Node.NodeIPV6 = "11.22.33.44"
				return c
			}(),
			expectErr: true,
		},
	}
	for _, tt := range ttests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.expectErr && err == nil {
				t.Fatal("Expecting error and received nothing")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("Not expecting error and received: %v", err)
			}
		})
	}
}

func TestMicroshiftConfigIsDefaultNodeName(t *testing.T) {
	c := NewDefault()
	if !c.isDefaultNodeName() {
		t.Errorf("expected default IsDefaultNodeName to be true")
	}

	c.Node.HostnameOverride += "-suffix"
	if c.isDefaultNodeName() {
		t.Errorf("expected default IsDefaultNodeName to be false")
	}
}

func TestCanonicalNodeName(t *testing.T) {
	hostname, _ := os.Hostname()

	var ttests = []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "default",
			value:    "",
			expected: strings.ToLower(hostname),
		},
		{
			name:     "upper-case",
			value:    "Hostname",
			expected: "hostname",
		},
	}

	for _, tt := range ttests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDefault()
			if tt.value != "" { // account for default
				c.Node.HostnameOverride = tt.value
			}
			assert.Equal(t, tt.expected, c.CanonicalNodeName())
		})
	}
}

func TestMicroshiftConfigNodeNameValidation(t *testing.T) {
	dataDir, cleanup := setupSuiteDataDir(t)
	defer cleanup()

	c := NewDefault()
	c.Node.HostnameOverride = "node1"

	if err := c.validateNodeName(IS_NOT_DEFAULT_NODENAME, dataDir); err != nil {
		t.Errorf("failed to validate node name on first call: %v", err)
	}

	nodeNameFile := filepath.Join(dataDir, ".nodename")
	if data, err := os.ReadFile(nodeNameFile); err != nil {
		t.Errorf("failed to read node name from file %q: %v", nodeNameFile, err)
	} else if string(data) != c.Node.HostnameOverride {
		t.Errorf("node name file doesn't match the node name in the saved file: %v", err)
	}

	if err := c.validateNodeName(IS_NOT_DEFAULT_NODENAME, dataDir); err != nil {
		t.Errorf("failed to validate node name on second call without changes: %v", err)
	}

	c.Node.HostnameOverride = "node2"
	if err := c.validateNodeName(IS_NOT_DEFAULT_NODENAME, dataDir); err == nil {
		t.Errorf("validation should have failed for nodename change: %v", err)
	}
}

func TestMicroshiftConfigNodeNameValidationFromDefault(t *testing.T) {
	dataDir, cleanup := setupSuiteDataDir(t)
	defer cleanup()

	c := NewDefault()

	if err := c.validateNodeName(IS_DEFAULT_NODENAME, dataDir); err != nil {
		t.Errorf("failed to validate node name on first call: %v", err)
	}

	hostName, _ := os.Hostname()
	nodeNameFile := filepath.Join(dataDir, ".nodename")
	if data, err := os.ReadFile(nodeNameFile); err != nil {
		t.Errorf("failed to read node name from file %q: %v", nodeNameFile, err)
	} else if string(data) != hostName {
		t.Errorf("node name file doesn't match the node name in the saved file: %v", err)
	}

	if err := c.validateNodeName(IS_DEFAULT_NODENAME, dataDir); err != nil {
		t.Errorf("failed to validate node name on second call without changes: %v", err)
	}

	c.Node.HostnameOverride = "node2"
	if err := c.validateNodeName(IS_DEFAULT_NODENAME, dataDir); err != nil {
		t.Errorf("validation should have failed in this case, it must be a warning in logs: %v", err)
	}
}

func TestMicroshiftConfigNodeNameValidationBadName(t *testing.T) {
	dataDir, cleanup := setupSuiteDataDir(t)
	defer cleanup()

	c := NewDefault()
	c.Node.HostnameOverride = "1.2.3.4"

	if err := c.validateNodeName(IS_DEFAULT_NODENAME, dataDir); err == nil {
		t.Errorf("failed to validate node name.")
	}
}
