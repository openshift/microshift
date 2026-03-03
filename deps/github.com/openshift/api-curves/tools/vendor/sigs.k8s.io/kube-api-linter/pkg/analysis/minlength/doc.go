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
minlength is an analyzer that checks that all string fields have a minimum length, and that all array fields have a minimum number of items,
that maps have a minimum number of properties, and that structs that do not have required fields have a minimum number of fields.

String fields that are not otherwise bound in length, through being an enum or formatted in a certain way, should have a minimum length.
This ensures that authors make a choice about whether or not the empty string is a valid choice for users.

Array fields should have a minimum number of items.
This ensures that empty arrays are not allowed.
Empty arrays are generally not recommended and API authors should generally not distinguish between empty and omitted arrays.
When the empty array is a valid choice, setting the minimum items marker to 0 can be used to indicate that this is an explicit choice.

Maps should have a minimum number of properties.
This ensures that empty maps are not allowed.
Empty maps are generally not recommended and API authors should generally not distinguish between empty and omitted maps.
When the empty map is a valid choice, setting the minimum properties marker to 0 can be used to indicate that this is an explicit choice.

Structs that do not have required fields and do not define an equivalent constraint, i.e., `kubebuilder:validation:ExactlyOneOf` or `kubebuilder:validation:AtLeastOneOf`,
should have a minimum number of fields.
This ensures that empty structs are not allowed.
Empty structs are generally not recommended and API authors should generally not distinguish between empty and omitted structs.
When the empty struct is a valid choice, setting the minimum properties marker to 0 can be used to indicate that this is an explicit choice.

For strings, the minimum length can be set using the `kubebuilder:validation:MinLength` tag.
For arrays, the minimum number of items can be set using the `kubebuilder:validation:MinItems` tag.
For maps, the minimum number of properties can be set using the `kubebuilder:validation:MinProperties` tag.
For structs, the minimum number of fields can be set using the `kubebuilder:validation:MinProperties` tag.

For arrays of strings, the minimum length of each string can be set using the `kubebuilder:validation:items:MinLength` tag,
on the array field itself.
Or, if the array uses a string type alias, the `kubebuilder:validation:MinLength` tag can be used on the alias.
*/
package minlength
