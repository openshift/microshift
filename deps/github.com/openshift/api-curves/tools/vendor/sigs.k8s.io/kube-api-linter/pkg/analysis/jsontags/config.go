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
package jsontags

// JSONTagsConfig contains configuration for the jsontags linter.
type JSONTagsConfig struct {
	// jsonTagRegex is the regular expression used to validate that json tags are in a particular format.
	// By default, the regex used is "^[a-z][a-z0-9]*(?:[A-Z][a-z0-9]*)*$" and is used to check for
	// camel case like string.
	JSONTagRegex string `json:"jsonTagRegex"`
}
