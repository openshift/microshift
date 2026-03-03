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
package conditions

// ConditionsFirstField is the policy for the conditions linter.
type ConditionsFirstField string

const (
	// ConditionsFirstFieldWarn indicates that the conditions should be the first field in the struct.
	ConditionsFirstFieldWarn ConditionsFirstField = "Warn"

	// ConditionsFirstFieldIgnore indicates that the conditions do not need to be the first field in the struct.
	ConditionsFirstFieldIgnore ConditionsFirstField = "Ignore"
)

// ConditionsUseProtobuf is the policy for the conditions linter.
type ConditionsUseProtobuf string

const (
	// ConditionsUseProtobufSuggestFix indicates that the linter will emit a warning if the conditions are not using protobuf tags and suggest a fix.
	ConditionsUseProtobufSuggestFix ConditionsUseProtobuf = "SuggestFix"

	// ConditionsUseProtobufWarn indicates that the linter will emit a warning if the conditions are not using protobuf tags.
	ConditionsUseProtobufWarn ConditionsUseProtobuf = "Warn"

	// ConditionsUseProtobufIgnore indicates that the linter will not emit a warning if the conditions are not using protobuf tags.
	ConditionsUseProtobufIgnore ConditionsUseProtobuf = "Ignore"

	// ConditionsUseProtobufForbid indicates that the linter will emit an error if the conditions are using protobuf tags, a fix will also be suggested.
	ConditionsUseProtobufForbid ConditionsUseProtobuf = "Forbid"
)

// ConditionsUsePatchStrategy is the policy for the conditions linter.
type ConditionsUsePatchStrategy string

const (
	// ConditionsUsePatchStrategySuggestFix indicates that the linter will emit a warning if the conditions are not using patch strategy tags and suggest a fix.
	ConditionsUsePatchStrategySuggestFix ConditionsUsePatchStrategy = "SuggestFix"

	// ConditionsUsePatchStrategyWarn indicates that the linter will emit a warning if the conditions are not using patch strategy tags.
	ConditionsUsePatchStrategyWarn ConditionsUsePatchStrategy = "Warn"

	// ConditionsUsePatchStrategyIgnore indicates that the linter will not emit a warning if the conditions are not using patch strategy tags.
	ConditionsUsePatchStrategyIgnore ConditionsUsePatchStrategy = "Ignore"

	// ConditionsUsePatchStrategyForbid indicates that the linter will emit an error if the conditions are using patch strategy tags, a fix will also be suggested.
	ConditionsUsePatchStrategyForbid ConditionsUsePatchStrategy = "Forbid"
)

// ConditionsConfig contains configuration for the conditions linter.
type ConditionsConfig struct {
	// isFirstField indicates whether the conditions should be the first field in the struct.
	// Valid values are Warn and Ignore.
	// When set to Warn, the linter will emit a warning if the conditions are not the first field in the struct.
	// When set to Ignore, the linter will not emit a warning if the conditions are not the first field in the struct.
	// When otherwise not specified, the default value is Warn.
	IsFirstField ConditionsFirstField `json:"isFirstField"`

	// useProtobuf indicates whether the linter should use protobuf tags.
	// Valid values are SuggestFix, Warn, Ignore and Forbid.
	// When set to SuggestFix, the linter will emit a warning if the conditions are not using protobuf tags and suggest a fix.
	// When set to Warn, the linter will emit a warning if the conditions are not using protobuf tags.
	// When set to Ignore, the linter will not emit a warning if the conditions are not using protobuf tags.
	// When set to Forbid, the linter will emit an error if the conditions are using protobuf tags, a fix will also be suggested.
	// When otherwise not specified, the default value is SuggestFix.
	UseProtobuf ConditionsUseProtobuf `json:"useProtobuf"`

	// usePatchStrategy indicates whether the linter should enforce the patch strategy tags.
	// Valid values are SuggestFix, Warn, Ignore and Forbid.
	// When set to SuggestFix, the linter will emit a warning if the conditions are not using patch strategy tags and suggest a fix.
	// When set to Warn, the linter will emit a warning if the conditions are not using patch strategy tags.
	// When set to Ignore, the linter will not emit a warning if the conditions are not using patch strategy tags.
	// When set to Forbid, the linter will emit an error if the conditions are using patch strategy tags, a fix will also be suggested.
	// When otherwise not specified, the default value is SuggestFix.
	UsePatchStrategy ConditionsUsePatchStrategy `json:"usePatchStrategy"`
}
