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

// Package arrayofstruct checks that arrays containing structs have at least
// one required field in the struct definition.
//
// This prevents ambiguous YAML representations where the absence of required
// fields can lead to configurations that are unclear or have dramatically
// different meanings. For example, in NetworkPolicy, the difference between
// "match all packets to 10.0.0.1, port 80" vs "match all packets to 10.0.0.1
// on any port, and also match all packets to port 80 on any IP" can be subtle
// but critical.
package arrayofstruct
