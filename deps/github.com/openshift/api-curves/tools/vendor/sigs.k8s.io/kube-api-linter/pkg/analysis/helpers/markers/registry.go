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
package markers

import (
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

// Registry is a thread-safe set of known marker identifiers.
type Registry interface {
	// Register adds the provided identifiers to the Registry.
	Register(ids ...string)

	// Match performs a greedy match to determine if an input
	// marker string matches a known marker identifier. It returns
	// the matched identifier and a boolean representing if a match was
	// found. If no match is found, the returned identifier will be an
	// empty string.
	Match(in string) (string, bool)
}

var defaultRegistry = NewRegistry() //nolint:gochecknoglobals

// DefaultRegistry is a global registry for known markers.
// New linters should register the markers they care about during
// an init() function.
func DefaultRegistry() Registry {
	return defaultRegistry
}

// Registry is a thread-safe set of known marker identifiers.
type registry struct {
	identifiers sets.Set[string]
	mu          sync.Mutex
}

// NewRegistry creates a new Registry.
func NewRegistry() Registry {
	return &registry{
		identifiers: sets.New[string](),
		mu:          sync.Mutex{},
	}
}

// Register adds the provided identifiers to the Registry.
func (r *registry) Register(ids ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.identifiers.Insert(ids...)
}

// Match performs a greedy match to determine if an input
// marker string matches a known marker identifier. It returns
// the matched identifier and a boolean representing if a match was
// found. If no match is found, the returned identifier will be an
// empty string.
func (r *registry) Match(in string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If there is an exact match, return early.
	// This is likely when using markers with no expressions like
	// optional, required, kubebuilder:validation:Required, etc.
	if ok := r.identifiers.Has(in); ok {
		return in, true
	}

	// Look for the longest matching known identifier
	bestMatch := ""
	foundMatch := false

	for _, id := range r.identifiers.UnsortedList() {
		if strings.HasPrefix(in, id) {
			if len(bestMatch) < len(id) {
				bestMatch = id
				foundMatch = true
			}
		}
	}

	return bestMatch, foundMatch
}
