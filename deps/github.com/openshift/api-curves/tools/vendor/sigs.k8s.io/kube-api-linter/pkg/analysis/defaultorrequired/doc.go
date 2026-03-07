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
Package defaultorrequired provides the defaultorrequired analyzer.

The defaultorrequired analyzer checks that fields marked as required do not have default values applied.

A field cannot be both required and have a default value, as these are conflicting concepts:
- A required field must be provided by the user and cannot be omitted
- A default value is used when a field is not provided

For example, the following would be flagged:

	// +kubebuilder:validation:Required
	// +kubebuilder:default:=value
	Field string `json:"field"`
*/
package defaultorrequired
