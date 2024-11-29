package healthcheck

import (
	"testing"

	"github.com/openshift/microshift/pkg/config"
	"github.com/stretchr/testify/assert"
)

func Test_csiComponentsAreExpected(t *testing.T) {
	testData := []struct {
		name           string
		cfg            config.Config
		expectedResult bool
	}{
		{
			name: "empty",
			cfg: config.Config{Storage: config.Storage{
				OptionalCSIComponents: []config.OptionalCsiComponent{},
			}},
			expectedResult: true,
		},
		{
			name: "none",
			cfg: config.Config{Storage: config.Storage{
				OptionalCSIComponents: []config.OptionalCsiComponent{config.CsiComponentNone},
			}},
			expectedResult: false,
		},
		{
			name: "only controller",
			cfg: config.Config{Storage: config.Storage{
				OptionalCSIComponents: []config.OptionalCsiComponent{config.CsiComponentSnapshot},
			}},
			expectedResult: true,
		},
		{
			name: "only webhook",
			cfg: config.Config{Storage: config.Storage{
				OptionalCSIComponents: []config.OptionalCsiComponent{config.CsiComponentSnapshotWebhook},
			}},
			expectedResult: true,
		},
		{
			name: "controller & webhook",
			cfg: config.Config{Storage: config.Storage{
				OptionalCSIComponents: []config.OptionalCsiComponent{config.CsiComponentSnapshot, config.CsiComponentSnapshotWebhook},
			}},
			expectedResult: true,
		},
	}
	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			cfg := td.cfg
			result := csiComponentsAreExpected(&cfg)
			assert.Equal(t, td.expectedResult, result)
		})
	}
}
