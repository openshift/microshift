package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"sigs.k8s.io/yaml"
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

func TestConfigFile(t *testing.T) {
	var ttests = []struct {
		config    Config
		expected  MicroshiftConfig
		expectErr bool
	}{
		{
			config: Config{
				DNS: DNS{
					BaseDomain: "example.com",
				},
				Network: Network{
					ClusterNetwork: []ClusterNetworkEntry{
						{
							CIDR: "10.20.30.40/16",
						},
					},
					ServiceNetwork:       []string{"40.30.20.10/16"},
					ServiceNodePortRange: "1024-32767",
				},
				Node: Node{
					HostnameOverride: "node1",
					NodeIP:           "1.2.3.4",
				},
				ApiServer: ApiServer{
					SubjectAltNames:  []string{"node1", "node2"},
					AdvertiseAddress: "6.7.8.9",
				},
				Debugging: Debugging{
					LogLevel: "Debug",
				},
				Etcd: EtcdConfig{
					QuotaBackendSize:        "2Gi",
					MinDefragSize:           "100Mi",
					MaxFragmentedPercentage: 45,
					DefragCheckFreq:         "5m",
					DoStartupDefrag:         true,
				},
			},
			expected: MicroshiftConfig{
				Debugging: Debugging{
					LogLevel: "Debug",
				},
				SubjectAltNames: []string{"node1", "node2"},
				Node: Node{
					HostnameOverride: "node1",
					NodeIP:           "1.2.3.4",
				},
				KASAdvertiseAddress: "6.7.8.9",
				DNS: DNS{
					BaseDomain: "example.com",
				},
				Cluster: ClusterConfig{
					URL:                  "https://localhost:6443",
					ClusterCIDR:          "10.20.30.40/16",
					ServiceCIDR:          "40.30.20.10/16",
					ServiceNodePortRange: "1024-32767",
				},
				Etcd: EtcdConfig{
					QuotaBackendSize:        "2Gi",
					QuotaBackendBytes:       2 * 1024 * 1024 * 1024,
					MinDefragSize:           "100Mi",
					MinDefragBytes:          100 * 1024 * 1024,
					MaxFragmentedPercentage: 45,
					DefragCheckFreq:         "5m",
					DefragCheckDuration:     5 * time.Minute,
					DoStartupDefrag:         true,
				},
			},
			expectErr: false,
		},
	}
	for _, tt := range ttests {
		t.Run("", func(t *testing.T) {
			f, err := os.CreateTemp("", "test")
			if err != nil {
				t.Errorf("unable to create temp file: %v", err)
			}
			defer os.Remove(f.Name())
			d, err := yaml.Marshal(&tt.config)
			if err != nil {
				t.Errorf("unable to marshal configuration: %v", err)
			}
			_, err = f.Write(d)
			if err != nil {
				t.Errorf("unable to write to file: %v", err)
			}
			config := NewMicroshiftConfig()
			err = config.ReadFromConfigFile(f.Name())
			if tt.expectErr && err == nil {
				t.Fatal("Expecting error and received nothing")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("Not expecting error and received: %v", err)
			}
			if !tt.expectErr && !reflect.DeepEqual(*config, tt.expected) {
				t.Errorf("ReadFromConfigFile() mismatch. got=%v, want=%v", *config, tt.expected)
			}
		})
	}
}

// test the MicroshiftConfig.ReadAndValidate function to verify that it configures MicroshiftConfig from
// a configuration file.
func TestMicroshiftConfigReadAndValidate(t *testing.T) {
	cleanup := setupSuiteDataDir(t)
	defer cleanup()

	var ttests = []struct {
		name      string
		config    Config
		expected  MicroshiftConfig
		expectErr bool
	}{
		{
			name: "Config OK full",
			config: Config{
				DNS: DNS{
					BaseDomain: "example.com",
				},
				Network: Network{
					ClusterNetwork: []ClusterNetworkEntry{
						{
							CIDR: "10.20.30.40/16",
						},
					},
					ServiceNetwork:       []string{"40.30.20.10/16"},
					ServiceNodePortRange: "1024-32767",
				},
				Node: Node{
					HostnameOverride: "node1",
					NodeIP:           "1.2.3.4",
				},
				ApiServer: ApiServer{
					SubjectAltNames:  []string{"node1", "node2"},
					AdvertiseAddress: "6.7.8.9",
				},
				Debugging: Debugging{
					LogLevel: "Debug",
				},
				Etcd: EtcdConfig{
					QuotaBackendSize:        "2Gi",
					MinDefragSize:           "100Mi",
					MaxFragmentedPercentage: 45,
					DefragCheckFreq:         "5m",
					DoStartupDefrag:         true,
				},
			},
			expected: MicroshiftConfig{
				Debugging: Debugging{
					LogLevel: "Debug",
				},
				SubjectAltNames: []string{"node1", "node2"},
				Node: Node{
					HostnameOverride: "node1",
					NodeIP:           "1.2.3.4",
				},
				KASAdvertiseAddress: "6.7.8.9",
				SkipKASInterface:    true,
				DNS: DNS{
					BaseDomain: "example.com",
				},
				Cluster: ClusterConfig{
					URL:                  "https://localhost:6443",
					ClusterCIDR:          "10.20.30.40/16",
					ServiceCIDR:          "40.30.20.10/16",
					ServiceNodePortRange: "1024-32767",
					DNS:                  "40.30.0.10",
				},
				Etcd: EtcdConfig{
					QuotaBackendSize:        "2Gi",
					QuotaBackendBytes:       2 * 1024 * 1024 * 1024,
					MinDefragSize:           "100Mi",
					MinDefragBytes:          100 * 1024 * 1024,
					MaxFragmentedPercentage: 45,
					DefragCheckFreq:         "5m",
					DefragCheckDuration:     5 * time.Minute,
					DoStartupDefrag:         true,
				},
			},
			expectErr: false,
		},
		{
			name: "Config NOK with bad SAN localhost",
			config: Config{
				ApiServer: ApiServer{
					SubjectAltNames: []string{"127.0.0.1", "localhost"},
				},
			},
			expected:  MicroshiftConfig{},
			expectErr: true,
		},
		{
			name: "Config NOK with bad SAN kubernetes service",
			config: Config{
				ApiServer: ApiServer{
					SubjectAltNames: []string{"kubernetes"},
				},
			},
			expected:  MicroshiftConfig{},
			expectErr: true,
		},
	}
	for _, tt := range ttests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "test")
			if err != nil {
				t.Errorf("unable to create temp file: %v", err)
			}
			defer os.Remove(f.Name())
			d, err := yaml.Marshal(&tt.config)
			if err != nil {
				t.Errorf("unable to marshal configuration: %v", err)
			}
			_, err = f.Write(d)
			if err != nil {
				t.Errorf("unable to write to file: %v", err)
			}
			config := NewMicroshiftConfig()
			err = config.ReadAndValidate(f.Name())
			if tt.expectErr && err == nil {
				t.Fatal("Expecting error and received nothing")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("Not expecting error and received: %v", err)
			}
			if !tt.expectErr && !reflect.DeepEqual(*config, tt.expected) {
				t.Errorf("ReadAndValidate() mismatch. got=%v, want=%v", *config, tt.expected)
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
