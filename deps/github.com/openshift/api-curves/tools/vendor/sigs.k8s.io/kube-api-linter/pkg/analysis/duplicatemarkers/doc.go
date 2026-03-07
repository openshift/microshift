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
duplicatemarkers is an analyzer that checks for duplicate markers in the API types.
It reports exact matches for marker definitions.

For example, something like:

	type Foo struct {
		// +kubebuilder:validation:MaxLength=10
		// +kubebuilder:validation:MaxLength=11
		type Bar string
	}

would not be reported while something like:

	type Foo struct {
		// +kubebuilder:validation:MaxLength=10
		// +kubebuilder:validation:MaxLength=10
		type Bar string
	}

would be reported.

This linter also be able to automatically fix all markers that are exact match to another markers.
If there are duplicates across fields and their underlying type, the marker on the type will be preferred and the marker on the field will be removed.
*/
package duplicatemarkers
