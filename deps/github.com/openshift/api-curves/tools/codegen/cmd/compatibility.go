package main

import (
	"fmt"

	"github.com/openshift/api/tools/codegen/pkg/compatibility"
	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/spf13/cobra"
)

// compatibilityCmd represents the compatibility command
var compatibilityCmd = &cobra.Command{
	Use:   "compatibility",
	Short: "compatibility generates compatibility level comments for API definitions",
	Long: `compatibility generates a compatibility level comment for each API defintion.
	The generation is controlled by a marker applied to the CRD struct defintiion.
	For example, this annotation would be +openshift:compatibility-gen:level=1 for a
	level 1 API.
	
	Valid API levels are 1, 2, 3 and 4. Version 1 is required for all stable APIs.
	Version 2 should be used for beta level APIs. Levels 3 and 4 may be used for
	alpha APIs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		genCtx, err := generation.NewContext(generation.Options{
			BaseDir:          baseDir,
			APIGroupVersions: apiGroupVersions,
		})
		if err != nil {
			return fmt.Errorf("could not build generation context: %w", err)
		}

		gen := newCompatibilityGenerator()

		return executeGenerators(genCtx, gen)
	},
}

func init() {
	rootCmd.AddCommand(compatibilityCmd)
}

// newCompatibilityhGenerator builds a new compatibility generator.
func newCompatibilityGenerator() generation.Generator {
	return compatibility.NewGenerator(compatibility.Options{
		Verify: verify,
	})
}
