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
nonpointerstructs is a linter that checks that non-pointer structs that contain required fields are marked as required.
Non-pointer structs that contain no required fields are marked as optional.

This linter is important for types validated in Go as there is no way to validate the optionality of the fields at runtime,
aside from checking the fields within them.

This linter is NOT intended to be used to check for CRD types.
The advice of this linter may be applied to CRD types, but it is not necessary for CRD types due to optionality being validated by openapi and no native Go code.
For CRD types, the optionalfields and requiredfields linters should be used instead.

If a struct is marked required, this can only be validated by having a required field within it.
If there are no required fields, the struct is implicitly optional and must be marked as so.

To have an optional struct field that includes required fields, the struct must be a pointer.
To have a required struct field that includes no required fields, the struct must be a pointer.
*/
package nonpointerstructs
