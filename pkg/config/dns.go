package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	HostsStatusEnabled  HostsStatusEnum = "Enabled"
	HostsStatusDisabled HostsStatusEnum = "Disabled"
)

type HostsStatusEnum string

type DNS struct {
	// baseDomain is the base domain of the cluster. All managed DNS records will
	// be sub-domains of this base.
	//
	// For example, given the base domain `example.com`, router exposed
	// domains will be formed as `*.apps.example.com` by default,
	// and API service will have a DNS entry for `api.example.com`,
	// as well as "api-int.example.com" for internal k8s API access.
	//
	// Once set, this field cannot be changed.
	// +kubebuilder:default=example.com
	// +kubebuilder:example=microshift.example.com
	BaseDomain string `json:"baseDomain"`

	// configFile is the path to a custom CoreDNS Corefile on the host filesystem.
	// When set, MicroShift uses this file as the Corefile in the dns-default ConfigMap,
	// fully replacing the default template-rendered configuration.
	// Changes to this file are detected at runtime and applied without restarting MicroShift.
	// Mutually exclusive with dns.hosts: setting both causes a startup error.
	// +optional
	// +kubebuilder:example="/etc/microshift/dns/Corefile"
	ConfigFile string `json:"configFile,omitempty"`

	// Hosts contains configuration for the hosts file.
	Hosts HostsConfig `json:"hosts,omitempty"`
}

// HostsConfig contains configuration for the hosts file .
type HostsConfig struct {
	// File is the path to the hosts file to monitor.
	// If not specified, defaults to "/etc/hosts".
	// +kubebuilder:default="/etc/hosts"
	// +kubebuilder:example="/etc/hosts"
	File string `json:"file,omitempty"`

	// Status controls whether the hosts file is enabled or disabled.
	// Allowed values are "Enabled" and "Disabled".
	// If not specified, defaults to "Disabled".
	// +kubebuilder:default="Disabled"
	// +kubebuilder:example="Enabled"
	// +kubebuilder:validation:Enum=Enabled;Disabled
	Status HostsStatusEnum `json:"status,omitempty"`
}

func dnsDefaults() DNS {
	return DNS{
		BaseDomain: "example.com",
		Hosts: HostsConfig{
			File:   "/etc/hosts",
			Status: HostsStatusDisabled,
		},
	}
}

func (t *DNS) validate() error {
	if t.ConfigFile != "" && t.Hosts.Status == HostsStatusEnabled {
		return fmt.Errorf("dns.configFile and dns.hosts are mutually exclusive")
	}

	if err := t.validateConfigFile(); err != nil {
		return err
	}

	return t.validateHosts()
}

func (t *DNS) validateConfigFile() error {
	if t.ConfigFile == "" {
		return nil
	}

	cleanPath := filepath.Clean(t.ConfigFile)
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("dns config file path must be absolute: got %s", t.ConfigFile)
	}

	fi, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("dns config file %s does not exist", t.ConfigFile)
	} else if err != nil {
		return fmt.Errorf("error checking dns config file %s: %v", t.ConfigFile, err)
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("dns config file %s must be a regular file", t.ConfigFile)
	}

	if fi.Size() == 0 {
		return fmt.Errorf("dns config file %s is empty", t.ConfigFile)
	}

	if fi.Size() > 1048576 {
		return fmt.Errorf("dns config file %s exceeds 1MiB ConfigMap size limit (got %d bytes)", t.ConfigFile, fi.Size())
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("dns config file %s is not readable: %v", t.ConfigFile, err)
	}
	return file.Close()
}

func (t *DNS) validateHosts() error {
	switch t.Hosts.Status {
	case HostsStatusEnabled:
		if t.Hosts.File == "" {
			break
		}

		cleanPath := filepath.Clean(t.Hosts.File)

		fi, err := os.Stat(cleanPath)
		if err == nil && fi.Size() > 1048576 {
			return fmt.Errorf("hosts file %s exceeds 1MiB ConfigMap (and internal buffer) size limit (got %d bytes)", t.Hosts.File, fi.Size())
		}
		if !filepath.IsAbs(cleanPath) {
			return fmt.Errorf("hosts file path must be absolute: got %s", t.Hosts.File)
		}

		_, err = os.Stat(cleanPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("hosts file %s does not exist", t.Hosts.File)
		} else if err != nil {
			return fmt.Errorf("error checking hosts file %s: %v", t.Hosts.File, err)
		}

		file, err := os.Open(t.Hosts.File)
		if err != nil {
			return fmt.Errorf("hosts file %s is not readable: %v", t.Hosts.File, err)
		}
		return file.Close()

	case HostsStatusDisabled:
		return nil
	default:
		return fmt.Errorf("invalid hosts status: %s", t.Hosts.Status)
	}
	return nil
}
