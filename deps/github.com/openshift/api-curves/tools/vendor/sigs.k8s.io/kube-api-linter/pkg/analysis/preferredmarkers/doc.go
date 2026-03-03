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
* The `preferredmarkers` linter ensures that types and fields use preferred markers
* instead of equivalent but different marker identifiers.
*
* By default, `preferredmarkers` is not enabled.
*
* This linter is useful for projects that want to enforce consistent marker usage
* across their codebase, especially when multiple equivalent markers exist.
* For example, Kubernetes has multiple ways to mark fields as optional:
* - `+k8s:optional`
* - `+kubebuilder:validation:Optional`
*
* The linter can be configured to enforce using one preferred marker identifier
* and report any equivalent markers that should be replaced.
*
* **Configuration:**
*
* The linter requires a configuration that specifies preferred markers and their
* equivalent identifiers.
*
* **Scenario:** Enforce using `+k8s:optional` instead of `+kubebuilder:validation:Optional`
*
* ```yaml
* linterConfig:
*   preferredmarkers:
*     markers:
*       - preferredIdentifier: "k8s:optional"
*         equivalentIdentifiers:
*           - "kubebuilder:validation:Optional"
* ```
*
* **Scenario:** Enforce using a custom marker instead of multiple equivalent markers
*
* ```yaml
* linterConfig:
*   preferredmarkers:
*     markers:
*       - preferredIdentifier: "custom:preferred"
*         equivalentIdentifiers:
*           - "custom:old:marker"
*           - "custom:deprecated:marker"
*           - "custom:legacy:marker"
* ```
*
* **Scenario:** Multiple preferred markers with different equivalents
*
* ```yaml
* linterConfig:
*   preferredmarkers:
*     markers:
*       - preferredIdentifier: "k8s:optional"
*         equivalentIdentifiers:
*           - "kubebuilder:validation:Optional"
*       - preferredIdentifier: "k8s:required"
*         equivalentIdentifiers:
*           - "kubebuilder:validation:Required"
* ```
*
* **Behavior:**
*
* When one or more equivalent markers are found, the linter will:
* 1. Report a diagnostic message indicating which marker(s) should be preferred
* 2. Suggest a fix that:
*    - If the preferred marker does not already exist: replaces the first equivalent
*      marker with the preferred identifier and preserves any marker expressions
*      (e.g., `=value` or `:key=value`)
*    - If the preferred marker already exists: removes all equivalent markers to
*      avoid duplicates
*    - Removes any additional equivalent markers
*
* For example, if both `+kubebuilder:validation:Optional` and `+custom:optional`
* are configured as equivalents to `+k8s:optional`, they will both be replaced
* with a single `+k8s:optional` marker. If `+k8s:optional` already exists alongside
* equivalent markers, only the equivalent markers will be removed.
*
* The linter checks both type-level and field-level markers, including markers
* inherited from type aliases.
*
 */
package preferredmarkers
