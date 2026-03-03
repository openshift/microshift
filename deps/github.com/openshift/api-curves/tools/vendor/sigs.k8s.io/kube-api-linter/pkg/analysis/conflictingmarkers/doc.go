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
conflictingmarkers is a linter that detects and reports when mutually exclusive markers are used on the same field.
This prevents common configuration errors and unexpected behavior in Kubernetes API types.

The linter reports issues when markers from two or more sets of a conflict definition are present on the same field.
It does NOT report issues when multiple markers from the same set are present - only when markers from
different sets within the same conflict definition are found together.

The linter is fully configurable and requires users to define all conflict sets they want to check.
There are no built-in conflict sets - all conflicts must be explicitly configured.

Each conflict set must specify:
- A unique name for the conflict
- Multiple sets of markers that are mutually exclusive with each other (at least 2 sets)
- A description explaining why the markers conflict

Example configuration:
```yaml

		lintersConfig:
	      conflictingmarkers:
	        conflicts:
	          - name: "default_vs_required"
	            sets:
	              - ["default", "kubebuilder:default"]
	              - ["required", "kubebuilder:validation:Required", "k8s:required"]
	            description: "A field with a default value cannot be required"
	          - name: "three_way_conflict"
	            sets:
	              - ["custom:marker1", "custom:marker2"]
	              - ["custom:marker3", "custom:marker4"]
	              - ["custom:marker5", "custom:marker6"]
	            description: "Three-way conflict between marker sets"

```

Configuration options:
- `conflicts`: Required list of conflict set definitions.

Note: This linter is not enabled by default and must be explicitly enabled in the configuration.

The linter does not provide automatic fixes as it cannot determine which conflicting marker should be removed.
*/
package conflictingmarkers
