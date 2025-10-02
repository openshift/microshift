package config

import (
	"fmt"
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

	// Hosts contains configuration for the hosts file watcher service.
	Hosts HostsConfig `json:"hosts,omitempty"`
}

// HostsConfig contains configuration for the hosts file watcher service.
type HostsConfig struct {
	// File is the path to the hosts file to monitor.
	// If not specified, defaults to "/etc/hosts".
	// +kubebuilder:default="/etc/hosts"
	// +kubebuilder:example="/etc/hosts"
	File string `json:"file,omitempty"`

	// Status controls whether the hosts file watcher service is enabled or disabled.
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

	if t.Hosts.Status != HostsStatusEnabled && t.Hosts.Status != HostsStatusDisabled {
		return fmt.Errorf("invalid hosts status: %s", t.Hosts.Status)
	}
	return nil
}
