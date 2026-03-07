package main

import (
	"fmt"

	"github.com/openshift/api/tools/codegen/pkg/crdify"
	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/spf13/cobra"
)

var crdifyComparisonBase = "master"

func newCrdifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crdify",
		Short: "crdify verifies compatibility of CRD API schemas",
		RunE: func(cmd *cobra.Command, args []string) error {
			genCtx, err := generation.NewContext(generation.Options{
				BaseDir:          baseDir,
				APIGroupVersions: apiGroupVersions,
			})
			if err != nil {
				return fmt.Errorf("could not build generation context: %w", err)
			}

			return executeGenerators(genCtx, newCrdifyGenerator())
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(newCrdifyCommand())
	rootCmd.PersistentFlags().StringVar(&crdifyComparisonBase, "crdify:comparison-base", crdifyComparisonBase, "base branch/commit to compare against")
}

func newCrdifyGenerator() generation.Generator {
	return crdify.NewGenerator(crdify.WithComparisonBase(crdifyComparisonBase))
}
