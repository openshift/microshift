package config

import (
	"fmt"
	"maps"
)

const (
	// These are configuration key names, not credentials.
	kubeletImageCredentialProviderConfigPathKey = "imageCredentialProviderConfigPath" //nolint:gosec // G101: not a credential
	kubeletImageCredentialProviderBinDirKey     = "imageCredentialProviderBinDir"     //nolint:gosec // G101: not a credential
)

// kubeletReservedKeys lists the keys under the kubelet: section that MicroShift
// consumes itself (as kubelet startup flags) instead of passing through into the
// generated KubeletConfiguration.
var kubeletReservedKeys = []string{
	kubeletImageCredentialProviderConfigPathKey,
	kubeletImageCredentialProviderBinDirKey,
}

// KubeletImageCredentialProviderRawPaths returns the raw (as-configured)
// credential-provider paths read from the kubelet map, erroring only if a value
// is present but not a string. They are the input to kubeletcredential.Validate,
// which the kubelet component runs at launch (see pkg/node) to resolve and
// validate them before handing the canonical paths to kubelet.
//
// The semantic and filesystem validation lives in pkg/config/kubeletcredential,
// deliberately outside this package: it needs the kubelet config scheme, whose
// transitive imports pull in the apiserver etcd client, and pkg/config must stay
// free of that so binaries that only read configuration (notably microshift-etcd)
// do not link it.
func (c *Config) KubeletImageCredentialProviderRawPaths() (configPath, binDir string, err error) {
	configPath, err = kubeletStringValue(c.Kubelet, kubeletImageCredentialProviderConfigPathKey)
	if err != nil {
		return "", "", err
	}
	binDir, err = kubeletStringValue(c.Kubelet, kubeletImageCredentialProviderBinDirKey)
	if err != nil {
		return "", "", err
	}
	return configPath, binDir, nil
}

// kubeletStringValue reads key from the kubelet map. A missing key, an explicit
// null, or an empty string all mean "unset"; a non-string value is an error.
func kubeletStringValue(m map[string]any, key string) (string, error) {
	if m == nil {
		return "", nil
	}
	raw, ok := m[key]
	if !ok || raw == nil {
		return "", nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("kubelet.%s must be a string, got %T", key, raw)
	}
	return s, nil
}

// KubeletPassthrough returns a copy of the kubelet map with the MicroShift-owned
// keys removed, so only genuine KubeletConfiguration settings are written to the
// generated kubelet config file. A nil map returns nil.
func (c *Config) KubeletPassthrough() map[string]any {
	if c.Kubelet == nil {
		return nil
	}
	out := maps.Clone(c.Kubelet)
	for _, k := range kubeletReservedKeys {
		delete(out, k)
	}
	return out
}
