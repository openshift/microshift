package node

import (
	"testing"

	"github.com/openshift/microshift/pkg/config"
	kubeletoptions "k8s.io/kubernetes/cmd/kubelet/app/options"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateConfig(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Kubelet = map[string]any{
		"cpuManagerPolicy": "static",
		// Reserved keys are MicroShift-owned kubelet flags and must never
		// appear in the generated KubeletConfiguration.
		"imageCredentialProviderConfigPath": "/etc/microshift/credential-providers.yaml",
		"imageCredentialProviderBinDir":     "/usr/libexec/microshift/credential-providers",
		"reservedMemory": []any{
			map[string]any{
				"limits": map[string]any{
					"memory": "1100Mi",
				},
				"numaNode": float64(0),
			},
		},
		"kubeReserved": map[string]any{
			"memory": "500Mi",
		},
		"evictionHard": map[string]any{
			"imagefs.available": "15%",
			"memory.available":  "100Mi",
			"nodefs.available":  "10%",
			"nodefs.inodesFree": "5%",
		},
	}

	expectedConfigPart := `cpuManagerPolicy: static
evictionHard:
  imagefs.available: 15%
  memory.available: 100Mi
  nodefs.available: 10%
  nodefs.inodesFree: 5%
kubeReserved:
  memory: 500Mi
reservedMemory:
- limits:
    memory: 1100Mi
  numaNode: 0`

	kubelet := &KubeletServer{}
	data, err := kubelet.generateConfig(cfg)
	assert.NoError(t, err)
	assert.Contains(t, string(data), expectedConfigPart)
	// The reserved keys are stripped from the passthrough config.
	assert.NotContains(t, string(data), "imageCredentialProviderConfigPath")
	assert.NotContains(t, string(data), "imageCredentialProviderBinDir")
}

func Test_GenerateConfig_EmptyKubelet(t *testing.T) {
	// An empty (or reserved-keys-only) kubelet map must not inject anything into
	// the generated KubeletConfiguration: the output must match the nil-map case,
	// with no stray "{}" appended.
	kubelet := &KubeletServer{}

	nilCfg := config.NewDefault()
	nilCfg.Kubelet = nil
	nilData, err := kubelet.generateConfig(nilCfg)
	assert.NoError(t, err)

	emptyCfg := config.NewDefault()
	emptyCfg.Kubelet = map[string]any{}
	emptyData, err := kubelet.generateConfig(emptyCfg)
	assert.NoError(t, err)

	assert.Equal(t, string(nilData), string(emptyData))
	assert.NotContains(t, string(emptyData), "{}")
}

func Test_setImageCredentialProviderFlags(t *testing.T) {
	t.Run("sets both flags to the canonical values when configured", func(t *testing.T) {
		cfg := &config.Config{
			KubeletImageCredentialProviderConfigPath: "/etc/microshift/credential-providers.yaml",
			KubeletImageCredentialProviderBinDir:     "/usr/libexec/microshift/credential-providers",
		}
		flags := kubeletoptions.NewKubeletFlags()
		setImageCredentialProviderFlags(flags, cfg)
		assert.Equal(t, "/etc/microshift/credential-providers.yaml", flags.ImageCredentialProviderConfigPath)
		assert.Equal(t, "/usr/libexec/microshift/credential-providers", flags.ImageCredentialProviderBinDir)
	})

	t.Run("leaves flags empty when not configured", func(t *testing.T) {
		cfg := &config.Config{}
		flags := kubeletoptions.NewKubeletFlags()
		setImageCredentialProviderFlags(flags, cfg)
		assert.Empty(t, flags.ImageCredentialProviderConfigPath)
		assert.Empty(t, flags.ImageCredentialProviderBinDir)
	})
}
