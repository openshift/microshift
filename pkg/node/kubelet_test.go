package node

import (
	"testing"

	"github.com/openshift/microshift/pkg/config"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateConfig(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Kubelet = map[string]any{
		"cpuManagerPolicy": "static",
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

	data, err := generateConfig(cfg)
	assert.NoError(t, err)
	assert.Contains(t, string(data), expectedConfigPart)
}
