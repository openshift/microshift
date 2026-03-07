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
package optionalfields

// OptionalFieldsConfig is the configuration for the optionalfields linter.
type OptionalFieldsConfig struct {
	// pointers is the policy for pointers in optional fields.
	// This defines how the linter should handle optional fields, and whether they should be pointers or not.
	// By default, all fields will be expected to be pointers, and the linter will suggest fixes if they are not.
	Pointers OptionalFieldsPointers `json:"pointers"`

	// omitempty is the policy for the `omitempty` tag within the json tag for fields.
	// This defines how the linter should handle optional fields, and whether they should have the omitempty tag or not.
	// By default, all fields will be expected to have the `omitempty` tag.
	OmitEmpty OptionalFieldsOmitEmpty `json:"omitempty"`

	// omitzero is the policy for the `omitzero` tag within the json tag for fields.
	// This defines how the linter should handle optional fields, and whether they should have the omitzero tag or not.
	// By default, all the struct fields will be expected to have the `omitzero` tag when their zero value is not an acceptable user choice.
	OmitZero OptionalFieldsOmitZero `json:"omitzero"`
}

// OptionalFieldsPointers is the configuration for pointers in optional fields.
type OptionalFieldsPointers struct {
	// preference determines whether the linter should prefer pointers for all optional fields,
	// or only for optional fields where validation or serialization requires a pointer.
	// Valid values are "Always" and "WhenRequired".
	// When set to "Always", the linter will prefer pointers for all optional fields.
	// When set to "WhenRequired", the linter will prefer pointers for optional fields where validation or serialization requires a pointer.
	// The "WhenRequired" option requires bounds on strings and numerical values to be able to accurately determine the correct pointer vs non-pointer decision.
	// When otherwise not specified, the default value is "Always".
	Preference OptionalFieldsPointerPreference `json:"preference"`
	// policy is the policy for the pointer preferences for optional fields.
	// Valid values are "SuggestFix" and "Warn".
	// When set to "SuggestFix", the linter will emit a warning if the pointer preference is not followed and suggest a fix.
	// When set to "Warn", the linter will emit a warning if the pointer preference is not followed.
	// When otherwise not specified, the default value is "SuggestFix".
	Policy OptionalFieldsPointerPolicy `json:"policy"`
}

// OptionalFieldsOmitEmpty is the configuration for the `omitempty` tag on optional fields.
type OptionalFieldsOmitEmpty struct {
	// policy determines whether the linter should require omitempty for all optional fields.
	// Valid values are "SuggestFix", "Warn" and "Ignore".
	// When set to "SuggestFix", the linter will suggest adding the `omitempty` tag when an optional field does not have it.
	// When set to "Warn", the linter will emit a warning if the field does not have the `omitempty` tag.
	// When set to "Ignore", and optional field missing the `omitempty` tag will be ignored.
	// Note, when set to "Ignore", and a field does not have the `omitempty` tag, this may affect whether the field should be a pointer or not.
	Policy OptionalFieldsOmitEmptyPolicy `json:"policy"`
}

// OptionalFieldsOmitZero is the configuration for the `omitzero` tag on optional fields.
type OptionalFieldsOmitZero struct {
	// policy determines whether the linter should require omitzero for all optional `struct` fields.
	// Valid values are "SuggestFix", "Warn" and "Forbid".
	// When set to "SuggestFix", the linter will suggest adding the `omitzero` tag when an optional field does not have it.
	// When set to "Warn", the linter will emit a warning if the field does not have the `omitzero` tag.
	// When set to "Forbid", 'omitzero' tags wont be considered.
	// Note, when set to "Forbid", and a field have the `omitzero` tag, the linter will suggest to remove the `omitzero` tag.
	// Note, `omitzero` tag is supported in go version starting from go 1.24.
	// Note, Configure omitzero policy to 'Forbid', if using with go version less than go 1.24.
	Policy OptionalFieldsOmitZeroPolicy `json:"policy"`
}

// OptionalFieldsPointerPreference is the preference for pointers in optional fields.
type OptionalFieldsPointerPreference string

const (
	// OptionalFieldsPointerPreferenceAlways indicates that the linter should prefer pointers for all optional fields.
	OptionalFieldsPointerPreferenceAlways OptionalFieldsPointerPreference = "Always"

	// OptionalFieldsPointerPreferenceWhenRequired indicates that the linter should prefer pointers for optional fields where validation or serialization requires a pointer.
	OptionalFieldsPointerPreferenceWhenRequired OptionalFieldsPointerPreference = "WhenRequired"
)

// OptionalFieldsPointerPolicy is the policy for pointers in optional fields.
type OptionalFieldsPointerPolicy string

const (
	// OptionalFieldsPointerPolicySuggestFix indicates that the linter will emit a warning if the pointer preference is not followed and suggest a fix.
	OptionalFieldsPointerPolicySuggestFix OptionalFieldsPointerPolicy = "SuggestFix"

	// OptionalFieldsPointerPolicyWarn indicates that the linter will emit a warning if the pointer preference is not followed.
	OptionalFieldsPointerPolicyWarn OptionalFieldsPointerPolicy = "Warn"
)

// OptionalFieldsOmitEmptyPolicy is the policy for the omitempty tag on optional fields.
type OptionalFieldsOmitEmptyPolicy string

const (
	// OptionalFieldsOmitEmptyPolicySuggestFix indicates that the linter will emit a warning if the field does not have omitempty, and suggest a fix.
	OptionalFieldsOmitEmptyPolicySuggestFix OptionalFieldsOmitEmptyPolicy = "SuggestFix"

	// OptionalFieldsOmitEmptyPolicyWarn indicates that the linter will emit a warning if the field does not have omitempty.
	OptionalFieldsOmitEmptyPolicyWarn OptionalFieldsOmitEmptyPolicy = "Warn"

	// OptionalFieldsOmitEmptyPolicyIgnore indicates that the linter will ignore any field missing the omitempty tag.
	OptionalFieldsOmitEmptyPolicyIgnore OptionalFieldsOmitEmptyPolicy = "Ignore"
)

// OptionalFieldsOmitZeroPolicy is the policy for the omitzero tag on optional fields.
type OptionalFieldsOmitZeroPolicy string

const (
	// OptionalFieldsOmitZeroPolicySuggestFix indicates that the linter will emit a warning if the field does not have omitzero, and suggest a fix.
	OptionalFieldsOmitZeroPolicySuggestFix OptionalFieldsOmitZeroPolicy = "SuggestFix"

	// OptionalFieldsOmitZeroPolicyWarn indicates that the linter will emit a warning if the field does not have omitzero.
	OptionalFieldsOmitZeroPolicyWarn OptionalFieldsOmitZeroPolicy = "Warn"

	// OptionalFieldsOmitZeroPolicyForbid indicates that the linter will forbid using omitzero tag.
	OptionalFieldsOmitZeroPolicyForbid OptionalFieldsOmitZeroPolicy = "Forbid"
)
