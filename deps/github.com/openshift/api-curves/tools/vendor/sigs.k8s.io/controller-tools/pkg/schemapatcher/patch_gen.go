package schemapatcher

import (
	"strings"

	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	kyaml "sigs.k8s.io/yaml"
)

// mayHandleFile returns true if this manifest should progress past the file collection stage.
// Currently, the only check is the feature-set annotation.
func mayHandleFile(filename string, rawContent []byte) bool {
	crdmarkers.FeatureGatesForCurrentFile = sets.String{}

	manifest := &unstructured.Unstructured{}
	if err := kyaml.Unmarshal(rawContent, &manifest); err != nil {
		return true
	}
	// always set the featuregate values
	crdmarkers.FeatureGatesForCurrentFile = featureGatesFromManifest(manifest)

	if len(crdmarkers.RequiredFeatureSets) > 0 {
		manifestFeatureSets := sets.String{}
		if manifestFeatureSetString := manifest.GetAnnotations()["release.openshift.io/feature-set"]; len(manifestFeatureSetString) > 0 {
			for _, curr := range strings.Split(manifestFeatureSetString, ",") {
				manifestFeatureSets.Insert(curr)
			}
		}
		return manifestFeatureSets.Equal(crdmarkers.RequiredFeatureSets)
	}

	return true
}

func featureGatesFromManifest(manifest *unstructured.Unstructured) sets.String {
	ret := sets.String{}
	for existingAnnotation := range manifest.GetAnnotations() {
		if strings.HasPrefix(existingAnnotation, "feature-gate.release.openshift.io/") {
			featureGateName := existingAnnotation[len("feature-gate.release.openshift.io/"):]
			ret.Insert(featureGateName)
		}
	}
	return ret
}
