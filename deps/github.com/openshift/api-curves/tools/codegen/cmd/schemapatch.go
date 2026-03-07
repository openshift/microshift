package main

import (
	"fmt"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/openshift/api/tools/codegen/pkg/schemapatch"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	controllerGen       string
	requiredFeatureSets []string
)

// schemapatchCmd represents the schemapatch command
var schemapatchCmd = &cobra.Command{
	Use:   "schemapatch",
	Short: "schemapatch updates CRD API schemas based on the API definition",
	Long: `schemapatch runs the controller-gen schemapatch generator
	against API groups to update CRD API schemas.
	CRD files must exist before the generator can patch the schema.

	Once the schema has been generated, the generator will apply any
	yaml-patch files found and then format the output yaml files.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		genCtx, err := generation.NewContext(generation.Options{
			BaseDir:          baseDir,
			APIGroupVersions: apiGroupVersions,
		})
		if err != nil {
			return fmt.Errorf("could not build generation context: %w", err)
		}

		gen := newSchemaPatchGenerator()

		return executeGenerators(genCtx, gen)
	},
}

func init() {
	rootCmd.AddCommand(schemapatchCmd)

	rootCmd.PersistentFlags().StringVar(&controllerGen, "controller-gen", "", "Path to the controller-gen tool to use. If omitted, will use the built in generator (Only applicable to the schemapatch generator)")
	rootCmd.PersistentFlags().StringSliceVar(&requiredFeatureSets, "required-feature-sets", []string{}, "Specific feature sets to generate CRDs schemas for (Only applicable to the schemapatch generator)")
}

// newSchemaPatchGenerator builds a new schemapatch generator.
func newSchemaPatchGenerator() generation.Generator {
	requiredFeatureSetsList := []sets.String{}
	if len(requiredFeatureSets) > 0 {
		requiredFeatureSetsList = append(requiredFeatureSetsList, sets.NewString(requiredFeatureSets...))
	}

	return schemapatch.NewGenerator(schemapatch.Options{
		ControllerGen:       controllerGen,
		RequiredFeatureSets: requiredFeatureSetsList,
		Verify:              verify,
	})
}
