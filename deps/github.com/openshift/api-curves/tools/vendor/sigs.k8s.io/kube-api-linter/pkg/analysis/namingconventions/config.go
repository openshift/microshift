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
package namingconventions

// Config represents the configuration
// of the namingconventions linter.
type Config struct {
	// conventions is the set of field naming
	// conventions to be enforced by the namingconventions
	// linter.
	// At least one convention is required.
	// Conventions must be unique, keyed on convention names.
	Conventions []Convention `json:"conventions,omitempty"`
}

// Convention represents a naming convention.
type Convention struct {
	// name is a required human-readable name to
	// associate with the field naming convention.
	// name is case-sensitive.
	Name string `json:"name,omitempty"`

	// violationMatcher is a required RE2 compliant regular expression
	// used to identify violating portions of a field name.
	ViolationMatcher string `json:"violationMatcher,omitempty"`

	// operation is a required configuration that tells
	// the namingconventions linter how violations
	// of this convention should be handled.
	//
	// Allowed values are Inform, DropField, Drop, and Replace.
	//
	// When set to Inform, the namingconventions linter will
	// inform when a field is violating the convention, but
	// will not suggest any fixes.
	//
	// When set to DropField, the namingconventions linter will
	// suggest that any fields matching this convention should
	// be removed in their entirety.
	//
	// When set to Drop, the namingconventions linter will suggest
	// that any fields matching this convention should drop the
	// portion of the field name matched by the violationMatcher
	// expression.
	//
	// When set to Replacement, the namingconventions linter will
	// suggest that any fields matching this conventions should
	// replace the matched portion of the field name with the value
	// specified in the replace field.
	Operation Operation `json:"operation,omitempty"`

	// replacement configures the string that should
	// replace the matched portion of a field name
	// that violates this conventions.
	// replacement is required when operation is set to Replacement
	// and forbidden otherwise.
	Replacement string `json:"replacement,omitempty,omitzero"`

	// message is a required human-readable message
	// to be included in the linter error if a field
	// is found to violate this naming convention.
	Message string `json:"message,omitempty"`
}

// Operation is a reference to the operation that should be used
// when evaluating a naming convention.
type Operation string

const (
	// OperationDropField signals that an entire field
	// should be removed when the naming convention is violated.
	OperationDropField Operation = "DropField"

	// OperationDrop signals that the offending text
	// should be removed from the field name when
	// the naming convention is violated.
	OperationDrop Operation = "Drop"

	// OperationReplacement signals that the offending text
	// should be replaced in the field name when
	// the naming convention is violated.
	OperationReplacement Operation = "Replacement"

	// OperationInform signals that no action
	// should be taken, beyond issuing a linter error
	// when the naming convention is violated.
	OperationInform Operation = "Inform"
)
