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
optionalorrequired is a linter to ensure that all fields are marked as either optional or required.
It also checks for the presence of optional or required markers on type declarations, and forbids this pattern.

By default, it searches for the `+optional` and `+required` markers, and ensures that all fields are marked
with at least one of these markers.

The linter can be configured to use different markers, by setting the `PreferredOptionalMarker` and `PreferredRequiredMarker`.
The default values are `+optional` and `+required`, respectively.
The available alternate values for each marker are:

For PreferredOptionalMarker:
  - `+optional`: The standard Kubernetes marker for optional fields.
  - `+kubebuilder:validation:Optional`: The Kubebuilder marker for optional fields.

For PreferredRequiredMarker:
  - `+required`: The standard Kubernetes marker for required fields.
  - `+kubebuilder:validation:Required`: The Kubebuilder marker for required fields.

When a field is marked with both the Kubernetes and Kubebuilder markers, the linter will suggest to remove the Kubebuilder marker.
When a field is marked only with the Kubebuilder marker, the linter will suggest to use the Kubernetes marker instead.
This behaviour is reversed when the `PreferredOptionalMarker` and `PreferredRequiredMarker` are set to the Kubebuilder markers.

Use the linter fix option to automatically apply suggested fixes.
*/
package optionalorrequired
