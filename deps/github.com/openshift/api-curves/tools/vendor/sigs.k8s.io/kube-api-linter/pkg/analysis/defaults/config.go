/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package defaults

// OmitEmptyPolicy is the policy for omitempty.
// SuggestFix will suggest a fix for the field to add omitempty.
// Warn will warn about the field to add omitempty.
// Ignore will ignore the absence of omitempty.
type OmitEmptyPolicy string

const (
	// OmitEmptyPolicySuggestFix will suggest a fix for the field.
	OmitEmptyPolicySuggestFix OmitEmptyPolicy = "SuggestFix"

	// OmitEmptyPolicyWarn will warn about the field.
	OmitEmptyPolicyWarn OmitEmptyPolicy = "Warn"

	// OmitEmptyPolicyIgnore will ignore the field.
	OmitEmptyPolicyIgnore OmitEmptyPolicy = "Ignore"
)

// OmitZeroPolicy is the policy for omitzero.
// SuggestFix will suggest a fix for the field to add omitzero.
// Warn will warn about the field to add omitzero.
// Forbid will forbid the field to have omitzero.
type OmitZeroPolicy string

const (
	// OmitZeroPolicySuggestFix will suggest a fix for the field.
	OmitZeroPolicySuggestFix OmitZeroPolicy = "SuggestFix"

	// OmitZeroPolicyWarn will warn about the field.
	OmitZeroPolicyWarn OmitZeroPolicy = "Warn"

	// OmitZeroPolicyForbid will forbid the field.
	OmitZeroPolicyForbid OmitZeroPolicy = "Forbid"
)

// DefaultsConfig contains configuration for the defaults linter.
type DefaultsConfig struct {
	// PreferredDefaultMarker is the preferred marker to use for default values.
	// If this field is not set, the default value is "default".
	// Valid values are "default" and "kubebuilder:default".
	PreferredDefaultMarker string `json:"preferredDefaultMarker"`

	// OmitEmpty is the configuration for the `omitempty` tag within the json tag for fields with defaults.
	// This defines how the linter should handle fields with defaults, and whether they should have the omitempty tag or not.
	// By default, all fields with defaults will be expected to have the `omitempty` tag.
	OmitEmpty DefaultsOmitEmpty `json:"omitempty"`

	// OmitZero is the configuration for the `omitzero` tag within the json tag for fields with defaults.
	// This defines how the linter should handle fields with defaults, and whether they should have the omitzero tag or not.
	// By default, struct fields with defaults will be expected to have the `omitzero` tag.
	OmitZero DefaultsOmitZero `json:"omitzero"`
}

// DefaultsOmitEmpty is the configuration for the `omitempty` tag on fields with defaults.
type DefaultsOmitEmpty struct {
	// Policy determines whether the linter should require omitempty for fields with defaults.
	// Valid values are "SuggestFix", "Warn" and "Ignore".
	// When set to "SuggestFix", the linter will suggest adding the `omitempty` tag when a field with default does not have it.
	// When set to "Warn", the linter will emit a warning if the field does not have the `omitempty` tag.
	// When set to "Ignore", a field with default missing the `omitempty` tag will be ignored.
	// When otherwise not specified, the default value is "SuggestFix".
	Policy OmitEmptyPolicy `json:"policy"`
}

// DefaultsOmitZero is the configuration for the `omitzero` tag on fields with defaults.
type DefaultsOmitZero struct {
	// Policy determines whether the linter should require omitzero for struct fields with defaults.
	// Valid values are "SuggestFix", "Warn" and "Forbid".
	// When set to "SuggestFix", the linter will suggest adding the `omitzero` tag when a struct field with default does not have it.
	// When set to "Warn", the linter will emit a warning if the field does not have the `omitzero` tag.
	// When set to "Forbid", 'omitzero' tags will not be considered.
	// Note, when set to "Forbid", and a field have the `omitzero` tag, the linter will not suggest adding it.
	// Note, `omitzero` tag is supported in go version starting from go 1.24.
	// Note, Configure omitzero policy to 'Forbid', if using with go version less than go 1.24.
	// When otherwise not specified, the default value is "SuggestFix".
	Policy OmitZeroPolicy `json:"policy"`
}
