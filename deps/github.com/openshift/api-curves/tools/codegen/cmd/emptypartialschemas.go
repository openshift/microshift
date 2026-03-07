package main

import (
	"fmt"

	"github.com/openshift/api/tools/codegen/pkg/emptypartialschemas"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/spf13/cobra"
)

var (
	emptyPartialSchemaReportOutputFileBaseName string = "zz_generated.featuregated-crd-manifests.yaml"
)

// emptyPartialSchemasCmd represents the empty-partial-schemas command
var emptyPartialSchemasCmd = &cobra.Command{
	Use:   "empty-partial-schemas",
	Short: "creates empty partial schema files to start from",
	RunE: func(cmd *cobra.Command, args []string) error {
		genCtx, err := generation.NewContext(generation.Options{
			BaseDir:          baseDir,
			APIGroupVersions: apiGroupVersions,
		})
		if err != nil {
			return fmt.Errorf("could not build generation context: %w", err)
		}

		gen := newEmptyPartialSchemaGenerator(genCtx)

		return executeGenerators(genCtx, gen)
	},
}

func init() {
	rootCmd.AddCommand(emptyPartialSchemasCmd)
}

// newDeepcopyhGenerator builds a new empty-partial-schemas generator.
func newEmptyPartialSchemaGenerator(genCtx generation.Context) generation.Generator {
	return emptypartialschemas.NewGenerator(emptypartialschemas.Options{
		OutputFileBaseName: emptyPartialSchemaReportOutputFileBaseName,
		Verify:             verify,
		GlobalParser:       genCtx.GlobalParser,
		Universe:           genCtx.Universe,
	})
}
