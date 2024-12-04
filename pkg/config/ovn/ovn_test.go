package ovn

import (
	"testing"

	"github.com/vishvananda/netlink"
)

// tests to make sure that the config file is parsed correctly
func TestNewOVNKubernetesConfigFromFileOrDefault(t *testing.T) {
	var ttests = []struct {
		configFile string
		err        error
	}{
		{"./test/", nil},
		{"./test/non-exist-dir", nil},
	}

	for _, tt := range ttests {
		_, err := NewOVNKubernetesConfigFromFileOrDefault(tt.configFile, false, netlink.FAMILY_V4)
		if (err != nil) != (tt.err != nil) {
			t.Errorf("NewOVNKubernetesConfigFromFileOrDefault() error = %v, wantErr %v", err, tt.err)
		}
	}
}
