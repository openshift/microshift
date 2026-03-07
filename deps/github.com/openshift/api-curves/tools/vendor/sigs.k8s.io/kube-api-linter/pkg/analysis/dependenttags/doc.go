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
// Package dependenttags enforces dependencies between markers.
//
// # Analyzer dependenttags
//
// The dependenttags analyzer validates that if a specific marker (identifier) is present on a field,
// a set of other markers (dependent tags) are also present. This is useful for enforcing API
// contracts where certain markers imply the presence of others.
//
// For example, a field marked with `+k8s:unionMember` must also be marked with `+k8s:optional`.
//
// # Configuration
//
// The linter is configured with a list of rules. Each rule specifies an identifier marker and a list of
// dependent markers. The `type` field is required and specifies how to interpret the dependsOn list:
// - `All`: all dependent markers are required.
// - `Any`: at least one of the dependent markers is required.
//
// This linter only checks for the presence or absence of markers; it does not inspect or enforce specific values within those markers. It also does not provide automatic fixes.
//
//	linters:
//	  dependenttags:
//	    rules:
//	    - identifier: "k8s:unionMember"
//	      type: "All"
//	      dependsOn:
//	      - "k8s:optional"
//	    - identifier: "listType"
//	      type: "All"
//	      dependsOn:
//	      - "k8s:listType"
//	    - identifier: "example:any"
//	      type: "Any"
//	      dependsOn:
//	      - "dep1"
//	      - "dep2"
package dependenttags
