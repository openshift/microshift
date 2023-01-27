package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

const (
	testConfigFile                   = "../../test/config.yaml"
	testConfigFileBadSubjectAltNames = "../../test/config_bad_subjectaltnames.yaml"
	IS_DEFAULT_NODENAME              = true
	IS_NOT_DEFAULT_NODENAME          = false
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

// tests to make sure that the config file is parsed correctly
func TestConfigFile(t *testing.T) {

	var ttests = []struct {
		configFile string
		err        error
	}{
		{testConfigFile, nil},
		{testConfigFileBadSubjectAltNames, nil},
	}

	for _, tt := range ttests {
		config := NewMicroshiftConfig()
		err := config.ReadFromConfigFile(tt.configFile)
		if (err != nil) != (tt.err != nil) {
			t.Errorf("ReadFromConfigFile() error = %v, wantErr %v", err, tt.err)
		}
	}
}

// test that MicroShift is able to properly read the config from the commandline
func TestCommandLineConfig(t *testing.T) {

	var ttests = []struct {
		config *MicroshiftConfig
		err    error
	}{
		{
			config: &MicroshiftConfig{
				LogVLevel:       4,
				SubjectAltNames: []string{"node1"},
				NodeName:        "node1",
				NodeIP:          "1.2.3.4",
				BaseDomain:      "example.com",
				Cluster: ClusterConfig{
					URL:                  "https://127.0.0.1:6443",
					ClusterCIDR:          "10.20.30.40/16",
					ServiceCIDR:          "40.30.20.10/16",
					ServiceNodePortRange: "1024-32767",
				},
			},
			err: nil,
		},
	}

	for _, tt := range ttests {
		config := NewMicroshiftConfig()

		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
		// all other flags unbound (looked up by name) and defaulted
		flags.Int("v", config.LogVLevel, "")
		flags.StringSlice("subject-alt-names", config.SubjectAltNames, "")
		flags.String("hostname-override", config.NodeName, "")
		flags.String("node-ip", config.NodeIP, "")
		flags.String("cluster-cidr", config.Cluster.ClusterCIDR, "")
		flags.String("service-cidr", config.Cluster.ServiceCIDR, "")
		flags.String("service-node-port-range", config.Cluster.ServiceNodePortRange, "")
		flags.String("base-domain", config.BaseDomain, "")

		// parse the flags
		var err error
		err = flags.Parse([]string{
			"--v=" + strconv.Itoa(tt.config.LogVLevel),
			"--subject-alt-names=" + strings.Join(tt.config.SubjectAltNames, ","),
			"--hostname-override=" + tt.config.NodeName,
			"--node-ip=" + tt.config.NodeIP,
			"--cluster-cidr=" + tt.config.Cluster.ClusterCIDR,
			"--service-cidr=" + tt.config.Cluster.ServiceCIDR,
			"--service-node-port-range=" + tt.config.Cluster.ServiceNodePortRange,
			"--base-domain=" + tt.config.BaseDomain,
		})
		if err != nil {
			t.Errorf("failed to parse command line flags: %s", err)
		}

		// validate that we can read the config from the commandline
		err = config.ReadFromCmdLine(flags)
		if (err != nil) != (tt.err != nil) {
			t.Errorf("failed to read config from commandline: %s", err)
		}
		if err == nil && !reflect.DeepEqual(config, tt.config) {
			t.Errorf("struct read from commandline does not match: expected %+v, got %+v", tt.config, config)
		}
	}
}

// test to verify that MicroShift is able to populate the config from the environment variables
func TestEnvironmentVariableConfig(t *testing.T) {
	// set up the table tests using the above environment variables & the MicroShift config struct
	var ttests = []struct {
		desiredMicroShiftConfig *MicroshiftConfig
		err                     error
		envList                 []struct {
			varName string
			value   string
		}
	}{
		{
			desiredMicroShiftConfig: &MicroshiftConfig{
				LogVLevel:       23,
				SubjectAltNames: []string{"node1", "node2"},
				NodeName:        "node1",
				NodeIP:          "1.2.3.4",
				BaseDomain:      "example.com",
				Cluster: ClusterConfig{
					URL:                  "https://cluster.com:4343/endpoint",
					ClusterCIDR:          "10.20.30.40/16",
					ServiceCIDR:          "40.30.20.10/16",
					ServiceNodePortRange: "1024-32767",
				},
			},
			err: nil,
			envList: []struct {
				varName string
				value   string
			}{
				{"MICROSHIFT_LOGVLEVEL", "23"},
				{"MICROSHIFT_NODENAME", "node1"},
				{"MICROSHIFT_SUBJECTALTNAMES", "node1,node2"},
				{"MICROSHIFT_NODEIP", "1.2.3.4"},
				{"MICROSHIFT_BASEDOMAIN", "example.com"},
				{"MICROSHIFT_CLUSTER_URL", "https://cluster.com:4343/endpoint"},
				{"MICROSHIFT_CLUSTER_CLUSTERCIDR", "10.20.30.40/16"},
				{"MICROSHIFT_CLUSTER_SERVICECIDR", "40.30.20.10/16"},
				{"MICROSHIFT_CLUSTER_SERVICENODEPORTRANGE", "1024-32767"},
			},
		},
		{
			desiredMicroShiftConfig: &MicroshiftConfig{
				LogVLevel:       23,
				SubjectAltNames: []string{"node1"},
				NodeName:        "node1",
				NodeIP:          "1.2.3.4",
				BaseDomain:      "another.example.com",
				Cluster: ClusterConfig{
					URL:                  "https://cluster.com:4343/endpoint",
					ClusterCIDR:          "10.20.30.40/16",
					ServiceCIDR:          "40.30.20.10/16",
					ServiceNodePortRange: "1024-32767",
				},
			},
			err: nil,
			envList: []struct {
				varName string
				value   string
			}{
				{"MICROSHIFT_LOGVLEVEL", "23"},
				{"MICROSHIFT_NODENAME", "node1"},
				{"MICROSHIFT_SUBJECTALTNAMES", "node1"},
				{"MICROSHIFT_NODEIP", "1.2.3.4"},
				{"MICROSHIFT_BASEDOMAIN", "another.example.com"},
				{"MICROSHIFT_CLUSTER_URL", "https://cluster.com:4343/endpoint"},
				{"MICROSHIFT_CLUSTER_CLUSTERCIDR", "10.20.30.40/16"},
				{"MICROSHIFT_CLUSTER_SERVICECIDR", "40.30.20.10/16"},
				{"MICROSHIFT_CLUSTER_SERVICENODEPORTRANGE", "1024-32767"},
			},
		},
	}

	for _, tt := range ttests {
		// first set the values
		for _, env := range tt.envList {
			os.Setenv(env.varName, env.value)
			defer os.Unsetenv(env.varName)
		}
		// then read the values
		microShiftconfig := NewMicroshiftConfig()
		err := microShiftconfig.ReadFromEnv()
		if (err != nil && tt.err == nil) || (err == nil && tt.err != nil) {
			t.Errorf("failed to read from env, expected error: %v, got: %v", tt.err, err)
		}
		if (err == nil && !reflect.DeepEqual(microShiftconfig, tt.desiredMicroShiftConfig)) ||
			(err != nil && reflect.DeepEqual(microShiftconfig, tt.desiredMicroShiftConfig)) {
			t.Errorf("structs don't match up, expected: %+v, got: %+v", tt.desiredMicroShiftConfig, microShiftconfig)
		}
	}
}

// test the MicroshiftConfig.ReadAndValidate function to verify that it configures MicroshiftConfig with a valid flagset
func TestMicroshiftConfigReadAndValidate(t *testing.T) {
	cleanup := setupSuiteDataDir(t)
	defer cleanup()

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.Int("v", 0, "")

	var ttests = []struct {
		configFile string
		expectErr  bool
	}{
		{
			configFile: testConfigFile,
			expectErr:  false,
		},
		{
			configFile: testConfigFileBadSubjectAltNames,
			expectErr:  true,
		},
	}
	for _, tt := range ttests {
		microShiftConfig := NewMicroshiftConfig()
		err := microShiftConfig.ReadAndValidate(tt.configFile, flags)
		if tt.expectErr && err == nil {
			t.Error("Expecting error and received nothing")
		}
		if !tt.expectErr && err != nil {
			t.Errorf("Not expecting error and received: %v", err)
		}
	}
}

func TestMicroshiftConfigIsDefaultNodeName(t *testing.T) {
	c := NewMicroshiftConfig()
	if !c.isDefaultNodeName() {
		t.Errorf("expected default IsDefaultNodeName to be true")
	}

	c.NodeName += "-suffix"
	if c.isDefaultNodeName() {
		t.Errorf("expected default IsDefaultNodeName to be false")
	}
}

func TestMicroshiftConfigNodeNameValidation(t *testing.T) {
	cleanup := setupSuiteDataDir(t)
	defer cleanup()

	c := NewMicroshiftConfig()
	c.NodeName = "node1"

	if err := c.validateNodeName(IS_NOT_DEFAULT_NODENAME); err != nil {
		t.Errorf("failed to validate node name on first call: %v", err)
	}

	nodeNameFile := filepath.Join(dataDir, ".nodename")
	if data, err := os.ReadFile(nodeNameFile); err != nil {
		t.Errorf("failed to read node name from file %q: %v", nodeNameFile, err)
	} else if string(data) != c.NodeName {
		t.Errorf("node name file doesn't match the node name in the saved file: %v", err)
	}

	if err := c.validateNodeName(IS_NOT_DEFAULT_NODENAME); err != nil {
		t.Errorf("failed to validate node name on second call without changes: %v", err)
	}

	c.NodeName = "node2"
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

	c.NodeName = "node2"
	if err := c.validateNodeName(IS_DEFAULT_NODENAME); err != nil {
		t.Errorf("validation should have failed in this case, it must be a warning in logs: %v", err)
	}
}

func TestMicroshiftConfigNodeNameValidationBadName(t *testing.T) {
	cleanup := setupSuiteDataDir(t)
	defer cleanup()

	c := NewMicroshiftConfig()
	c.NodeName = "1.2.3.4"

	if err := c.validateNodeName(IS_DEFAULT_NODENAME); err == nil {
		t.Errorf("failed to validate node name.")
	}
}

// tests that the global flags have been initialized
func TestHideUnsupportedFlags(t *testing.T) {
	flags := pflag.NewFlagSet("test-flags", pflag.ContinueOnError)

	flags.String("url", "", "version usage")
	flags.String("v", "10", "v usage")
	flags.String("log_dir", "/tmp", "log_dir usage")
	flags.String("version", "", "version usage")

	HideUnsupportedFlags(flags)

	if flags.Lookup("url").Hidden {
		t.Errorf("url should not be hidden")
	}
	if flags.Lookup("v").Hidden {
		t.Errorf("v should not be hidden")
	}
	if !flags.Lookup("version").Hidden {
		t.Errorf("version should be hidden")
	}
	if !flags.Lookup("log_dir").Hidden {
		t.Errorf("log_dir should be hidden")
	}
}
