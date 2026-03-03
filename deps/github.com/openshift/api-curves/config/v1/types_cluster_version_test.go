package v1

import (
	"testing"
)

// TestKnownClusterVersionCapabilities verifies that all capabilities
// referenced by capability sets are contained in
// KnownClusterVersionCapabilities.
func TestKnownClusterVersionCapabilities(t *testing.T) {
	exists := struct{}{}
	known := make(map[ClusterVersionCapability]struct{}, len(KnownClusterVersionCapabilities))
	for _, cap := range KnownClusterVersionCapabilities {
		known[cap] = exists
	}

	for set, caps := range ClusterVersionCapabilitySets {
		for _, cap := range caps {
			if _, ok := known[cap]; !ok {
				t.Errorf("Capability set %s contains %s, which needs to be added to KnownClusterVersionCapabilities", set, cap)
			}
		}
	}
}
