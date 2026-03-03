package main

import (
	"fmt"
	"strings"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/openshift/api/tools/codegen/pkg/schemacheck"
	"github.com/openshift/crd-schema-checker/pkg/cmd/options"
	"github.com/spf13/cobra"
)

var (
	defaultComparisonConfig = options.NewComparatorOptions()
	comparisonBase          = "master"
)

// schemacheckCmd represents the schemacheck command
var schemacheckCmd = &cobra.Command{
	Use:   "schemacheck",
	Short: "schemacheck validates CRD API schemas based on the best practices",
	RunE: func(cmd *cobra.Command, args []string) error {
		genCtx, err := generation.NewContext(generation.Options{
			BaseDir:          baseDir,
			APIGroupVersions: apiGroupVersions,
		})
		if err != nil {
			return fmt.Errorf("could not build generation context: %w", err)
		}

		gen := newSchemaCheckGenerator()

		return executeGenerators(genCtx, gen)
	},
}

func init() {
	rootCmd.AddCommand(schemacheckCmd)

	knownComparators := strings.Join(defaultComparisonConfig.KnownComparators, ", ")
	rootCmd.PersistentFlags().StringSliceVar(&defaultComparisonConfig.DisabledComparators, "schemacheck:disabled-validators", defaultComparisonConfig.DisabledComparators, fmt.Sprintf("list of comparators that must be disabled. Available comparators: %s", knownComparators))
	rootCmd.PersistentFlags().StringSliceVar(&defaultComparisonConfig.EnabledComparators, "schemacheck:enabled-validators", defaultComparisonConfig.EnabledComparators, fmt.Sprintf("list of comparators that must be enabled. Available comparators: %s", knownComparators))
	rootCmd.PersistentFlags().StringVar(&comparisonBase, "schemacheck:comparison-base", comparisonBase, "base branch/commit to compare against")
}

// newSchemaCheckGenerator builds a new schemacheck generator.
func newSchemaCheckGenerator() generation.Generator {
	return schemacheck.NewGenerator(schemacheck.Options{
		EnabledComparators:  defaultComparisonConfig.EnabledComparators,
		DisabledComparators: defaultComparisonConfig.DisabledComparators,
		ComparisonBase:      comparisonBase,
	})
}
