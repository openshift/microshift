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

package runner

import (
	"sigs.k8s.io/crdify/pkg/validations"
	"sigs.k8s.io/crdify/pkg/validations/crd/existingfieldremoval"
	"sigs.k8s.io/crdify/pkg/validations/crd/scope"
	"sigs.k8s.io/crdify/pkg/validations/crd/storedversionremoval"
	"sigs.k8s.io/crdify/pkg/validations/property"
)

//nolint:gochecknoglobals
var defaultRegistry = validations.NewRegistry()

func init() {
	existingfieldremoval.Register(defaultRegistry)
	scope.Register(defaultRegistry)
	storedversionremoval.Register(defaultRegistry)
	property.RegisterDefault(defaultRegistry)
	property.RegisterEnum(defaultRegistry)
	property.RegisterMaximum(defaultRegistry)
	property.RegisterMaxItems(defaultRegistry)
	property.RegisterMaxLength(defaultRegistry)
	property.RegisterMaxProperties(defaultRegistry)
	property.RegisterMinimum(defaultRegistry)
	property.RegisterMinItems(defaultRegistry)
	property.RegisterMinLength(defaultRegistry)
	property.RegisterMinProperties(defaultRegistry)
	property.RegisterRequired(defaultRegistry)
	property.RegisterType(defaultRegistry)
	property.RegisterDescription(defaultRegistry)
}

// DefaultRegistry returns a pre-configured validations.Registry.
func DefaultRegistry() validations.Registry {
	return defaultRegistry
}
