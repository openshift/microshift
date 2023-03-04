package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	IS_DEFAULT_NODENAME     = true
	IS_NOT_DEFAULT_NODENAME = false
)

func setupSuiteDataDir(t *testing.T) func() {
	tmpdir, err := os.MkdirTemp("", "microshift")
	if err != nil {
		t.Errorf("failed to create temp dir: %v", err)
	}
	dataDir = tmpdir
	return func() {
		os.RemoveAll(tmpdir)
	}
}

func TestParse(t *testing.T) {
	mkConfig := func() *Config {
		c := NewMicroshiftConfig()
		c.ApiServer.SkipInterface = true
		return c
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
			expected: mkConfig(),
		},
		{
			name: "dns",
			config: `
dns:
  baseDomain: test-example.com
`,
			expected: func() *Config {
				c := mkConfig()
				c.DNS.BaseDomain = "test-example.com"
				return c
			}(),
		},
		{
			name: "network",
			config: `
network:
  clusterNetwork:
  - cidr: "10.20.30.40/16"
  serviceNetwork:
  - "40.30.20.10/16"
  serviceNodePortRange: "1024-32767"
`,
			expected: func() *Config {
				c := mkConfig()
				c.Network.ClusterNetwork = []ClusterNetworkEntry{
					{
						CIDR: "10.20.30.40/16",
					},
				}
				c.Network.ServiceNetwork = []string{"40.30.20.10/16"}
				c.Network.ServiceNodePortRange = "1024-32767"
				c.ApiServer.AdvertiseAddress = "40.30.0.1"
				c.updateComputedValues() // recomputes DNS field
				return c
			}(),
		},
		{
			name: "node",
			config: `
node:
  hostnameOverride: "node1"
  nodeIP: "1.2.3.4"
`,
			expected: func() *Config {
				c := mkConfig()
				c.Node.HostnameOverride = "node1"
				c.Node.NodeIP = "1.2.3.4"
				return c
			}(),
		},
		{
			name: "api-server",
			config: `
apiServer:
  subjectAltNames:
  - node1
  - node2
`,
			expected: func() *Config {
				c := mkConfig()
				c.ApiServer.SubjectAltNames = []string{
					"node1", "node2",
				}
				return c
			}(),
		},
		{
			name: "debugging",
			config: `
debugging:
  logLevel: Info
`,
			expected: func() *Config {
				c := mkConfig()
				c.Debugging.LogLevel = "Info"
				return c
			}(),
		},
		{
			name: "etcd",
			config: `
etcd:
  quotaBackendSize: 100Gi
  minDefragSize: 1Gi
  maxFragmentedPercentage: 99
  defragCheckFreq: 55m
  doStartupDefrag: true
`,
			expected: func() *Config {
				c := mkConfig()
				c.Etcd.QuotaBackendSize = "100Gi"
				c.Etcd.MinDefragSize = "1Gi"
				c.Etcd.MaxFragmentedPercentage = 99
				c.Etcd.DefragCheckFreq = "55m"
				c.Etcd.DoStartupDefrag = true
				c.updateComputedValues() // parse the time duration
				// set some of the other expected values explicitly
				c.Etcd.QuotaBackendBytes = 100 * 1024 * 1024 * 1024
				c.Etcd.MinDefragBytes = 1024 * 1024 * 1024
				return c
			}(),
		},
	}

	for _, tt := range ttests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parse([]byte(tt.config))
			if tt.expectErr && err == nil {
				t.Fatal("Expecting error and received nothing")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("Not expecting error and received: %v", err)
			}
			if !tt.expectErr {
				assert.Equal(t, tt.expected, config)
			}
		})
	}
}

// Test the validation logic
func TestValidate(t *testing.T) {
	cleanup := setupSuiteDataDir(t)
	defer cleanup()

	mkConfig := func() *Config {
		c := NewMicroshiftConfig()
		c.ApiServer.SkipInterface = true
		return c
	}

	var ttests = []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name:      "defaults-ok",
			config:    NewMicroshiftConfig(),
			expectErr: false,
		},
		{
			name: "subject-alt-names-with-localhost",
			config: func() *Config {
				c := mkConfig()
				c.ApiServer.SubjectAltNames = []string{"localhost"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "subject-alt-names-with-loopback-ipv4",
			config: func() *Config {
				c := mkConfig()
				c.ApiServer.SubjectAltNames = []string{"127.0.0.1"}
				return c
			}(),
			expectErr: true,
		},
		{
			name: "subject-alt-names-with-kubernetes",
			config: func() *Config {
				c := mkConfig()
				c.ApiServer.SubjectAltNames = []string{"kubernetes"}
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
	c := NewMicroshiftConfig()
	if !c.isDefaultNodeName() {
		t.Errorf("expected default IsDefaultNodeName to be true")
	}

	c.Node.HostnameOverride += "-suffix"
	if c.isDefaultNodeName() {
		t.Errorf("expected default IsDefaultNodeName to be false")
	}
}

func TestMicroshiftConfigNodeNameValidation(t *testing.T) {
	cleanup := setupSuiteDataDir(t)
	defer cleanup()

	c := NewMicroshiftConfig()
	c.Node.HostnameOverride = "node1"

	if err := c.validateNodeName(IS_NOT_DEFAULT_NODENAME); err != nil {
		t.Errorf("failed to validate node name on first call: %v", err)
	}

	nodeNameFile := filepath.Join(dataDir, ".nodename")
	if data, err := os.ReadFile(nodeNameFile); err != nil {
		t.Errorf("failed to read node name from file %q: %v", nodeNameFile, err)
	} else if string(data) != c.Node.HostnameOverride {
		t.Errorf("node name file doesn't match the node name in the saved file: %v", err)
	}

	if err := c.validateNodeName(IS_NOT_DEFAULT_NODENAME); err != nil {
		t.Errorf("failed to validate node name on second call without changes: %v", err)
	}

	c.Node.HostnameOverride = "node2"
	if err := c.validateNodeName(IS_NOT_DEFAULT_NODENAME); err == nil {
		t.Errorf("validation should have failed for nodename change: %v", err)
	}
}

func TestMicroshiftConfigNodeNameValidationFromDefault(t *testing.T) {
	cleanup := setupSuiteDataDir(t)
	defer cleanup()

	c := NewMicroshiftConfig()

	if err := c.validateNodeName(IS_DEFAULT_NODENAME); err != nil {
		t.Errorf("failed to validate node name on first call: %v", err)
	}

	hostName, _ := os.Hostname()
	nodeNameFile := filepath.Join(dataDir, ".nodename")
	if data, err := os.ReadFile(nodeNameFile); err != nil {
		t.Errorf("failed to read node name from file %q: %v", nodeNameFile, err)
	} else if string(data) != hostName {
		t.Errorf("node name file doesn't match the node name in the saved file: %v", err)
	}

	if err := c.validateNodeName(IS_DEFAULT_NODENAME); err != nil {
		t.Errorf("failed to validate node name on second call without changes: %v", err)
	}

	c.Node.HostnameOverride = "node2"
	if err := c.validateNodeName(IS_DEFAULT_NODENAME); err != nil {
		t.Errorf("validation should have failed in this case, it must be a warning in logs: %v", err)
	}
}

func TestMicroshiftConfigNodeNameValidationBadName(t *testing.T) {
	cleanup := setupSuiteDataDir(t)
	defer cleanup()

	c := NewMicroshiftConfig()
	c.Node.HostnameOverride = "1.2.3.4"

	if err := c.validateNodeName(IS_DEFAULT_NODENAME); err == nil {
		t.Errorf("failed to validate node name.")
	}
}
