package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
)

// tests to make sure that the config file is parsed correctly
func TestConfigFile(t *testing.T) {
	config := NewMicroshiftConfig()
	config.ConfigFile = "../../test/config.yaml"
	if err := config.ReadFromConfigFile(); err != nil {
		t.Errorf("failed to read config file: %v", err)
	}
	if config.DataDir != "/tmp/microshift/data" {
		t.Errorf("failed to read data dir from config file: %s", config.DataDir)
	}
	if config.LogDir != "/tmp/microshift/logs" {
		t.Errorf("failed to read log dir from config file: %s", config.LogDir)
	}
	if config.LogVLevel != 4 {
		t.Errorf("failed to read log vlevel from config file: %d", config.LogVLevel)
	}
	if config.LogVModule != "microshift=4" {
		t.Errorf("failed to read log vmodule from config file: %s", config.LogVModule)
	}
	if config.LogAlsotostderr != true {
		t.Errorf("failed to read log alsotostderr from config file: %t", config.LogAlsotostderr)
	}
	if len(config.Roles) != 2 {
		t.Errorf("failed to read roles from config file: %v", config.Roles)
	}
	if config.NodeName != "node1" {
		t.Errorf("failed to read node name from config file: %s", config.NodeName)
	}
	if config.NodeIP != "1.2.3.4" {
		t.Errorf("failed to read node ip from config file: %s", config.NodeIP)
	}
	if config.Cluster.URL != "https://1.2.3.4:6443" {
		t.Errorf("failed to read cluster url from config file: %s", config.Cluster.URL)
	}
	if config.Cluster.ClusterCIDR != "10.20.30.40/16" {
		t.Errorf("failed to read cluster cidr from config file: %s", config.Cluster.ClusterCIDR)
	}
	if config.Cluster.ServiceCIDR != "40.30.20.10/16" {
		t.Errorf("failed to read cluster service cidr from config file: %s", config.Cluster.ServiceCIDR)
	}
	if config.Cluster.DNS != "cluster.dns" {
		t.Errorf("failed to read cluster dns from config file: %s", config.Cluster.DNS)
	}
	if config.Cluster.Domain != "cluster.local" {
		t.Errorf("failed to read cluster domain from config file: %s", config.Cluster.Domain)
	}
}

// test that Microshift is able to properly read the config from the commandline
func TestCommandLineConfig(t *testing.T) {
	config := NewMicroshiftConfig()
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.StringVar(&config.DataDir, "data-dir", "", "")
	flags.StringVar(&config.LogDir, "log-dir", "", "")
	flags.IntVar(&config.LogVLevel, "v", 0, "")
	flags.StringVar(&config.LogVModule, "vmodule", "", "")
	flags.BoolVar(&config.LogAlsotostderr, "alsologtostderr", false, "")
	flags.StringSliceVar(&config.Roles, "roles", []string{}, "")
	flags.StringVar(&config.NodeName, "node-name", "", "")
	flags.StringVar(&config.NodeIP, "node-ip", "", "")
	flags.StringVar(&config.Cluster.URL, "cluster-url", "", "")
	flags.StringVar(&config.Cluster.ClusterCIDR, "cluster-cidr", "", "")
	flags.StringVar(&config.Cluster.ServiceCIDR, "service-cidr", "", "")
	flags.StringVar(&config.Cluster.DNS, "cluster-dns", "", "")
	flags.StringVar(&config.Cluster.Domain, "cluster-domain", "", "")
	flags.Parse([]string{
		"--data-dir=/tmp/microshift/data",
		"--log-dir=/tmp/microshift/logs",
		"--v=4",
		"--vmodule=microshift=4",
		"--alsologtostderr",
		"--roles=controlplane,node",
		"--node-name=node1",
		"--node-ip=1.2.3.4",
		"--cluster-url=https://1.2.3.4:6443",
		"--cluster-cidr=10.20.30.40/16",
		"--service-cidr=40.30.20.10/16",
		"--cluster-dns=10.43.0.10",
		"--cluster-domain=cluster.local",
	})

	if err := config.ReadFromCmdLine(flags); err != nil {
		t.Errorf("failed to read config from commandline: %v", err)
	}

	if config.DataDir != "/tmp/microshift/data" {
		t.Errorf("failed to read data dir from commandline: %s", config.DataDir)
	}
	if config.LogDir != "/tmp/microshift/logs" {
		t.Errorf("failed to read log dir from commandline: %s", config.LogDir)
	}
	if config.LogVLevel != 4 {
		t.Errorf("failed to read log vlevel from commandline: %d", config.LogVLevel)
	}
	if config.LogVModule != "microshift=4" {
		t.Errorf("failed to read log vmodule from commandline: %s", config.LogVModule)
	}
	if config.LogAlsotostderr != true {
		t.Errorf("failed to read log alsotostderr from commandline: %v", config.LogAlsotostderr)
	}
	if len(config.Roles) != 2 {
		t.Errorf("failed to read roles from commandline: %v", config.Roles)
	}
	if config.NodeName != "node1" {
		t.Errorf("failed to read node name from commandline: %s", config.NodeName)
	}
	if config.NodeIP != "1.2.3.4" {
		t.Errorf("failed to read node ip from commandline: %s", config.NodeIP)
	}
	if config.Cluster.URL != "https://1.2.3.4:6443" {
		t.Errorf("failed to read cluster url from commandline: %s", config.Cluster.URL)
	}
	if config.Cluster.ClusterCIDR != "10.20.30.40/16" {
		t.Errorf("failed to read cluster cidr from commandline: %s", config.Cluster.ClusterCIDR)
	}
	if config.Cluster.ServiceCIDR != "40.30.20.10/16" {
		t.Errorf("failed to read cluster service cidr from commandline: %s", config.Cluster.ServiceCIDR)
	}
	if config.Cluster.DNS != "10.43.0.10" {
		t.Errorf("failed to read cluster dns from commandline: %s", config.Cluster.DNS)
	}
	if config.Cluster.Domain != "cluster.local" {
		t.Errorf("failed to read cluster domain from commandline: %s", config.Cluster.Domain)
	}
}

