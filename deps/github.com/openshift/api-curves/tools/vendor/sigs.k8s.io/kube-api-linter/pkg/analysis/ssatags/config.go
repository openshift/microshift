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
package ssatags

// SSATagsConfig contains configuration for the ssatags linter.
type SSATagsConfig struct {
	// listTypeSetUsage is the policy for the listType=set usage.
	// Valid values are "Warn" and "Ignore".
	// When set to "Warn", the linter will emit a warning if a listType=set is used on object arrays.
	// When set to "Ignore", the linter will not emit a warning if a listType=set is used on object arrays.
	// Note: listType=set is only flagged on object arrays, not primitive arrays, due to
	// Server-Side Apply compatibility issues specific to object arrays.
	// When otherwise not specified, the default value is "Warn".
	ListTypeSetUsage SSATagsListTypeSetUsage `json:"listTypeSetUsage"`
}

// SSATagsListTypeSetUsage is the policy for the listType=set usage in the ssatags linter.
type SSATagsListTypeSetUsage string

const (
	// SSATagsListTypeSetUsageWarn indicates that the linter will emit a warning if a listType=set is used on object arrays.
	SSATagsListTypeSetUsageWarn SSATagsListTypeSetUsage = "Warn"

	// SSATagsListTypeSetUsageIgnore indicates that the linter will not emit a warning if a listType=set is used on object arrays.
	SSATagsListTypeSetUsageIgnore SSATagsListTypeSetUsage = "Ignore"
)
