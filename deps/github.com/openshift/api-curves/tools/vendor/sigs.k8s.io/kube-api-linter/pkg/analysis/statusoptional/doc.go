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
The statusoptional linter ensures that all first-level children fields within a status struct
are marked as optional.

This is important because status fields should be optional to allow for partial updates
and backward compatibility.

This linter checks:
1. For structs with a JSON tag of "status"
2. All direct child fields of the status struct
3. Ensures each child field has an optional marker

The linter will report an issue if any field in the status struct is not marked as optional
and will suggest a fix to add the appropriate optional marker.
*/
package statusoptional
