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
The `nodurations` linter checks that fields in the API types do not contain `Duration` type ether from the `time` package or the `k8s.io/apimachinery/pkg/apis/meta/v1` package.

It is recommended to avoid the use of Duration types. Their use ties the API to Go's notion of duration parsing, which may be hard to implement in other languages.

Instead, use an integer based field with a unit in the name, e.g. `FooSeconds`.
*/

package nodurations
