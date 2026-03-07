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
package requiredfields

// RequiredFieldsConfig contains configuration for the requiredfields linter.
type RequiredFieldsConfig struct {
	// pointers is the policy for pointers in required fields.
	// This will allow configuration of whether or not to suggest fixes for pointers in required fields,
	// or just warn about their absence.
	// Pointers on required fields are only recommended when the zero value is a valid user choice.
	Pointers RequiredFieldsPointers `json:"pointers"`

	// omitempty is the policy for the `omitempty` tag on required fields.
	// This will allow configuration of whether or not to suggest fixes for the `omitempty` tag on required fields,
	// or just warn about their absence.
	// Required fields should have the `omitempty` tag to allow structured clients to make an explicit choice
	// to include the field.
	OmitEmpty RequiredFieldsOmitEmpty `json:"omitempty"`

	// omitzero is the policy for the `omitzero` tag within the json tag for fields.
	// This defines how the linter should handle optional fields, and whether they should have the omitzero tag or not.
	// By default, all the struct fields will be expected to have the `omitzero` tag when their zero value is not an acceptable user choice.
	// Note, `omitzero` tag is supported in go version starting from go 1.24.
	// Note, Configure omitzero policy to 'Forbid', if using with go version less than go 1.24.
	OmitZero RequiredFieldsOmitZero `json:"omitzero"`
}

// RequiredFieldsPointers is the configuration for pointers in required fields.
type RequiredFieldsPointers struct {
	// policy is the policy for the pointer preferences for required fields.
	// Valid values are "SuggestFix" and "Warn".
	// When set to "SuggestFix", the linter will emit a warning and suggest a fix if the field is a pointer and doesn't need to be, or, where it needs to be a pointer, but isn't.
	// When set to "Warn", the linter will emit a warning per the above, but without suggesting a fix.
	// When otherwise not specified, the default value is "SuggestFix".
	Policy RequiredFieldsPointerPolicy `json:"policy"`
}

// RequiredFieldsOmitEmpty is the configuration for the `omitempty` tag on required fields.
type RequiredFieldsOmitEmpty struct {
	// policy is the policy for the `omitempty` tag on required fields.
	// Valid values are "SuggestFix", "Warn" and "Ignore".
	// When set to "SuggestFix", the linter will suggest adding the `omitempty` tag when an required field does not have it.
	// When set to "Warn", the linter will emit a warning if the field does not have the `omitempty` tag.
	// When set to "Ignore", a required field missing the `omitempty` tag will be ignored.
	// Note, when set to "Ignore", and a field does not have the `omitempty` tag, this may affect whether the field should be a pointer or not.
	Policy RequiredFieldsOmitEmptyPolicy `json:"policy"`
}

// RequiredFieldsOmitZero is the configuration for the `omitzero` tag on required fields.
type RequiredFieldsOmitZero struct {
	// policy is the policy for the `omitzero` tag on required fields.
	// Valid values are "SuggestFix", "Warn" and "Forbid".
	// When set to "SuggestFix", the linter will suggest adding the `omitzero` tag when an required field does not have it.
	// When set to "Warn", the linter will emit a warning if the field does not have the `omitzero` tag.
	// When set to "Forbid", 'omitzero' tags will not be considered.
	// Note, when set to "Forbid", and a field have the `omitzero` tag, the linter will suggest to remove the `omitzero` tag.
	// Note, `omitzero` tag is supported in go version starting from go 1.24.
	// Note, Configure omitzero policy to 'Forbid', if using with go version less than go 1.24.
	Policy RequiredFieldsOmitZeroPolicy `json:"policy"`
}

// RequiredFieldsPointerPolicy is the policy for pointers in required fields.
type RequiredFieldsPointerPolicy string

const (
	// RequiredFieldsPointerPolicyWarn indicates that the linter will emit a warning if a required field is a pointer.
	RequiredFieldsPointerPolicyWarn RequiredFieldsPointerPolicy = "Warn"

	// RequiredFieldsPointerPolicySuggestFix indicates that the linter will emit a warning if a required field is a pointer and suggest a fix.
	RequiredFieldsPointerPolicySuggestFix RequiredFieldsPointerPolicy = "SuggestFix"
)

// RequiredFieldsOmitEmptyPolicy is the policy for the `omitempty` tag on required fields.
type RequiredFieldsOmitEmptyPolicy string

const (
	// RequiredFieldsOmitEmptyPolicySuggestFix indicates that the linter will emit a warning if the field does not have omitempty, and suggest a fix.
	RequiredFieldsOmitEmptyPolicySuggestFix RequiredFieldsOmitEmptyPolicy = "SuggestFix"

	// RequiredFieldsOmitEmptyPolicyWarn indicates that the linter will emit a warning if the field does not have omitempty.
	RequiredFieldsOmitEmptyPolicyWarn RequiredFieldsOmitEmptyPolicy = "Warn"

	// RequiredFieldsOmitEmptyPolicyIgnore indicates that a required field missing the `omitempty` tag will be ignored.
	RequiredFieldsOmitEmptyPolicyIgnore RequiredFieldsOmitEmptyPolicy = "Ignore"
)

// RequiredFieldsOmitZeroPolicy is the policy for the `omitzero` tag on required fields.
type RequiredFieldsOmitZeroPolicy string

const (
	// RequiredFieldsOmitZeroPolicySuggestFix indicates that the linter will suggest adding the `omitzero` tag when an required field does not have it.
	RequiredFieldsOmitZeroPolicySuggestFix RequiredFieldsOmitZeroPolicy = "SuggestFix"

	// RequiredFieldsOmitZeroPolicyWarn indicates that the linter will emit a warning if the field does not have omitzero.
	RequiredFieldsOmitZeroPolicyWarn RequiredFieldsOmitZeroPolicy = "Warn"

	// RequiredFieldsOmitZeroPolicyForbid indicates that the linter will not consider `omitzero` tags, and will suggest to remove the `omitzero` tag if a field has it.
	RequiredFieldsOmitZeroPolicyForbid RequiredFieldsOmitZeroPolicy = "Forbid"
)
