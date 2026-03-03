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

package validations

import (
	"errors"
	"fmt"
	"sync"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/crdify/pkg/config"
)

// Comparable is a generic interface that represents either a
// CustomResourceDefinition or a JSONSchemaProps.
type Comparable interface {
	apiextensionsv1.CustomResourceDefinition | apiextensionsv1.JSONSchemaProps
}

// Comparator is a generic interface for comparing two objects and getting back
// a ComparisonResult.
type Comparator[T Comparable] interface {
	// Compare compares two instances of type T and returns a ComparisonResult
	Compare(a, b *T) ComparisonResult
}

// Validation is an interface to represent the minimal set of
// functionality that needs to be implemented by a validation.
type Validation interface {
	// Name is the name of the validation
	Name() string
	// SetEnforcement sets the enforcement policy for the validation
	SetEnforcement(policy config.EnforcementPolicy)
}

// ComparisonResult contains the results of running a Comparator.
type ComparisonResult struct {
	// Name is the name of the Comparator implementation that
	// performed the comparison
	Name string `json:"name"`

	// Errors is the set of errors encountered during comparison
	Errors []string `json:"errors,omitempty"`

	// Warnings is the set of warnings encountered during comparison
	Warnings []string `json:"warnings,omitempty"`
}

// Factory is a function used for creating a Validation based on a
// provided configuration. Should return an error if the Validation
// cannot be successfully created with the provided configuration.
type Factory func(config map[string]interface{}) (Validation, error)

// Registry is a registry for Validations.
type Registry interface {
	// Register registers a name and how to create an instance of a Validation with that name
	Register(name string, validation Factory)

	// Registered returns a list of the registered Validation names
	Registered() []string

	// Validation returns a Validation for the provided name and configuration
	Validation(name string, config map[string]interface{}) (Validation, error)
}

// validationRegistry is an implementation of Registry.
type validationRegistry struct {
	lock        sync.Mutex
	validations map[string]Factory
}

// NewRegistry creates a new Registry.
func NewRegistry() Registry {
	return &validationRegistry{
		lock:        sync.Mutex{},
		validations: make(map[string]Factory),
	}
}

// Register registers a name and how to create an instance of a Validation with that name.
// If a validation has already been registered, this method will panic.
func (vr *validationRegistry) Register(name string, validation Factory) {
	vr.lock.Lock()
	defer vr.lock.Unlock()

	if _, ok := vr.validations[name]; ok {
		panic(fmt.Sprintf("validation %q has already been registered", name))
	}

	vr.validations[name] = validation
}

// Registered returns the set of registered validation names.
func (vr *validationRegistry) Registered() []string {
	vr.lock.Lock()
	defer vr.lock.Unlock()

	keys := []string{}

	for k := range vr.validations {
		keys = append(keys, k)
	}

	return keys
}

// Validation creates the Validation for the provided validation name and configuration if it is registered.
func (vr *validationRegistry) Validation(name string, config map[string]interface{}) (Validation, error) {
	vr.lock.Lock()
	defer vr.lock.Unlock()

	factory, ok := vr.validations[name]
	if !ok {
		return nil, fmt.Errorf("%w : %q", errUnknownValidation, name)
	}

	return factory(config)
}

var errUnknownValidation = errors.New("unknown validation")
