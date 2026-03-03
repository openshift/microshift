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
The statussubresource linter ensures that a type marked with
kubebuilder:object:root:=true and containing a status field is marked with
kubebuilder:subresource:status

Status fields should always be considered a subresource for the API, and this
way the linter guarantees that the right marker is added on this field.

The linter will report an issue if the root object has a status field and does
not contain the marker 'kubebuilder:subresource:status'
*/
package statussubresource
