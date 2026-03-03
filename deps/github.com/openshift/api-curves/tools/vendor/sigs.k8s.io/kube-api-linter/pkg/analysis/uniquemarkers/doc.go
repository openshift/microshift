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
* The `uniquemarkers` linter ensures that types and fields do not contain more than a single
* definition of a marker that should only be present once.
*
* Because this linter has no way of determining which marker definition was intended it does not suggest any fixes.
*
* It can configured to include a set of custom markers in the analysis by setting:
* ```yaml
* lintersConfig:
*   uniqueMarkers:
*     customMarkers:
*       - identifier: custom:SomeCustomMarker
*         attributes:
*           - fruit
* ```
* For each custom marker, it must specify an `identifier` and optionally some `attributes`.
* As an example, take the marker definition `kubebuilder:validation:XValidation:rule='has(self.foo)',message='should have foo',fieldPath='.foo'`.
* The identifier for the marker is `kubebuilder:validation:XValidation` and its attributes are `rule`, `message`, and `fieldPath`.
*
* When specifying `attributes`, those attributes are included in the uniqueness identification of a marker definition.
*
* Taking the example configuration from above:
*
* - Marker definitions of `custom:SomeCustomMarker:fruit=apple,color=red` and `custom:SomeCustomMarker:fruit=apple,color=green` would violate the uniqueness requirement and be flagged.
* - Marker definitions of `custom:SomeCustomMarker:fruit=apple,color=red` and `custom:SomeCustomMarker:fruit=orange,color=red` would _not_ violate the uniqueness requirement.
*
* Each entry in `customMarkers` must have a unique `identifier`.
 */
package uniquemarkers
