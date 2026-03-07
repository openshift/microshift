package swaggerdocs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// Options contains the configuration required for the swaggerdocs generator.
type Options struct {
	// Disabled indicates whether the swaggerdocs generator is enabled or not.
	// This default to false as the swaggerdocs generator is enabled by default.
	Disabled bool

	// CommentPolicy determines how, when verifying swaggerdocs, the generator
	// should handle missing comments.
	// Valid values are `Ignore`, `Warn` and `Enforce`.
	// This defaults to `Warn`.
	// When set to `Ignore`, the generator will ignore any missing comments.
	// When set to `Warn`, the generator will emit a warning for any missing comments.
	// When set to `Enforce`, the generator will return an error for any missing comments.
	CommentPolicy string

	// OutputFileName is the file name to use for writing the generated swagger
	// docs to. This file will be created for each group version.
	OutputFileName string

	// Verify determines whether the generator should verify the content instead
	// of updating the generated file.
	Verify bool
}

// generator implements the generation.Generator interface.
// It is designed to generate swaggerdocs documentation for a particular API group.
type generator struct {
	disabled       bool
	commentPolicy  string
	outputFileName string
	verify         bool
}

// NewGenerator builds a new schemapatch generator.
func NewGenerator(opts Options) generation.Generator {
	return &generator{
		disabled:       opts.Disabled,
		commentPolicy:  opts.CommentPolicy,
		outputFileName: opts.OutputFileName,
		verify:         opts.Verify,
	}
}

// ApplyConfig creates returns a new generator based on the configuration passed.
// If the schemapatch configuration is empty, the existing generation is returned.
func (g *generator) ApplyConfig(config *generation.Config) generation.Generator {
	if config == nil || config.SwaggerDocs == nil {
		return g
	}

	outputFileName := DefaultOutputFileName
	if config.SwaggerDocs.OutputFileName != "" {
		outputFileName = config.SwaggerDocs.OutputFileName
	}

	return NewGenerator(Options{
		Disabled:       config.SwaggerDocs.Disabled,
		CommentPolicy:  config.SwaggerDocs.CommentPolicy,
		OutputFileName: outputFileName,
		Verify:         g.verify,
	})
}

// Name returns the name of the generator.
func (g *generator) Name() string {
	return "swaggerdocs"
}

// GenGroup runs the schemapatch generator against the given group context.
func (g *generator) GenGroup(groupCtx generation.APIGroupContext) ([]generation.Result, error) {
	if g.disabled {
		klog.V(2).Infof("Skipping swaggerdocs generation for %s", groupCtx.Name)
		return nil, nil
	}

	for _, version := range groupCtx.Versions {
		if err := g.generateGroupVersion(groupCtx.Name, version); err != nil {
			return nil, fmt.Errorf("error generating swagger docs for %s/%s: %w", groupCtx.Name, version.Name, err)
		}
	}

	return nil, nil
}

// generateGroupVersion generates swagger docs for the group version.
func (g *generator) generateGroupVersion(groupName string, version generation.APIVersionContext) error {
	outFilePath := filepath.Join(version.Path, g.outputFileName)

	versionGlob := filepath.Join(version.Path, typesGlob)
	files, err := filepath.Glob(versionGlob)
	if err != nil {
		return fmt.Errorf("could not read types*.go files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no match for types*.go glob in path %s", version.Path)
	}

	docsForTypes := []kruntime.KubeTypes{}
	for _, file := range files {
		docsForTypes = append(docsForTypes, kruntime.ParseDocumentationFrom(file)...)
	}

	if g.verify {
		klog.V(1).Infof("Verifiying swagger docs for %s/%s", groupName, version.Name)

		return verifySwaggerDocs(version.PackageName, outFilePath, docsForTypes, g.commentPolicy)
	}

	klog.V(1).Infof("Generating swagger docs for %s/%s", groupName, version.Name)

	generatedDocs, err := generateSwaggerDocs(version.PackageName, docsForTypes)
	if err != nil {
		return fmt.Errorf("error generating swagger docs: %w", err)
	}

	if err := os.WriteFile(outFilePath, generatedDocs, 0644); err != nil {
		return fmt.Errorf("error writing swagger docs output: %w", err)
	}

	return nil
}
