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
package errors

import "errors"

var (
	// ErrCouldNotCreateMarkers is returned when the markers could not be created.
	ErrCouldNotCreateMarkers = errors.New("could not create markers")

	// ErrCouldNotCreateStructFieldTags is returned when the struct field tags could not be created.
	ErrCouldNotCreateStructFieldTags = errors.New("could not create new structFieldTags")

	// ErrCouldNotGetInspector is returned when the inspector could not be retrieved.
	ErrCouldNotGetInspector = errors.New("could not get inspector")

	// ErrCouldNotGetMarkers is returned when the markers analyzer could not be retrieved.
	ErrCouldNotGetMarkers = errors.New("could not get markers")

	// ErrCouldNotGetJSONTags is returned when the JSON tags could not be retrieved.
	ErrCouldNotGetJSONTags = errors.New("could not get json tags")
)
