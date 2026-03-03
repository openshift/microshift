package emptypartialschemas

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/parser"
	"k8s.io/gengo/v2/types"
	"k8s.io/klog/v2"
)

// Options contains the configuration required for the compatibility generator.
type Options struct {
	// Disabled indicates whether the empty-partial-schemas generator is enabled or not.
	// This default to false as the empty-partial-schemas generator is enabled by default.
	Disabled bool

	// OutputFileBaseName is the base name of the output file.
	// When omitted, DefaultOutputFileBaseName is used.
	// The current value of DefaultOutputFileBaseName is "MISSING".
	OutputFileBaseName string

	// Verify determines whether the generator should verify the content instead
	// of updating the generated file.
	Verify bool

	// GlobalParser is the parser for the global package.
	// This loads all packages found in the base directory.
	GlobalParser *parser.Parser

	// Universe is the universe for the global package.
	Universe types.Universe
}

// emptyPartialSchemasGenerator implements the generation.Generator interface.
// It is designed to generate empty-partial-schemas function for a particular API group.
type emptyPartialSchemasGenerator struct {
	disabled           bool
	outputBaseFileName string
	verify             bool
	globalParser       *parser.Parser
	universe           types.Universe
}

// NewGenerator builds a new empty-partial-schemas generator.
func NewGenerator(opts Options) generation.Generator {
	outputFileBaseName := "MISSING"
	if opts.OutputFileBaseName != "" {
		outputFileBaseName = opts.OutputFileBaseName
	}

	return &emptyPartialSchemasGenerator{
		disabled:           opts.Disabled,
		outputBaseFileName: outputFileBaseName,
		verify:             opts.Verify,
		globalParser:       opts.GlobalParser,
		universe:           opts.Universe,
	}
}

// ApplyConfig creates returns a new generator based on the configuration passed.
// If the empty-partial-schemas configuration is empty, the existing generation is returned.
func (g *emptyPartialSchemasGenerator) ApplyConfig(config *generation.Config) generation.Generator {
	if config == nil || config.EmptyPartialSchema == nil {
		return g
	}

	return NewGenerator(Options{
		Disabled:           config.EmptyPartialSchema.Disabled,
		OutputFileBaseName: g.outputBaseFileName,
		Verify:             g.verify,
		GlobalParser:       g.globalParser,
		Universe:           g.universe,
	})
}

// Name returns the name of the generator.
func (g *emptyPartialSchemasGenerator) Name() string {
	return "partial-crd-manifests"
}

// GenGroup runs the empty-partial-schemas generator against the given group context.
func (g *emptyPartialSchemasGenerator) GenGroup(groupCtx generation.APIGroupContext) ([]generation.Result, error) {
	if g.disabled {
		klog.V(2).Infof("Skipping %q generation for %s", g.Name(), groupCtx.Name)
		return nil, nil
	}

	for _, version := range groupCtx.Versions {
		action := "Generating"
		if g.verify {
			action = "Verifying"
		}

		klog.Infof("%s %q functions for for %s/%s", action, g.Name(), groupCtx.Name, version.Name)

		if err := g.generatePartialSchemaFiles(g.globalParser, g.universe, version.Path, version.PackagePath, g.verify); err != nil {
			return nil, fmt.Errorf("could not generate %v functions for %s/%s: %w", g.Name(), groupCtx.Name, version.Name, err)
		}
	}

	return nil, nil
}

// generatePartialSchemaFiles generates the DeepCopy functions for the given API package paths.
func (g *emptyPartialSchemasGenerator) generatePartialSchemaFiles(p *parser.Parser, universe types.Universe, path, packagePath string, verify bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// The empty-partial-schemas generator cannot import from an absolute path.
	inputPath, err := filepath.Rel(wd, path)
	if err != nil {
		return fmt.Errorf("failed to get relative path for %s: %w", path, err)
	}
	// The path must start with `./` to be considered a relative path
	// by the generator.
	inputPath = fmt.Sprintf(".%s%s", string(os.PathSeparator), inputPath)

	// Create result gatherer to collect CRD information during generation
	gengoGeneratorResults := &generatorResultGatherer{
		crdNamesToFeatureGates: map[string]*CRDInfo{},
		boundingDirs:           []string{inputPath},
		outputFileBaseName:     g.outputBaseFileName,
	}

	myTargets := func(context *generator.Context) []generator.Target {
		return gengoGeneratorResults.GetTargets(context)
	}

	if err := generation.Execute(p, universe,
		NameSystems(),
		DefaultNameSystem(),
		myTargets,
		[]string{packagePath},
	); err != nil {
		return fmt.Errorf("error executing %v generator: %w", g.Name(), err)
	}

	directoryForPartialContent := filepath.Join(inputPath, "zz_generated.featuregated-crd-manifests")
	return createFeatureGatedCRDManifests(gengoGeneratorResults.crdNamesToFeatureGates, directoryForPartialContent, g.verify)
}
