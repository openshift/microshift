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
package conflictingmarkers

// ConflictingMarkersConfig contains the configuration for the conflictingmarkers linter.
type ConflictingMarkersConfig struct {
	// Conflicts allows users to define sets of conflicting markers.
	// Each entry defines a conflict between multiple sets of markers.
	Conflicts []ConflictSet `json:"conflicts"`
}

// ConflictSet represents a conflict between multiple sets of markers.
// Markers within each set are mutually exclusive with markers in all other sets.
// The linter will emit a diagnostic when a field has markers from two or more sets.
type ConflictSet struct {
	// Name is a human-readable name for this conflict set.
	// This name will appear in diagnostic messages to identify the type of conflict.
	Name string `json:"name"`
	// Sets contains the sets of markers that are mutually exclusive with each other.
	// Each set is a slice of marker identifiers.
	// The linter will emit a diagnostic when a field has markers from two or more sets.
	Sets [][]string `json:"sets"`
	// Description provides a description of why these markers conflict.
	// The linter will include this description in the diagnostic message when a conflict is detected.
	Description string `json:"description"`
}
