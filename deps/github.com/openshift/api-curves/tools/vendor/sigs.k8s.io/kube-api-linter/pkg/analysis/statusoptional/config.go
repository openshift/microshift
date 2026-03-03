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
package statusoptional

// StatusOptionalConfig contains configuration for the statusoptional linter.
type StatusOptionalConfig struct {
	// preferredOptionalMarker is the preferred marker to use for optional fields.
	// If this field is not set, the default value is "optional".
	// Valid values are "optional", "kubebuilder:validation:Optional" and "k8s:optional".
	PreferredOptionalMarker string `json:"preferredOptionalMarker"`
}
