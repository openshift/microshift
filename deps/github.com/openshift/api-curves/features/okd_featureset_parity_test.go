package features

import (
	"sort"
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// TestOKDHasAllDefaultFeatureGates verifies that all featuregates enabled in the Default
// featureset are also enabled in the OKD featureset.
// OKD may have additional featuregates beyond Default (e.g., from TechPreviewNoUpgrade
// or DevPreviewNoUpgrade), but it must not be missing any Default featuregates.
func TestOKDHasAllDefaultFeatureGates(t *testing.T) {
	allFeatureSets := AllFeatureSets()

	// Check each cluster profile
	for clusterProfile, byFeatureSet := range allFeatureSets {
		defaultGates, hasDefault := byFeatureSet[configv1.Default]
		okdGates, hasOKD := byFeatureSet[configv1.OKD]

		if !hasOKD || !hasDefault {
			continue
		}

		// Collect enabled feature gate names from Default and OKD
		defaultEnabled := sets.NewString()
		for _, gate := range defaultGates.Enabled {
			defaultEnabled.Insert(string(gate.FeatureGateAttributes.Name))
		}

		okdEnabled := sets.NewString()
		for _, gate := range okdGates.Enabled {
			okdEnabled.Insert(string(gate.FeatureGateAttributes.Name))
		}

		// Check that all Default featuregates are in OKD
		missingInOKD := defaultEnabled.Difference(okdEnabled)

		if missingInOKD.Len() > 0 {
			missingList := missingInOKD.List()
			sort.Strings(missingList)

			t.Errorf("ClusterProfile %q: OKD featureset is missing %d featuregate(s) that are enabled in Default:\n  - %s\n\nAll featuregates enabled in Default must also be enabled in OKD.",
				clusterProfile,
				missingInOKD.Len(),
				strings.Join(missingList, "\n  - "),
			)
		}
	}
}
