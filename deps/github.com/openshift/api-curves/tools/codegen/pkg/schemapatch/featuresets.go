package schemapatch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/api/tools/codegen/pkg/generation"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	kyaml "sigs.k8s.io/yaml"
)

// shouldProcessGroupVersion determines, based on the required feature sets, whether this group version should be
// generated or not.
func shouldProcessGroupVersion(version generation.APIVersionContext, requiredFeatureSets []sets.String) (bool, error) {
	dirEntries, err := os.ReadDir(version.Path)
	if err != nil {
		return false, fmt.Errorf("could not read file info for directory %s: %v", version.Path, err)
	}

	for _, fileInfo := range dirEntries {
		// Find all files that are yaml-patches
		if fileInfo.IsDir() || filepath.Ext(fileInfo.Name()) != ".yaml" {
			continue
		}

		fileName := filepath.Join(version.Path, fileInfo.Name())
		data, err := os.ReadFile(fileName)
		if err != nil {
			return false, fmt.Errorf("could not read CRD file %s: %v", fileInfo.Name(), err)
		}

		if mayHandleFile(data, requiredFeatureSets) {
			// At least one file needs to be processed, process the whole group version.
			return true, nil
		}
	}

	return false, nil
}

// mayHandleFile determines, from the feature sets, whether this patch should be handled.
// Currently, the only check is the feature-set annotation.
func mayHandleFile(rawContent []byte, requiredFeatureSets []sets.String) bool {
	manifest := &unstructured.Unstructured{}
	if err := kyaml.Unmarshal(rawContent, &manifest); err != nil {
		return true
	}

	if len(requiredFeatureSets) == 0 {
		return mayHandleObject(manifest, sets.NewString())
	}

	for _, requiredFeatureSet := range requiredFeatureSets {
		if mayHandleObject(manifest, requiredFeatureSet) {
			return true
		}
	}

	return false
}

// mayHandleObject determines, from the feature sets, whether a kube like object should be handled.
// Currently, the only check is the feature-set annotation.
func mayHandleObject(manifest metav1.Object, requiredFeatureSets sets.String) bool {
	manifestFeatureSets := getObjectFeatureSets(manifest)

	return manifestFeatureSets.Equal(requiredFeatureSets)
}

// getObjectFeatureSets returns the feature sets for a kube like object.
func getObjectFeatureSets(manifest metav1.Object) sets.String {
	manifestFeatureSets := sets.NewString()
	if manifestFeatureSetString := manifest.GetAnnotations()["release.openshift.io/feature-set"]; len(manifestFeatureSetString) > 0 {
		for _, curr := range strings.Split(manifestFeatureSetString, ",") {
			manifestFeatureSets.Insert(curr)
		}
	}

	return manifestFeatureSets
}
