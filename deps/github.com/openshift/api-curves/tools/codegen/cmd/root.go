package main

import (
	"flag"
	"fmt"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/openshift/api/tools/codegen/pkg/utils"
	"github.com/spf13/cobra"

	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
)

var (
	apiGroupVersions []string
	baseDir          string
	verify           bool

	// version will be set by the makefile when the binary is built.
	// It should be the git commit hash of the last commit that
	// affected the code used to build this tool.
	version      = "Unknown"
	printVersion bool
)

// rootCmd represents the base command when called without any subcommands.
// This will run all generators in the preferred order for OpenShift APIs.
var rootCmd = &cobra.Command{
	Use:           "codegen",
	Short:         "Codegen runs code generators for the OpenShift API definitions",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if printVersion {
			fmt.Printf("%s\n", version)
			return nil
		}

		genCtx, err := generation.NewContext(generation.Options{
			BaseDir:          baseDir,
			APIGroupVersions: apiGroupVersions,
		})
		if err != nil {
			return fmt.Errorf("could not build generation context: %w", err)
		}

		generators := allGenerators(genCtx)
		if verify {
			generators = append(generators, allVerifiers()...)
		}

		if err := executeGenerators(genCtx, generators...); err != nil {
			return fmt.Errorf("could not run generators: %w", err)
		}

		return executeMultiGroupGenerators(genCtx, allMultiGroupGenerators(genCtx)...)
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		klog.Fatalf("Error running codegen: %v", err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&apiGroupVersions, "api-group-versions", []string{}, "A list of API group versions in the form <group>/<version>. The group should be fully qualified, e.g. machine.openshift.io/v1. The generator will generate against all group versions found within the base directory when no specific group versions are provided.")
	rootCmd.PersistentFlags().StringVar(&baseDir, "base-dir", ".", "Base directory to search for API group versions")
	rootCmd.PersistentFlags().BoolVar(&verify, "verify", false, "Verifies the content of generated files are up to date.")
	rootCmd.PersistentFlags().BoolVar(&printVersion, "version", false, "Print the version of the codegen tool and exit")

	klog.InitFlags(flag.CommandLine)
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
}

// executeGenerators runs each generator for each group in the generation context.
// If an error occurs for a generator within a group, the rest of the generators are ignored for that group.
// Subsequent groups will continue to generate.
func executeGenerators(genCtx generation.Context, generators ...generation.Generator) error {
	errs := []error{}
	resultsByGroup := map[string][]generation.Result{}

	for _, group := range genCtx.APIGroups {
		klog.Infof("Running generators for %s", group.Name)

		for _, gen := range generators {
			g := gen
			if group.Config != nil {
				g = g.ApplyConfig(group.Config)
			}

			results, err := g.GenGroup(group)
			if err != nil {
				errs = append(errs, fmt.Errorf("error running generator %s on group %s: %w", gen.Name(), group.Name, err))
			}

			resultsByGroup[group.Name] = append(resultsByGroup[group.Name], results...)
		}
	}

	if err := utils.PrintResults(resultsByGroup); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return utils.NewAggregatePrinter(kerrors.NewAggregate(errs))
	}

	return nil
}

// executeMultiGroupGenerators runs each multi-group generator for the generation context.
// Each generator error is aggregated and returned.
func executeMultiGroupGenerators(genCtx generation.Context, generators ...generation.MultiGroupGenerator) error {
	errs := []error{}

	for _, gen := range generators {
		if err := gen.GenGroups(genCtx.APIGroups); err != nil {
			errs = append(errs, fmt.Errorf("error running generator %s: %w", gen.Name(), err))
		}
	}

	if len(errs) > 0 {
		return kerrors.NewAggregate(errs)
	}

	return nil
}

// allGenerators returns an ordered list of generators to run when
// the root command is executed.
func allGenerators(genCtx generation.Context) []generation.Generator {
	return []generation.Generator{
		newCompatibilityGenerator(),
		newDeepcopyGenerator(genCtx),
		newSwaggerDocsGenerator(),
		// The empty partial schema, schema patch and manifest merge must run in order.
		newEmptyPartialSchemaGenerator(genCtx),
		newSchemaPatchGenerator(),
		newCRDManifestMerger(),
	}
}

// allVerifiers returns an ordered list of verifiers to run when
// the root command is executed with the --verify flag.
func allVerifiers() []generation.Generator {
	return []generation.Generator{
		// Schema checker and crdify are invoked separately as we can override these
		// depending on circumstances.
		// All generators/verifiers that are part of codegen and executed with a bare
		// codegen invocation must be absolutely required/not overrideable.
		// newSchemaCheckGenerator(),
		// newCrdifyGenerator(),
	}
}

// allMultiGroupGenerators returns an ordered list of multi-group
// generators to run when the root command is executed.
func allMultiGroupGenerators(genCtx generation.Context) []generation.MultiGroupGenerator {
	return []generation.MultiGroupGenerator{
		newOpenAPIGenerator(genCtx),
	}
}
