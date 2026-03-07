package assets

import (
	"fmt"
	"path/filepath"
)

// SubstituteAndCopyFiles read files from the input dir, selects some by predicate, transforms them, and writes the content to output dir.
func SubstituteAndCopyFiles(assetInputDir, assetOutputDir, featureSet, clusterProfile string, templateData interface{}, additionalPredicates ...FileInfoPredicate) error {
	defaultPredicates := []FileInfoPredicate{OnlyYaml}
	manifestPredicates := []FileContentsPredicate{
		InstallerFeatureSet(featureSet),
		ClusterProfile(clusterProfile),
		BootstrapRequiredCRD(),
	}

	// write assets
	manifests, err := New(
		assetInputDir,
		templateData,
		manifestPredicates,
		append(additionalPredicates, defaultPredicates...)...,
	)
	if err != nil {
		return fmt.Errorf("failed rendering assets: %v", err)
	}
	if err := manifests.WriteFiles(filepath.Join(assetOutputDir)); err != nil {
		return fmt.Errorf("failed writing assets to %q: %v", filepath.Join(assetOutputDir), err)
	}

	return nil
}
