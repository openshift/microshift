package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubeletImageCredentialProviderRawPaths(t *testing.T) {
	ttests := []struct {
		name        string
		kubelet     map[string]any
		wantConfig  string
		wantBinDir  string
		expectError string
	}{
		{
			name:    "nil map",
			kubelet: nil,
		},
		{
			name:    "keys absent",
			kubelet: map[string]any{"cpuManagerPolicy": "static"},
		},
		{
			name: "both present",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": "/etc/microshift/cp.yaml",
				"imageCredentialProviderBinDir":     "/usr/libexec/cp",
			},
			wantConfig: "/etc/microshift/cp.yaml",
			wantBinDir: "/usr/libexec/cp",
		},
		{
			name: "only config present",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": "/etc/microshift/cp.yaml",
			},
			wantConfig: "/etc/microshift/cp.yaml",
		},
		{
			name: "empty string is unset",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": "",
				"imageCredentialProviderBinDir":     "",
			},
		},
		{
			name: "explicit null is unset",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": nil,
				"imageCredentialProviderBinDir":     nil,
			},
		},
		{
			name: "int type is rejected",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": 42,
			},
			expectError: "kubelet.imageCredentialProviderConfigPath must be a string, got int",
		},
		{
			name: "bool type is rejected",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": "/etc/microshift/cp.yaml",
				"imageCredentialProviderBinDir":     true,
			},
			expectError: "kubelet.imageCredentialProviderBinDir must be a string, got bool",
		},
		{
			name: "map type is rejected",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": map[string]any{"a": "b"},
			},
			expectError: "kubelet.imageCredentialProviderConfigPath must be a string, got map[string]interface {}",
		},
	}

	for _, tt := range ttests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Kubelet: tt.kubelet}
			gotConfig, gotBinDir, err := c.KubeletImageCredentialProviderRawPaths()
			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantConfig, gotConfig)
			assert.Equal(t, tt.wantBinDir, gotBinDir)
			// c.Kubelet must never be modified during reading.
			assert.Equal(t, tt.kubelet, c.Kubelet)
		})
	}
}

func TestKubeletPassthrough(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		c := &Config{Kubelet: nil}
		assert.Nil(t, c.KubeletPassthrough())
	})

	t.Run("empty map returns empty, non-nil map", func(t *testing.T) {
		c := &Config{Kubelet: map[string]any{}}
		got := c.KubeletPassthrough()
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("map with only reserved keys returns empty map", func(t *testing.T) {
		c := &Config{Kubelet: map[string]any{
			"imageCredentialProviderConfigPath": "/etc/microshift/cp.yaml",
			"imageCredentialProviderBinDir":     "/usr/libexec/cp",
		}}
		assert.Empty(t, c.KubeletPassthrough())
	})

	t.Run("drops exactly the reserved keys and preserves the rest", func(t *testing.T) {
		c := &Config{Kubelet: map[string]any{
			"imageCredentialProviderConfigPath": "/etc/microshift/cp.yaml",
			"imageCredentialProviderBinDir":     "/usr/libexec/cp",
			"cpuManagerPolicy":                  "static",
			"kubeReserved":                      map[string]any{"memory": "500Mi"},
		}}
		got := c.KubeletPassthrough()
		assert.Equal(t, map[string]any{
			"cpuManagerPolicy": "static",
			"kubeReserved":     map[string]any{"memory": "500Mi"},
		}, got)
		// The original map is untouched.
		assert.Contains(t, c.Kubelet, "imageCredentialProviderConfigPath")
		assert.Contains(t, c.Kubelet, "imageCredentialProviderBinDir")
	})
}
