package main

import (
	"fmt"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/openshift/api/tools/codegen/pkg/manifestmerge"
	"github.com/spf13/cobra"
)

var (
	manifestMergePayloadManifestPath string
)

// schemapatchCmd represents the schemapatch command
var crdManifestMerge = &cobra.Command{
	Use:   "crd-manifest-merge",
	Short: "crd-manifest-merge takes all CRD manifests with the same name and merges them together",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		genCtx, err := generation.NewContext(generation.Options{
			BaseDir:          baseDir,
			APIGroupVersions: apiGroupVersions,
		})
		if err != nil {
			return fmt.Errorf("could not build generation context: %w", err)
		}

		gen := newCRDManifestMerger()

		return executeGenerators(genCtx, gen)
	},
}

func init() {
	rootCmd.AddCommand(crdManifestMerge)

	rootCmd.PersistentFlags().StringVar(&manifestMergePayloadManifestPath, "manifest-merge:payload-manifest-path", manifestmerge.DefaultPayloadFeatureGatePath, "path to directory containing the FeatureGate YAMLs for each FeatureSet,ClusterProfile tuple.")
}

// newSchemaPatchGenerator builds a new schemapatch generator.
func newCRDManifestMerger() generation.Generator {
	return manifestmerge.NewGenerator(manifestmerge.Options{
		PayloadFeatureGatePath: manifestMergePayloadManifestPath,
		Verify:                 verify,
	})
}
