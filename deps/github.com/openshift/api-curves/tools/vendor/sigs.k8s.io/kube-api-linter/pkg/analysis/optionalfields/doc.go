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
optionalfields is a linter to check that fields that are marked as optional are marshalled properly depending on the configured policies.

By default, all optional fields should be pointers and have omitempty tags. The exception to this would be arrays and maps where the empty value can be omitted without the need for a pointer.

However, where the zero value for a field is not a valid value (e.g. the empty string, or 0), the field does not need to be a pointer as the zero value could never be admitted.
In this case, the field may not need to be a pointer, and, with the WhenRequired preference, the linter will point out where the fields do not need to be pointers.

Structs are also inspected to determine if they require a pointer.
If a struct has any required fields, or a minimum number of properties, then fields leveraging the struct should be pointers.

Optional structs do not always need to be pointers, but may be marshalled as `{}` because the JSON marshaller in Go cannot determine whether a struct is empty or not.
*/
package optionalfields
