package generation

import (
	crdifycfg "sigs.k8s.io/crdify/pkg/config"
)

// Config represents the configuration of a API group version
// and the configuration for each generator within it.
type Config struct {
	// Compatibility represents the configuration of the compatiblity generator.
	// When omitted, the default configuration will be used.
	Compatibility *CompatibilityConfig `json:"compatibility,omitempty"`

	// Crdify represents the configuration of the crdify generator.
	// When omitted, the default configuration will be used.
	Crdify *CrdifyConfig `json:"crdify,omitempty"`

	// Deepcopy represents the configuration of the deepcopy generator.
	// When omitted, the default configuration will be used.
	Deepcopy *DeepcopyConfig `json:"deepcopy,omitempty"`

	// OpenAPI represents the configuration of the openapi generator.
	// When omitted, the default configuration will be used.
	OpenAPI *OpenAPIConfig `json:"openapi,omitempty"`

	// SchemaCheck represents the configuration for the schemacheck generator.
	// When omitted, the default configuration will be used.
	// When provided, any equivalent flag provided values are ignored.
	SchemaCheck *SchemaCheckConfig `json:"schemacheck,omitempty"`

	// SchemaPatch represents the configuration for the schemapatch generator.
	// When omitted, the default configuration will be used.
	// When provided, any equivalent flag provided values are ignored.
	SchemaPatch *SchemaPatchConfig `json:"schemapatch,omitempty"`

	// ManifestMerge represents the configuration for the manifest merge generator.
	// When omitted, the default configuration will be used.
	// When provided, any equivalent flag provided values are ignored.
	ManifestMerge *ManifestMerge `json:"manifestMerge,omitempty"`

	// EmptyPartialSchema represents the configuration for the manifest merge generator.
	// When omitted, the default configuration will be used.
	// When provided, any equivalent flag provided values are ignored.
	EmptyPartialSchema *EmptyPartialSchema `json:"emptyPartialSchema,omitempty"`

	// SwaggerDocs represents the configuration for the swaggerdocs generator.
	// When omitted, the default configuration will be used.
	// When provided, any equivalent flag provided values are ignored.
	SwaggerDocs *SwaggerDocsConfig `json:"swaggerdocs,omitempty"`
}

// CompatibilityConfig is the configuration for the compatibility generator.
type CompatibilityConfig struct {
	// Disabled determines whether the compatibility generator should be run or not.
	// This generator is enabled by default so this field defaults to false.
	Disabled bool `json:"disabled,omitempty"`
}

type CrdifyConfig struct {
	// Disabled determines whether the crdify generator should be run or not.
	// This generator is enabled by default so this field defaults to false.
	Disabled bool `json:"disabled,omitempty"`

	// Config configures the validations that crdify performs and how they should be run.
	// When omitted, a default configuration is used.
	Config *crdifycfg.Config `json:"config,omitempty"`
}

// DeepcopyConfig is the configuration for the deepcopy generator.
type DeepcopyConfig struct {
	// Disabled determines whether the deepcopy generator should be run or not.
	// This generator is enabled by default so this field defaults to false.
	Disabled bool `json:"disabled,omitempty"`

	// HeaderFilePath is the path to the file containing the boilerplate header text.
	// When omitted, no header is added to the generated files.
	HeaderFilePath string `json:"headerFilePath,omitempty"`

	// OutputFileBaseName is the base name of the output file.
	// When omitted, DefaultOutputFileBaseName is used.
	// The current value of DefaultOutputFileBaseName is "zz_generated.deepcopy".
	OutputFileBaseName string `json:"outputFileBaseName,omitempty"`
}

// OpenAPIConfig is the configuration for the openapi generator.
type OpenAPIConfig struct {
	// Disabled determines whether the openapi generator should include this
	// group or not.
	// This generator is enabled by default so this field defaults to false.
	Disabled bool `json:"disabled,omitempty"`
}

// SchemaCheckConfig is the configuration for the schemacheck generator.
type SchemaCheckConfig struct {
	// Disabled determines whether the schemacheck generator should be run or not.
	// This generator is enabled by default so this field defaults to false.
	Disabled bool `json:"disabled,omitempty"`

	// EnabledValidators is a list of the validators that should be enabled.
	// If this is empty, the default validators are enabled.
	EnabledValidators []string `json:"enabledValidators,omitempty"`

	// DisabledValidators is a list of the validators that should be disabled.
	// If this is empty, no default validators are disabled.
	DisabledValidators []string `json:"disabledValidators,omitempty"`
}

// SchemaPatchConfig is the configuration for the schemapatch generator.
type SchemaPatchConfig struct {
	// Disabled determines whether the schemapatch generator should be run or not.
	// This generator is enabled by default so this field defaults to false.
	Disabled bool `json:"disabled,omitempty"`

	// RequiredFeatureSets is a list of feature sets combinations that should be
	// generated for this API group.
	// Each entry in this list is a comma separated list of feature set names
	// which will be matched with the `release.openshift.io/feature-set` annotation
	// on the CRD definition.
	// When omitted, any manifest with a feature set annotation will be ignored.
	// Example entries are `""` (empty string), `"TechPreviewNoUpgrade"` or `"TechPreviewNoUpgrade,CustomNoUpgrade"`.
	RequiredFeatureSets []string `json:"requiredFeatureSets,omitempty"`
}

// ManifestMerge is the configuration for the manifest merge generator.
type ManifestMerge struct {
	// Disabled determines whether the schemapatch generator should be run or not.
	// This generator is enabled by default so this field defaults to false.
	Disabled bool `json:"disabled,omitempty"`
}

// ManifestMerge is the configuration for the manifest merge generator.
type EmptyPartialSchema struct {
	// Disabled determines whether the schemapatch generator should be run or not.
	// This generator is enabled by default so this field defaults to false.
	Disabled bool `json:"disabled,omitempty"`
}

// SwaggerDocsConfig is the configuration for the swaggerdocs generator.
type SwaggerDocsConfig struct {
	// Disabled determines whether the swaggerdocs generator should be run or not.
	// This generator is enabled by default so this field defaults to false.
	Disabled bool `json:"disabled,omitempty"`

	// CommentPolicy determines how, when verifying swaggerdocs, the generator
	// should handle missing comments.
	// Valid values are `Ignore`, `Warn` and `Enforce`.
	// This defaults to `Warn`.
	// When set to `Ignore`, the generator will ignore any missing comments.
	// When set to `Warn`, the generator will emit a warning for any missing comments.
	// When set to `Enforce`, the generator will return an error for any missing comments.
	CommentPolicy string `json:"commentPolicy,omitempty"`

	// OutputFileName is the file name to use for writing the generated swagger
	// docs to. This file will be created for each group version.
	// Whem omitted, this will default to `zz_generated.swagger_doc_generated.go`.
	OutputFileName string `json:"outputFileName,omitempty"`
}
