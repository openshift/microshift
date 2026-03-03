package crd

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// mayHandleField returns true if the field should be considered by this invocation of the generator.
// Right now, the only skip is based on the featureset marker.
func mayHandleField(field markers.FieldInfo) bool {
	if len(crdmarkers.RequiredFeatureSets) > 0 {
		if uncastFeatureSet := field.Markers.Get(crdmarkers.OpenShiftFeatureSetMarkerName); uncastFeatureSet != nil {
			featureSetsForField, ok := uncastFeatureSet.([]string)
			if !ok {
				panic(fmt.Sprintf("actually got %t", uncastFeatureSet))
			}
			//  if any of the field's declared featureSets match any of the manifest's declared featuresets, include the field.
			for _, currFeatureSetForField := range featureSetsForField {
				if crdmarkers.RequiredFeatureSets.Has(currFeatureSetForField) {
					return true
				}
			}
		}
		return false
	}

	// fetch the values for all feature gate markers on the field. If any
	// of the specified feature gates for the field matches the manifest's declared
	// feature gates, include the field.
	featureGateMarkerValues, ok := field.Markers[crdmarkers.OpenShiftFeatureGateMarkerName]
	if !ok {
		// This field is not gated by any feature gates and therefore should be included
		return true
	}

	featureGateValuesSet := sets.New[string]()
	for _, featureGateMarkerValue := range featureGateMarkerValues {
		switch vals := featureGateMarkerValue.(type) {
		case []string:
			featureGateValuesSet.Insert(vals...)
		default:
			panic(fmt.Sprintf("recieved unexpected value type for marker %q, got %t", crdmarkers.OpenShiftFeatureGateMarkerName, vals))
		}
	}

	return crdmarkers.FeatureGatesForCurrentFile.HasAny(featureGateValuesSet.UnsortedList()...)
}
