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
inspector is a helper package that iterates over fields in structs, calling an inspection function on fields
that should be considered for analysis.

The inspector extracts common logic of iterating and filtering through struct fields, so that analyzers
need not re-implement the same filtering over and over.

For example, the inspector filters out struct definitions that are not type declarations, and fields that are ignored.

Example:

	type A struct {
		// This field is included in the analysis.
		Field string `json:"field"`

		// This field, and the fields within are ignored due to the json tag.
		F struct {
			Field string `json:"field"`
		} `json:"-"`
	}

	// Any struct defined within a function is ignored.
	func Foo() {
		type Bar struct {
			Field string
		}
	}

	// All fields within interface declarations are ignored.
	type Bar interface {
		Name() string
	}
*/
package inspector
