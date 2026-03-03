package main

import (
	"fmt"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/openshift/api/tools/codegen/pkg/openapi"
	"github.com/spf13/cobra"
)

var (
	openapiHeaderFilePath    string
	openapiOutputFileName    string
	openapiOutputPackagePath string
)

// openapiCmd represents the openapi command
var openapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: "openapi generates openapi schema for API types",
	RunE: func(cmd *cobra.Command, args []string) error {
		genCtx, err := generation.NewContext(generation.Options{
			BaseDir:          baseDir,
			APIGroupVersions: apiGroupVersions,
		})
		if err != nil {
			return fmt.Errorf("could not build generation context: %w", err)
		}

		gen := newOpenAPIGenerator(genCtx)

		return executeMultiGroupGenerators(genCtx, gen)
	},
}

func init() {
	rootCmd.AddCommand(openapiCmd)

	rootCmd.PersistentFlags().StringVar(&openapiHeaderFilePath, "openapi:header-file-path", "", "Path to file containing boilerplate header text. The string YEAR will be replaced with the current 4-digit year. When omitted, no header is added to the generated files.")
	rootCmd.PersistentFlags().StringVar(&openapiOutputFileName, "openapi:output-file-name", openapi.DefaultOutputFileName, "Name of the output file. When omitted, zz_generated.openapi.go is used.")
	rootCmd.PersistentFlags().StringVar(&openapiOutputPackagePath, "openapi:output-package-path", openapi.DefaultOutputPackagePath, "Package path where the generated golang files will be written.")
}

// newOpenAPIGenerator builds a new openapi generator.
func newOpenAPIGenerator(genCtx generation.Context) generation.MultiGroupGenerator {
	return openapi.NewGenerator(openapi.Options{
		HeaderFilePath:    openapiHeaderFilePath,
		OutputFileName:    openapiOutputFileName,
		OutputPackagePath: openapiOutputPackagePath,
		Verify:            verify,
		GlobalParser:      genCtx.GlobalParser,
		Universe:          genCtx.Universe,
	})
}
