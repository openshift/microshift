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

/*
defaults is a linter to check that fields with default markers are configured correctly.

Fields with default markers (+default, +kubebuilder:default, or +k8s:default) should also be marked as optional.
Additionally, fields with default markers should have "omitempty" or "omitzero" in their json tags
to ensure that the default values are applied correctly during serialization and deserialization.

Example of a well-configured field with a default:

	// +optional
	// +default="default-value"
	Field string `json:"field,omitempty"`

Example of issues this linter will catch:

	// Missing optional marker
	// +default="value"
	Field string `json:"field,omitempty"`

	// Missing omitempty tag
	// +optional
	// +default="value"
	Field string `json:"field"`
*/
package defaults
