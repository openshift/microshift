package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// tests to make sure that the config file is parsed correctly
func TestConfigFile(t *testing.T) {

	var ttests = []struct {
		configFile string
		err        error
	}{
		{"../../test/config.yaml", nil},
	}

	for _, tt := range ttests {
		config := NewMicroshiftConfig()
		config.ConfigFile = tt.configFile
		err := config.ReadFromConfigFile()
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
				ConfigFile:  "/path/to/config.yaml",
				DataDir:     "/tmp/microshift/data",
				AuditLogDir: "/tmp/microshift/logs",
				LogVLevel:   4,
				Roles:       []string{"controlplane", "node"},
				NodeName:    "node1",
				NodeIP:      "1.2.3.4",
				Cluster: ClusterConfig{
					URL:                  "https://1.2.3.4:6443",
					ClusterCIDR:          "10.20.30.40/16",
					ServiceCIDR:          "40.30.20.10/16",
					ServiceNodePortRange: "1024-32767",
					DNS:                  "cluster.dns",
					Domain:               "cluster.local",
					MTU:                  "1200",
				},
				Manifests: []string{defaultManifestDirLib, defaultManifestDirEtc, "/tmp/microshift/data/manifests"},
				Debug: DebugConfig{
					Pprof: true,
				},
			},
			err: nil,
		},
	}

	for _, tt := range ttests {
		config := NewMicroshiftConfig()

		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
		// bind the config file flag to the struct
		flags.StringVar(&config.ConfigFile, "config", config.ConfigFile, "File to read configuration from.")
		// all other flags unbound (looked up by name) and defaulted
		flags.String("data-dir", config.DataDir, "")
		flags.String("audit-log-dir", config.AuditLogDir, "")
		flags.Int("v", config.LogVLevel, "")
		flags.StringSlice("roles", config.Roles, "")
		flags.String("node-name", config.NodeName, "")
		flags.String("node-ip", config.NodeIP, "")
		flags.String("url", config.Cluster.URL, "")
		flags.String("cluster-cidr", config.Cluster.ClusterCIDR, "")
		flags.String("service-cidr", config.Cluster.ServiceCIDR, "")
		flags.String("service-node-port-range", config.Cluster.ServiceNodePortRange, "")
		flags.String("cluster-dns", config.Cluster.DNS, "")
		flags.String("cluster-domain", config.Cluster.Domain, "")
		flags.String("cluster-mtu", config.Cluster.MTU, "")
		flags.Bool("debug.pprof", false, "")

		// parse the flags
		var err error
		err = flags.Parse([]string{
			"--config=" + tt.config.ConfigFile,
			"--data-dir=" + tt.config.DataDir,
			"--audit-log-dir=" + tt.config.AuditLogDir,
			"--v=" + strconv.Itoa(tt.config.LogVLevel),
			"--roles=" + strings.Join(tt.config.Roles, ","),
			"--node-name=" + tt.config.NodeName,
			"--node-ip=" + tt.config.NodeIP,
			"--url=" + tt.config.Cluster.URL,
			"--cluster-cidr=" + tt.config.Cluster.ClusterCIDR,
			"--service-cidr=" + tt.config.Cluster.ServiceCIDR,
			"--service-node-port-range=" + tt.config.Cluster.ServiceNodePortRange,
			"--cluster-dns=" + tt.config.Cluster.DNS,
			"--cluster-domain=" + tt.config.Cluster.Domain,
			"--cluster-mtu=" + tt.config.Cluster.MTU,
			"--debug.pprof=" + strconv.FormatBool(tt.config.Debug.Pprof),
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
				ConfigFile:  "/to/config/file",
				DataDir:     "/tmp/microshift/data",
				AuditLogDir: "/tmp/microshift/logs",
				LogVLevel:   23,
				Roles:       []string{"controlplane", "node"},
				NodeName:    "node1",
				NodeIP:      "1.2.3.4",
				Cluster: ClusterConfig{
					URL:                  "https://cluster.com:4343/endpoint",
					ClusterCIDR:          "10.20.30.40/16",
					ServiceCIDR:          "40.30.20.10/16",
					ServiceNodePortRange: "1024-32767",
					DNS:                  "10.43.0.10",
					Domain:               "cluster.local",
					MTU:                  "1400",
				},
				Manifests: []string{defaultManifestDirLib, defaultManifestDirEtc, "/tmp/microshift/data/manifests"},
			},
			err: nil,
			envList: []struct {
				varName string
				value   string
			}{
				{"MICROSHIFT_CONFIGFILE", "/to/config/file"},
				{"MICROSHIFT_DATADIR", "/tmp/microshift/data"},
				{"MICROSHIFT_AUDITLOGDIR", "/tmp/microshift/logs"},
				{"MICROSHIFT_LOGVLEVEL", "23"},
				{"MICROSHIFT_ROLES", "controlplane,node"},
				{"MICROSHIFT_NODENAME", "node1"},
				{"MICROSHIFT_NODEIP", "1.2.3.4"},
				{"MICROSHIFT_CLUSTER_URL", "https://cluster.com:4343/endpoint"},
				{"MICROSHIFT_CLUSTER_CLUSTERCIDR", "10.20.30.40/16"},
				{"MICROSHIFT_CLUSTER_SERVICECIDR", "40.30.20.10/16"},
				{"MICROSHIFT_CLUSTER_SERVICENODEPORTRANGE", "1024-32767"},
				{"MICROSHIFT_CLUSTER_DNS", "10.43.0.10"},
				{"MICROSHIFT_CLUSTER_DOMAIN", "cluster.local"},
			},
		},
		{
			desiredMicroShiftConfig: &MicroshiftConfig{
				ConfigFile:  "/to/config/file",
				DataDir:     "/tmp/microshift/data",
				AuditLogDir: "/tmp/microshift/logs",
				LogVLevel:   23,
				Roles:       []string{"controlplane", "node"},
				NodeName:    "node1",
				NodeIP:      "1.2.3.4",
				Cluster: ClusterConfig{
					URL:                  "https://cluster.com:4343/endpoint",
					ClusterCIDR:          "10.20.30.40/16",
					ServiceCIDR:          "40.30.20.10/16",
					ServiceNodePortRange: "1024-32767",
					DNS:                  "10.43.0.10",
					Domain:               "cluster.local",
					MTU:                  "1300",
				},
				Manifests: []string{"/my/manifests1", "/my/manifests2"},
			},
			err: nil,
			envList: []struct {
				varName string
				value   string
			}{
				{"MICROSHIFT_CONFIGFILE", "/to/config/file"},
				{"MICROSHIFT_DATADIR", "/tmp/microshift/data"},
				{"MICROSHIFT_AUDITLOGDIR", "/tmp/microshift/logs"},
				{"MICROSHIFT_LOGVLEVEL", "23"},
				{"MICROSHIFT_ROLES", "controlplane,node"},
				{"MICROSHIFT_NODENAME", "node1"},
				{"MICROSHIFT_NODEIP", "1.2.3.4"},
				{"MICROSHIFT_CLUSTER_URL", "https://cluster.com:4343/endpoint"},
				{"MICROSHIFT_CLUSTER_CLUSTERCIDR", "10.20.30.40/16"},
				{"MICROSHIFT_CLUSTER_SERVICECIDR", "40.30.20.10/16"},
				{"MICROSHIFT_CLUSTER_SERVICENODEPORTRANGE", "1024-32767"},
				{"MICROSHIFT_CLUSTER_DNS", "10.43.0.10"},
				{"MICROSHIFT_MANIFESTS", "/my/manifests1,/my/manifests2"},
				{"MICROSHIFT_CLUSTER_MTU", "1300"},
			},
		},
	}

	for _, tt := range ttests {
		// first set the values
		for _, env := range tt.envList {
			os.Setenv(env.varName, env.value)
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
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("data-dir", "", "")
	flags.String("audit-log-dir", "", "")
	flags.Int("v", 0, "")
	flags.StringSlice("roles", []string{}, "")

	c := NewMicroshiftConfig()
	if err := c.ReadAndValidate(flags); err != nil {
		t.Errorf("failed to read and validate config: %v", err)
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
		t.Errorf("v should not be hidden")
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
