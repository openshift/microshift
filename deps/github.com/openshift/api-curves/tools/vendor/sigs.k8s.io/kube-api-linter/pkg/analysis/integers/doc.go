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
integers is an analyzer that checks for usage of unsupported integer types.

According to the API conventions, only int32 and int64 types should be used in Kubernetes APIs.

int32 is preferred and should be used in most cases, unless the use case requireds representing
values larger than int32.

It also states that unsigned integers should be replaced with signed integers, and then numeric
lower bounds added to prevent negative integers.

Succinctly this analyzer checks for int, int8, int16, uint, uint8, uint16, uint32 and uint64 types
and highlights that they should not be used.
*/
package integers
