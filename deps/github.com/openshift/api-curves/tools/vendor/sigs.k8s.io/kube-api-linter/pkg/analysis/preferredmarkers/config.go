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
package preferredmarkers

// Config is the configuration type
// for the preferredmarkers linter.
type Config struct {
	// markers is the unique set of preferred markers
	// and their equivalent identifiers.
	// Uniqueness is keyed on the `preferredIdentifier`
	// field of entries.
	// Must have at least one entry.
	Markers []Marker `json:"markers"`
}

// Marker is a representation of a preferred marker
// and its equivalent identifiers that should be replaced.
type Marker struct {
	// preferredIdentifier is the identifier for the preferred marker.
	PreferredIdentifier string `json:"preferredIdentifier"`

	// equivalentIdentifiers is a unique set of marker identifiers
	// that are equivalent to the preferred identifier.
	// When any of these markers are found, they will be reported
	// and a fix will be suggested to replace them with the
	// preferred identifier.
	// Must have at least one entry.
	EquivalentIdentifiers []EquivalentIdentifier `json:"equivalentIdentifiers"`
}

// EquivalentIdentifier represents a marker identifier that should be
// replaced with the preferred identifier.
type EquivalentIdentifier struct {
	// identifier is the marker identifier to replace.
	Identifier string `json:"identifier"`
}
