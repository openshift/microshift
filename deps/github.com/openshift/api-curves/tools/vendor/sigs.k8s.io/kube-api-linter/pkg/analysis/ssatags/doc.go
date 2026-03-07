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

// Package ssatags provides an analyzer to enforce proper Server-Side Apply (SSA) tags on array fields.
//
// This analyzer ensures that array fields in Kubernetes API objects have the appropriate
// listType markers (atomic, set, or map) for proper Server-Side Apply behavior.
//
// Server-Side Apply (SSA) is a Kubernetes feature that allows multiple controllers to manage
// different parts of an object. The listType markers help SSA understand how to merge arrays:
//
// - listType=atomic: The entire list is replaced when updated
// - listType=set: List elements are treated as a set (no duplicates, order doesn't matter)
// - listType=map: Elements are identified by specific key fields for granular updates
//
// Important Note on listType=set:
// The use of listType=set is discouraged for object arrays due to Server-Side Apply
// compatibility issues. When multiple controllers attempt to apply changes to an object
// array with listType=set, the merge behavior can be unpredictable and may lead to
// data loss or unexpected conflicts. For object arrays, use listType=atomic for simple
// replacement semantics or listType=map for granular field-level merging.
// listType=set is safe to use with primitive arrays (strings, integers, etc.).
//
// The analyzer checks for:
//
// 1. Missing listType markers on array fields
// 2. Invalid listType values (must be atomic, set, or map)
// 3. Usage of listType=set on object arrays (discouraged due to compatibility issues)
// 4. Missing listMapKey markers for listType=map arrays
// 5. Incorrect usage of listType=map on primitive arrays
//
// Configuration options:
//
//   - listTypeSetUsage: Control warnings for listType=set usage on object arrays
//     Valid values: "Warn" (default) or "Ignore"
package ssatags
