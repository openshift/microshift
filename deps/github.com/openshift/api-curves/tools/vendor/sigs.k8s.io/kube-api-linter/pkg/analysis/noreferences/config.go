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
package noreferences

// Policy defines the policy for handling references in field names.
type Policy string

const (
	// PolicyPreferAbbreviatedReference allows abbreviated forms (Ref/Refs) in field names.
	// It suggests replacing Reference/References with Ref/Refs.
	PolicyPreferAbbreviatedReference Policy = "PreferAbbreviatedReference"
	// PolicyNoReferences forbids any reference-related words in field names.
	// It suggests removing Ref/Refs/Reference/References entirely.
	PolicyNoReferences Policy = "NoReferences"
)

// Config represents the configuration for the noreferences linter.
type Config struct {
	// policy controls how reference-related words are handled in field names.
	// When set to PreferAbbreviatedReference (default), Reference/References are replaced with Ref/Refs.
	// When set to NoReferences, all reference-related words are suggested to be removed.
	Policy Policy `json:"policy,omitempty"`
}