// test to verify that Microshift is able to populate the config from the environment variables
func TestEnvironmentVariableConfig(t *testing.T) {
	os.Setenv("MICROSHIFT_CONFIGFILE", "/to/config/file")
	os.Setenv("MICROSHIFT_DATADIR", "/tmp/microshift/data")
	os.Setenv("MICROSHIFT_LOGDIR", "/tmp/microshift/logs")
	os.Setenv("MICROSHIFT_LOGVLEVEL", "23")
	os.Setenv("MICROSHIFT_LOGVMODULE", "microshift=23")
	os.Setenv("MICROSHIFT_LOGALSOTOSTDERR", "true")
	os.Setenv("MICROSHIFT_ROLES", "controlplane,node")
	os.Setenv("MICROSHIFT_NODENAME", "node1")
	os.Setenv("MICROSHIFT_NODEIP", "1.2.3.4")
	os.Setenv("MICROSHIFT_CLUSTER_URL", "https://cluster.com:4343/endpoint")
	os.Setenv("MICROSHIFT_CLUSTER_CLUSTERCIDR", "10.20.30.40/16")
	os.Setenv("MICROSHIFT_CLUSTER_SERVICECIDR", "40.30.20.10/16")
	os.Setenv("MICROSHIFT_CLUSTER_DNS", "10.43.0.10")
	os.Setenv("MICROSHIFT_CLUSTER_DOMAIN", "cluster.local")

	config := NewMicroshiftConfig()
	if err := config.ReadFromEnv(); err != nil {
		t.Errorf("failed to read from environment variables: %v", err)
	}

	if config.ConfigFile != "/to/config/file" {
		t.Errorf("expected ConfigFile to be empty, got %s", config.ConfigFile)
	}
	if config.DataDir != "/tmp/microshift/data" {
		t.Errorf("expected DataDir to be /tmp/microshift/data, got %s", config.DataDir)
	}
	if config.LogDir != "/tmp/microshift/logs" {
		t.Errorf("expected LogDir to be empty, got %s", config.LogDir)
	}
	if config.LogVLevel != 23 {
		t.Errorf("expected LogVLevel to be 23, got %d", config.LogVLevel)
	}
	if config.LogVModule != "microshift=23" {
		t.Errorf("expected LogVModule to be microshift=23, got %s", config.LogVModule)
	}
	if config.LogAlsotostderr != true {
		t.Errorf("expected LogAlsotostderr to be true, got %v", config.LogAlsotostderr)
	}
	if len(config.Roles) != 2 {
		t.Errorf("expected Roles to be of length 2, got %v", config.Roles)
	}
	if config.NodeName != "node1" {
		t.Errorf("expected NodeName to not be node1, got %s", config.NodeName)
	}
	if config.NodeIP != "1.2.3.4" {
		t.Errorf("expected NodeIP to be 1.2.3.4, got %s", config.NodeIP)
	}
	if config.Cluster.URL != "https://cluster.com:4343/endpoint" {
		t.Errorf("expected Cluster.URL to be https://cluster.com:4343/endpoint, got %s", config.Cluster.URL)
	}
	if config.Cluster.ClusterCIDR != "10.20.30.40/16" {
		t.Errorf("expected Cluster.ClusterCIDR to be 10.20.30.40/16, got %s", config.Cluster.ClusterCIDR)
	}
	if config.Cluster.ServiceCIDR != "40.30.20.10/16" {
		t.Errorf("expected Cluster.ServiceCIDR to be 40.30.20.10/16, got %s", config.Cluster.ServiceCIDR)
	}
	if config.Cluster.DNS != "10.43.0.10" {
		t.Errorf("expected Cluster.DNS to be 10.43.0.10, got %s", config.Cluster.DNS)
	}
	if config.Cluster.Domain != "cluster.local" {
		t.Errorf("expected Cluster.Domain to be cluster.local, got %s", config.Cluster.Domain)
	}
}

// test the MicroshiftConfig.ReadAndValidate function to verify that it configures MicroshiftConfig with a valid flagset
func TestMicroshiftConfigReadAndValidate(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("data-dir", "", "")
	flags.String("log-dir", "", "")
	flags.Int("v", 0, "")
	flags.String("vmodule", "", "")
	flags.Bool("alsologtostderr", false, "")
	flags.StringSlice("roles", []string{}, "")

	c := NewMicroshiftConfig()
	if err := c.ReadAndValidate(flags); err != nil {
		t.Errorf("failed to read and validate config: %v", err)
	}
}

// tests that the global flags have been initialized
func TestGlobalInitFlags(t *testing.T) {
	InitGlobalFlags()
	// pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()
}

// tests that the default config file is being read
func TestDefaultConfigFile(t *testing.T) {
	os.Setenv("MICROSHIFT_CONFIG_FILE", "")
	config := NewMicroshiftConfig()
	if config.ConfigFile != "" {
		t.Errorf("expected config file to be empty, got %s", config.ConfigFile)
	}
}
