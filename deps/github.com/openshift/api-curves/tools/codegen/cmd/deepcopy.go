package main

import (
	"fmt"

	"github.com/openshift/api/tools/codegen/pkg/deepcopy"
	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/spf13/cobra"
)

var (
	deepcopyHeaderFilePath     string
	deepcopyOutputFileBaseName string
)

// deepcopyCmd represents the deepcopy command
var deepcopyCmd = &cobra.Command{
	Use:   "deepcopy",
	Short: "deepcopy generates deepcopy functions for API types",
	RunE: func(cmd *cobra.Command, args []string) error {
		genCtx, err := generation.NewContext(generation.Options{
			BaseDir:          baseDir,
			APIGroupVersions: apiGroupVersions,
		})
		if err != nil {
			return fmt.Errorf("could not build generation context: %w", err)
		}

		gen := newDeepcopyGenerator(genCtx)

		return executeGenerators(genCtx, gen)
	},
}

func init() {
	rootCmd.AddCommand(deepcopyCmd)

	rootCmd.PersistentFlags().StringVar(&deepcopyHeaderFilePath, "deepcopy:header-file-path", "", "Path to file containing boilerplate header text. The string YEAR will be replaced with the current 4-digit year. When omitted, no header is added to the generated files.")
	rootCmd.PersistentFlags().StringVar(&deepcopyOutputFileBaseName, "deepcopy:output-file-base-name", deepcopy.DefaultOutputFileBaseName, "Base name of the output file. When omitted, zz_generated.deepcopy is used.")
}

// newDeepcopyhGenerator builds a new deepcopy generator.
func newDeepcopyGenerator(genCtx generation.Context) generation.Generator {
	return deepcopy.NewGenerator(deepcopy.Options{
		HeaderFilePath:     deepcopyHeaderFilePath,
		OutputFileBaseName: deepcopyOutputFileBaseName,
		Verify:             verify,
		GlobalParser:       genCtx.GlobalParser,
		Universe:           genCtx.Universe,
	})
}
