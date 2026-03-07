package deepcopy

import (
	"fmt"
	"os"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"k8s.io/gengo/v2/parser"
	"k8s.io/gengo/v2/types"
	"k8s.io/klog/v2"
)

const (
	// DefaultOutputFileBaseName is the default output file base name for the generated deepcopy functions.
	DefaultOutputFileBaseName = "zz_generated.deepcopy.go"
)

// Options contains the configuration required for the compatibility generator.
type Options struct {
	// Disabled indicates whether the deepcopy generator is enabled or not.
	// This default to false as the deepcopy generator is enabled by default.
	Disabled bool

	// HeaderFilePath is the path to the file containing the boilerplate header text.
	// When omitted, no header is added to the generated files.
	HeaderFilePath string

	// OutputFileBaseName is the base name of the output file.
	// When omitted, DefaultOutputFileBaseName is used.
	// The current value of DefaultOutputFileBaseName is "zz_generated.deepcopy".
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

// generator implements the generation.Generator interface.
// It is designed to generate deepcopy function for a particular API group.
type generator struct {
	disabled           bool
	headerFilePath     string
	outputBaseFileName string
	verify             bool
	globalParser       *parser.Parser
	universe           types.Universe
}

// NewGenerator builds a new deepcopy generator.
func NewGenerator(opts Options) generation.Generator {
	outputFileBaseName := DefaultOutputFileBaseName
	if opts.OutputFileBaseName != "" {
		outputFileBaseName = opts.OutputFileBaseName
	}

	return &generator{
		disabled:           opts.Disabled,
		headerFilePath:     opts.HeaderFilePath,
		outputBaseFileName: outputFileBaseName,
		verify:             opts.Verify,
		globalParser:       opts.GlobalParser,
		universe:           opts.Universe,
	}
}

// ApplyConfig creates returns a new generator based on the configuration passed.
// If the deepcopy configuration is empty, the existing generation is returned.
func (g *generator) ApplyConfig(config *generation.Config) generation.Generator {
	if config == nil || config.Deepcopy == nil {
		return g
	}

	return NewGenerator(Options{
		Disabled:           config.Deepcopy.Disabled,
		HeaderFilePath:     config.Deepcopy.HeaderFilePath,
		OutputFileBaseName: config.Deepcopy.OutputFileBaseName,
		Verify:             g.verify,
		GlobalParser:       g.globalParser,
		Universe:           g.universe,
	})
}

// Name returns the name of the generator.
func (g *generator) Name() string {
	return "deepcopy"
}

// GenGroup runs the deepcopy generator against the given group context.
func (g *generator) GenGroup(groupCtx generation.APIGroupContext) ([]generation.Result, error) {
	if g.disabled {
		klog.V(2).Infof("Skipping deepcopy generation for %s", groupCtx.Name)
		return nil, nil
	}

	// If there is no header file, create an empty file and pass that through.
	headerFilePath := g.headerFilePath
	if headerFilePath == "" {
		tmpFile, err := os.CreateTemp("", "deepcopy-header-*.txt")
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary file: %w", err)
		}
		tmpFile.Close()

		defer os.Remove(tmpFile.Name())

		headerFilePath = tmpFile.Name()
	}

	for _, version := range groupCtx.Versions {
		action := "Generating"
		if g.verify {
			action = "Verifying"
		}

		klog.V(1).Infof("%s deepcopy functions for for %s/%s", action, groupCtx.Name, version.Name)

		if err := generateDeepcopyFunctions(g.globalParser, g.universe, version.Path, version.PackagePath, g.outputBaseFileName, headerFilePath, g.verify); err != nil {
			return nil, fmt.Errorf("could not generate deepcopy functions for %s/%s: %w", groupCtx.Name, version.Name, err)
		}
	}

	return nil, nil
}
