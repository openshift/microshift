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
package validation

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kube-api-linter/pkg/analysis/registry"
	"sigs.k8s.io/kube-api-linter/pkg/config"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateLinters is used to validate the configuration in the config.Linters struct.
//
//nolint:cyclop
func ValidateLinters(l config.Linters, fldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}

	enable := sets.New(l.Enable...)
	enablePath := fldPath.Child("enable")

	switch {
	case len(enable) != len(l.Enable):
		fieldErrors = append(fieldErrors, field.Invalid(enablePath, l.Enable, "values in 'enable' must be unique"))
	case enable.Has(config.Wildcard) && enable.Len() != 1:
		fieldErrors = append(fieldErrors, field.Invalid(enablePath, l.Enable, "wildcard ('*') must not be specified with other values"))
	case !enable.Has(config.Wildcard) && enable.Difference(registry.DefaultRegistry().AllLinters()).Len() > 0:
		fieldErrors = append(fieldErrors, field.Invalid(enablePath, l.Enable, fmt.Sprintf("unknown linters: %s", strings.Join(enable.Difference(registry.DefaultRegistry().AllLinters()).UnsortedList(), ","))))
	}

	disable := sets.New(l.Disable...)
	disablePath := fldPath.Child("disable")

	switch {
	case len(disable) != len(l.Disable):
		fieldErrors = append(fieldErrors, field.Invalid(disablePath, l.Disable, "values in 'disable' must be unique"))
	case disable.Has(config.Wildcard) && disable.Len() != 1:
		fieldErrors = append(fieldErrors, field.Invalid(disablePath, l.Disable, "wildcard ('*') must not be specified with other values"))
	case !disable.Has(config.Wildcard) && disable.Difference(registry.DefaultRegistry().AllLinters()).Len() > 0:
		fieldErrors = append(fieldErrors, field.Invalid(disablePath, l.Disable, fmt.Sprintf("unknown linters: %s", strings.Join(disable.Difference(registry.DefaultRegistry().AllLinters()).UnsortedList(), ","))))
	}

	if enable.Intersection(disable).Len() > 0 {
		fieldErrors = append(fieldErrors, field.Invalid(fldPath, l, fmt.Sprintf("values in 'enable' and 'disable may not overlap, overlapping values: %s", strings.Join(enable.Intersection(disable).UnsortedList(), ","))))
	}

	return fieldErrors
}
