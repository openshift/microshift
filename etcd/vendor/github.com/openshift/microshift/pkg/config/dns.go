package config

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	HostsStatusEnabled  HostsStatusEnum = "Enabled"
	HostsStatusDisabled HostsStatusEnum = "Disabled"
)

type HostsStatusEnum string

// DNSResources configures the CPU and memory resources for the dns container
// in the dns-default DaemonSet.
type DNSResources struct {
	// Requests specifies the minimum resources required for the dns container.
	// Valid keys are "cpu" and "memory". Values must be valid Kubernetes resource quantities.
	// When not set, defaults to cpu=50m, memory=70Mi.
	Requests map[string]string `json:"requests,omitempty"`

	// Limits specifies the maximum resources the dns container can use.
	// Valid keys are "cpu" and "memory". Values must be valid Kubernetes resource quantities.
	// When not set, no limits are applied.
	Limits map[string]string `json:"limits,omitempty"`
}

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

	// Resources configures the CPU and memory resources for the dns container.
	Resources DNSResources `json:"resources,omitempty"`
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
		Resources: DNSResources{
			Requests: map[string]string{
				"cpu":    "50m",
				"memory": "70Mi",
			},
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

	if err := t.validateHosts(); err != nil {
		return err
	}
	return t.validateResources()
}

func (t *DNS) validateConfigFile() error {
	if t.ConfigFile == "" {
		return nil
	}
	return validateFilePath(t.ConfigFile, "dns config file")
}

func (t *DNS) validateHosts() error {
	switch t.Hosts.Status {
	case HostsStatusEnabled:
		if t.Hosts.File == "" {
			return nil
		}
		return validateFilePath(t.Hosts.File, "hosts file")
	case HostsStatusDisabled:
		return nil
	default:
		return fmt.Errorf("invalid hosts status: %s", t.Hosts.Status)
	}
}

func dnsMinimumRequests() map[string]resource.Quantity {
	defaults := dnsDefaults()
	mins := make(map[string]resource.Quantity, len(defaults.Resources.Requests))
	for k, v := range defaults.Resources.Requests {
		mins[k] = resource.MustParse(v)
	}
	return mins
}

func (t *DNS) validateResources() error {
	allowed := map[string]struct{}{
		"cpu":    {},
		"memory": {},
	}
	mins := dnsMinimumRequests()
	for key, val := range t.Resources.Requests {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("unsupported dns resource request key %q: allowed keys are cpu, memory", key)
		}
		qty, err := resource.ParseQuantity(val)
		if err != nil {
			return fmt.Errorf("invalid dns resource request %s=%q: %v", key, val, err)
		}
		if minQty, ok := mins[key]; ok && qty.Cmp(minQty) < 0 {
			return fmt.Errorf("dns resource request %s=%q is below minimum %s", key, val, minQty.String())
		}
	}
	for key, val := range t.Resources.Limits {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("unsupported dns resource limit key %q: allowed keys are cpu, memory", key)
		}
		if _, err := resource.ParseQuantity(val); err != nil {
			return fmt.Errorf("invalid dns resource limit %s=%q: %v", key, val, err)
		}
	}
	for key, limitVal := range t.Resources.Limits {
		reqVal, ok := t.Resources.Requests[key]
		if !ok {
			continue
		}
		limit := resource.MustParse(limitVal)
		req := resource.MustParse(reqVal)
		if limit.Cmp(req) < 0 {
			return fmt.Errorf("dns resource limit %s=%q must be greater than or equal to request %s=%q", key, limitVal, key, reqVal)
		}
	}
	return nil
}

func validateFilePath(path, label string) error {
	cleanPath := filepath.Clean(path)
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("%s path must be absolute: got %s", label, path)
	}

	fi, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s %s does not exist", label, path)
	} else if err != nil {
		return fmt.Errorf("error checking %s %s: %v", label, path, err)
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("%s %s must be a regular file", label, path)
	}

	if fi.Size() == 0 {
		return fmt.Errorf("%s %s is empty", label, path)
	}

	if fi.Size() > 1048576 {
		return fmt.Errorf("%s %s exceeds 1MiB size limit (got %d bytes)", label, path, fi.Size())
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("%s %s is not readable: %v", label, path, err)
	}
	return file.Close()
}
