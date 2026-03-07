package generation

type Result struct {
	// Generator is the name of the generator that produced this result.
	Generator string

	// Group is the name of the API group that was processed.
	Group string

	// Version is the version of the API group that was processed.
	Version string

	// Manifest is the path to the file the was processed.
	Manifest string

	// Info contains informational messages from the generator.
	Info []string

	// Warnings contains warning messages from the generator.
	Warnings []string

	// Errors contains error messages from the generator.
	Errors []error
}
