package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNS_ValidateConfigFile_MutualExclusivity(t *testing.T) {
	tmpFile := createTempCorefile(t, ".:5353 { whoami }")

	dns := DNS{
		ConfigFile: tmpFile,
		Hosts: HostsConfig{
			Status: HostsStatusEnabled,
			File:   "/etc/hosts",
		},
	}
	err := dns.validate()
	assert.ErrorContains(t, err, "dns.configFile and dns.hosts are mutually exclusive")
}

func TestDNS_ValidateConfigFile_WithHostsDisabled(t *testing.T) {
	tmpFile := createTempCorefile(t, ".:5353 { whoami }")

	dns := DNS{
		ConfigFile: tmpFile,
		Hosts: HostsConfig{
			Status: HostsStatusDisabled,
		},
	}
	assert.NoError(t, dns.validate())
}

func TestDNS_ValidateConfigFile_EmptyConfigFilePreservesDefault(t *testing.T) {
	dns := DNS{
		ConfigFile: "",
		Hosts: HostsConfig{
			Status: HostsStatusDisabled,
		},
	}
	assert.NoError(t, dns.validate())
}

func TestDNS_ValidateConfigFile_NonAbsolutePath(t *testing.T) {
	dns := DNS{
		ConfigFile: "relative/path/Corefile",
		Hosts: HostsConfig{
			Status: HostsStatusDisabled,
		},
	}
	err := dns.validate()
	assert.ErrorContains(t, err, "dns config file path must be absolute")
}

func TestDNS_ValidateConfigFile_NonExistentFile(t *testing.T) {
	dns := DNS{
		ConfigFile: "/tmp/nonexistent-corefile-test-12345",
		Hosts: HostsConfig{
			Status: HostsStatusDisabled,
		},
	}
	err := dns.validate()
	assert.ErrorContains(t, err, "does not exist")
}

func TestDNS_ValidateConfigFile_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	dns := DNS{
		ConfigFile: tmpDir,
		Hosts: HostsConfig{
			Status: HostsStatusDisabled,
		},
	}
	err := dns.validate()
	assert.ErrorContains(t, err, "must be a regular file")
}

func TestDNS_ValidateConfigFile_EmptyFile(t *testing.T) {
	tmpFile := createTempCorefile(t, "")

	dns := DNS{
		ConfigFile: tmpFile,
		Hosts: HostsConfig{
			Status: HostsStatusDisabled,
		},
	}
	err := dns.validate()
	assert.ErrorContains(t, err, "is empty")
}

func TestDNS_ValidateConfigFile_ExceedsSizeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Corefile")
	// 1 MiB + 1 byte
	data := make([]byte, 1048576+1)
	require.NoError(t, os.WriteFile(tmpFile, data, 0644))

	dns := DNS{
		ConfigFile: tmpFile,
		Hosts: HostsConfig{
			Status: HostsStatusDisabled,
		},
	}
	err := dns.validate()
	assert.ErrorContains(t, err, "exceeds 1MiB ConfigMap size limit")
}

func TestDNS_ValidateConfigFile_ValidFile(t *testing.T) {
	tmpFile := createTempCorefile(t, ".:5353 {\n    whoami\n    reload\n}\n")

	dns := DNS{
		ConfigFile: tmpFile,
		Hosts: HostsConfig{
			Status: HostsStatusDisabled,
		},
	}
	assert.NoError(t, dns.validate())
}

func TestDNS_ConfigFile_IncorporatedFromDropIn(t *testing.T) {
	tmpFile := createTempCorefile(t, ".:5353 {\n    whoami\n    reload\n}\n")

	yamlConfig := fmt.Sprintf("dns:\n  configFile: %s\n", tmpFile)
	config, err := getActiveConfigFromYAMLDropins([][]byte{[]byte(yamlConfig)})
	require.NoError(t, err)
	assert.Equal(t, tmpFile, config.DNS.ConfigFile)
}

func createTempCorefile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Corefile")
	require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0644))
	return tmpFile
}
