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
nobools is an analyzer that checks for usage of bool types.

Boolean values can only ever have 2 states, true or false.
Over time, needs may change, and with a bool type, there is no way to add additional states.
This problem then leads to pairs of bools, where values of one are only valid given the value of another.
This is confusing and error-prone.

It is recommended instead to use a string type with a set of constants to represent the different states,
creating an enum.

By using an enum, not only can you provide meaningful names for the various states of the API,
but you can also add additional states in the future without breaking the API.
*/
package nobools
