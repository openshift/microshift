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
package nomaps

// NoMapsPolicy is the policy for the nomaps linter.
type NoMapsPolicy string

const (
	// NoMapsEnforce indicates that all declarations for maps are rejected.
	NoMapsEnforce NoMapsPolicy = "Enforce"

	// NoMapsAllowStringToStringMaps indicates that only string to string maps are allowed.
	NoMapsAllowStringToStringMaps NoMapsPolicy = "AllowStringToStringMaps"

	// NoMapsIgnore indicates that all declarations which the value type is a primitive type are allowed.
	NoMapsIgnore NoMapsPolicy = "Ignore"
)

// NoMapsConfig contains configuration for the nomaps linter.
type NoMapsConfig struct {
	// policy is the policy for the nomaps linter.
	// Valid values are "Enforce", "AllowStringToStringMaps" and "Ignore".
	// When set to "Enforce", all declarations for maps are rejected.
	// When set to "AllowStringToStringMaps", only string to string maps are allowed.
	// When set to "Ignore", maps of primitive types are allowed, but maps containing complex types are not allowed.
	// When otherwise not specified, the default value is "AllowStringToStringMaps".
	Policy NoMapsPolicy `json:"policy"`
}
