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
package forbiddenmarkers

// Config is the configuration type
// for the forbiddenmarkers linter.
type Config struct {
	// markers is the unique set of markers
	// that are forbidden on types/fields.
	// Uniqueness is keyed on the `identifier`
	// field of entries.
	// Must have at least one entry.
	Markers []Marker `json:"markers"`
}

// Marker is a representation of a
// type/field marker that should be forbidden.
type Marker struct {
	// identifier is the identifier for the forbidden marker.
	Identifier string `json:"identifier"`

	// ruleSets is an optional set of rules that are used to determine
	// if a marker definition is forbidden.
	// When specified, if an instance of a marker with the identifier
	// specified in 'identifier' satisfies at least one of the rule sets
	// defined, it will be considered a forbidden marker definition.
	// When not specified, any instances of a marker with the identifier
	// specified in 'identifier' will be considered a forbidden marker
	// definition.
	RuleSets []RuleSet `json:"ruleSets,omitempty"`
}

// RuleSet is a representation of a
// set of rules that applies to a marker
// when determining if it should be forbidden.
type RuleSet struct {
	// attributes is a unique set of
	// attributes that is forbidden for this marker.
	// Uniqueness is keyed on the `name` field of entries.
	// When specified, only instances of this marker
	// that contains all the attributes will be considered
	// forbidden.
	Attributes []MarkerAttribute `json:"attributes,omitempty"`
}

// MarkerAttribute is a representation of an
// attribute for a marker.
type MarkerAttribute struct {
	// name is the name of the forbidden attribute
	Name string `json:"name"`
	// values is an optional unique set of
	// values that are forbidden for this marker.
	// When specified, only the instances of this
	// attribute that set one of these forbidden values
	// will be considered forbidden.
	// When not specified, any use of this attribute
	// will be considered forbidden.
	Values []string `json:"values,omitempty"`
}
