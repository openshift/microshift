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
package uniquemarkers

// UniqueMarkersConfig contains the configuration for the uniquemarkers linter.
type UniqueMarkersConfig struct {
	// customMarkers is the set of custom marker/attribute combinations that
	// should not appear more than once on a type/field.
	// Entries must have unique identifiers.
	// Entries must start and end with alpha characters and must consist of only alpha characters and colons (':').
	CustomMarkers []UniqueMarker `json:"customMarkers"`
}

// UniqueMarker represents an instance of a marker that should
// be unique for a field/type. A marker consists
// of an identifier and attributes that can be used
// to dictate uniqueness.
type UniqueMarker struct {
	// identifier configures the marker identifier that should be unique.
	// Some common examples are "kubebuilder:validation:Enum" and "kubebuilder:validation:XValidation".
	Identifier string `json:"identifier"`
	// attributes configures the attributes that should be considered
	// as part of the uniqueness evaluation.
	// If an attribute in this list is not found in a marker definition,
	// it is interpreted as the empty value.
	//
	// Entries must be unique.
	Attributes []string `json:"attributes"`
}
