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
requiredfields is a linter to check that fields that are marked as required are marshalled properly.
The linter will check for fields that are marked as required using the +required marker, or the +kubebuilder:validation:Required marker.

Required fields should have omitempty or omitzero tags to prevent "mess" in the encoded object.
omitzero is handled only for fields with struct type.

Fields are not typically pointers.
A field doesn't need to be a pointer if its zero value is not a valid value, as this zero value could never be accepted.
However, if the zero value is valid, the field should be a pointer to differentiate between an unset state and a valid zero value.
*/
package requiredfields
