package generation

// Generator is an interface for running a generator against a particular API group.
type Generator interface {
	// Name returns a name identifier for the generator.
	Name() string

	// GenGroup runs the generator against the given APIGroupContext.
	GenGroup(APIGroupContext) ([]Result, error)

	// ApplyConfig creates a new generator instance with the given configuration.
	ApplyConfig(*Config) Generator
}

// MultiGroupGenerator is an interface for running a generator against multiple API groups.
// This is used for generators that the the context of multiple groups simultaneously.
type MultiGroupGenerator interface {
	// Name returns a name identifier for the generator.
	Name() string

	// GenGroups runs the generator against the given APIGroupContexts.
	GenGroups([]APIGroupContext) error
}
