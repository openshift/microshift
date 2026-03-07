package compatibility

import (
	"fmt"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"k8s.io/klog/v2"
)

// Options contains the configuration required for the compatibility generator.
type Options struct {
	// Disabled indicates whether the compatibility generator is enabled or not.
	// This default to false as the compatibility generator is enabled by default.
	Disabled bool

	// Verify determines whether the generator should verify the content instead
	// of updating the generated file.
	Verify bool
}

// generator implements the generation.Generator interface.
// It is designed to generate compatibility level comments for a particular API group.
type generator struct {
	disabled bool
	verify   bool
}

// NewGenerator builds a new compatibility generator.
func NewGenerator(opts Options) generation.Generator {
	return &generator{
		disabled: opts.Disabled,
		verify:   opts.Verify,
	}
}

// ApplyConfig creates returns a new generator based on the configuration passed.
// If the compatibility configuration is empty, the existing generation is returned.
func (g *generator) ApplyConfig(config *generation.Config) generation.Generator {
	if config == nil || config.Compatibility == nil {
		return g
	}

	return NewGenerator(Options{
		Disabled: config.Compatibility.Disabled,
		Verify:   g.verify,
	})
}

// Name returns the name of the generator.
func (g *generator) Name() string {
	return "compatibility"
}

// GenGroup runs the compatibility generator against the given group context.
func (g *generator) GenGroup(groupCtx generation.APIGroupContext) ([]generation.Result, error) {
	if g.disabled {
		klog.V(2).Infof("Skipping compatibility generation for %s", groupCtx.Name)
		return nil, nil
	}

	for _, version := range groupCtx.Versions {
		action := "Generating"
		if g.verify {
			action = "Verifying"
		}

		klog.V(1).Infof("%s compatibility level comments for %s/%s", action, groupCtx.Name, version.Name)

		if err := insertCompatibilityLevelComments(version.Path, g.verify); err != nil {
			return nil, fmt.Errorf("could not insert compatibility level comments for %s/%s: %w", groupCtx.Name, version.Name, err)
		}
	}

	return nil, nil
}
