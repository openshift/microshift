package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			name: "api-server-advertise-address",
			config: dedent(`
            apiServer:
              advertiseAddress: 127.0.0.1
            `),
			expected: func() *Config {
				c := mkDefaultConfig()
				c.ApiServer.AdvertiseAddress = "127.0.0.1"
				c.ApiServer.SkipInterface = true
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
	}

	for _, tt := range ttests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := getActiveConfigFromYAML([]byte(tt.config))

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
