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
	switch t.Hosts.Status {
	case HostsStatusEnabled:
		if t.Hosts.File != "" {
			cleanPath := filepath.Clean(t.Hosts.File)

			fi, err := os.Stat(cleanPath)
			// Enforce ConfigMap requirement: the file must not exceed 1MiB, as it will be mounted into a ConfigMap.
			if err == nil && fi.Size() > 1048576 {
				return fmt.Errorf("hosts file %s exceeds 1MiB ConfigMap (and internal buffer) size limit (got %d bytes)", t.Hosts.File, fi.Size())
			}
			if !filepath.IsAbs(cleanPath) {
				return fmt.Errorf("hosts file path must be absolute: got %s", t.Hosts.File)
			}

			if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
				return fmt.Errorf("hosts file %s does not exist", t.Hosts.File)
			} else if err != nil {
				return fmt.Errorf("error checking hosts file %s: %v", t.Hosts.File, err)
			}
			file, err := os.Open(t.Hosts.File)
			if err != nil {
				return fmt.Errorf("hosts file %s is not readable: %v", t.Hosts.File, err)
			}
			file.Close()
		}

	case HostsStatusDisabled:
		// valid status
	default:
		return fmt.Errorf("invalid hosts status: %s", t.Hosts.Status)
	}
	return nil
}
