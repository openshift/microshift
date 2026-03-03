// Copyright 2025 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package slices

// Translate is a generic function for translating an input slice
// of type S into a new slice of type E.
func Translate[S any, E any](translation func(S) E, in ...S) []E {
	e := []E{}

	for _, s := range in {
		e = append(e, translation(s))
	}

	return e
}
