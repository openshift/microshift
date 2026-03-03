package generation

import (
	"fmt"

	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/namer"
	"k8s.io/gengo/v2/parser"
	"k8s.io/gengo/v2/types"
)

// Execute implements the target execution from gengo.
func Execute(p *parser.Parser, universe types.Universe, nameSystems namer.NameSystems, defaultSystem string, getTargets func(*generator.Context) []generator.Target, inputPaths []string) error {
	c, err := generator.NewContext(p, nameSystems, defaultSystem)
	if err != nil {
		return fmt.Errorf("failed making a context: %v", err)
	}

	// Use the pre-extracted universe since we can only extract the universe from the parser once.
	c.Universe = universe

	// Limit to just the path for this API version.
	c.Inputs = inputPaths

	// Handle namesystems and extracting the types that relate to this API from the universe.
	// Copied from the generator.NewContext function.
	for name, systemNamer := range nameSystems {
		c.Namers[name] = systemNamer
		if name == defaultSystem {
			orderer := namer.Orderer{Namer: systemNamer}
			c.Order = orderer.OrderUniverse(universe)
		}
	}

	targets := getTargets(c)
	if err := c.ExecuteTargets(targets); err != nil {
		return fmt.Errorf("failed executing generator: %v", err)
	}

	return nil
}
