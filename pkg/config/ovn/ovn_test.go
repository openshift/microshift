package ovn

import (
	"fmt"
	"testing"
)

func TestValidateOVSBridge(t *testing.T) {

	var ttests = []struct {
		name string
		err  error
	}{
		{"lo", nil},
		{"unexist-bridge-interface-name", fmt.Errorf("failed to validate bridge interface name")},
	}

	o := new(OVNKubernetesConfig)
	for _, tt := range ttests {
		err := o.ValidateOVSBridge(tt.name)
		if (err != nil) != (tt.err != nil) {
			t.Errorf("ValidateOVSBridge() error = %v, wantErr %v", err, tt.err)
		}
	}
}

// tests to make sure that the config file is parsed correctly
func TestNewOVNKubernetesConfigFromFileOrDefault(t *testing.T) {
	var ttests = []struct {
		configFile string
		err        error
	}{
		{"./test/ovn.yaml", nil},
		{"./test/non-exist.yaml", nil},
	}

	for _, tt := range ttests {
		_, err := NewOVNKubernetesConfigFromFileOrDefault(tt.configFile)
		if (err != nil) != (tt.err != nil) {
			t.Errorf("NewOVNKubernetesConfigFromFileOrDefault() error = %v, wantErr %v", err, tt.err)
		}
	}
}
