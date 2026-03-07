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
package registry

import (
	"fmt"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/initializer"
	"sigs.k8s.io/kube-api-linter/pkg/config"
	"sigs.k8s.io/yaml"

	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// defaultRegistry is the default registry instance.
//
//nolint:gochecknoglobals
var defaultRegistry = NewRegistry()

// DefaultRegistry is the default registry instance.
// It is global and allows blank import style registration of linters.
func DefaultRegistry() Registry {
	return defaultRegistry
}

// Registry is used to fetch and initialize analyzers.
type Registry interface {
	// RegisterLinter adds the given linter to the registry.
	RegisterLinter(initializer.AnalyzerInitializer)

	// DefaultLinters returns the names of linters that are enabled by default.
	DefaultLinters() sets.Set[string]

	// AllLinters returns the names of all registered linters.
	AllLinters() sets.Set[string]

	// InitializeLinters returns a set of newly initialized linters based on the
	// provided configuration.
	InitializeLinters(config.Linters, config.LintersConfig) ([]*analysis.Analyzer, error)
}

type registry struct {
	lock         sync.RWMutex
	initializers []initializer.AnalyzerInitializer
}

// NewRegistry returns a new registry, from which analyzers can be fetched.
func NewRegistry() Registry {
	return &registry{
		initializers: []initializer.AnalyzerInitializer{},
	}
}

// RegisterLinter registers the linter with the registry.
func (r *registry) RegisterLinter(initializer initializer.AnalyzerInitializer) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.initializers = append(r.initializers, initializer)
}

// DefaultLinters returns the list of linters that are registered
// as being enabled by default.
func (r *registry) DefaultLinters() sets.Set[string] {
	r.lock.RLock()
	defer r.lock.RUnlock()

	defaultLinters := sets.New[string]()

	for _, initializer := range r.initializers {
		if initializer.Default() {
			defaultLinters.Insert(initializer.Name())
		}
	}

	return defaultLinters
}

// AllLinters returns the list of all known linters that are known
// to the registry.
func (r *registry) AllLinters() sets.Set[string] {
	r.lock.RLock()
	defer r.lock.RUnlock()

	linters := sets.New[string]()

	for _, initializer := range r.initializers {
		linters.Insert(initializer.Name())
	}

	return linters
}

// InitializeLinters returns a list of initialized linters based on the provided config.
func (r *registry) InitializeLinters(cfg config.Linters, lintersCfg config.LintersConfig) ([]*analysis.Analyzer, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if errs := r.validateLintersConfig(cfg, lintersCfg, field.NewPath("lintersConfig")); len(errs) > 0 {
		return nil, fmt.Errorf("error validating linters config: %w", errs.ToAggregate())
	}

	analyzers := []*analysis.Analyzer{}
	errs := []error{}

	for _, init := range r.getEnabledInitializers(cfg) {
		var linterConfig any

		if ci, ok := isConfigurable(init); ok {
			var err error

			linterConfig, err = getLinterTypedConfig(ci, lintersCfg)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to get linter config: %w", err))
			}
		}

		a, err := init.Init(linterConfig)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to initialize linter %s: %w", init.Name(), err))
			continue
		}

		analyzers = append(analyzers, a)
	}

	return analyzers, kerrors.NewAggregate(errs)
}

// validateLintersConfig validates the provided linters config
// against the set or registered linters.
func (r *registry) validateLintersConfig(cfg config.Linters, lintersCfg config.LintersConfig, fieldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}
	validatedLinters := sets.New[string]()

	for _, init := range r.getEnabledInitializers(cfg) {
		if ci, ok := isConfigurable(init); ok {
			linterConfig, err := getLinterTypedConfig(ci, lintersCfg)
			if err != nil {
				fieldErrors = append(fieldErrors, field.Invalid(fieldPath.Child(init.Name()), linterConfig, err.Error()))
				continue
			}

			fieldErrors = append(fieldErrors, ci.ValidateConfig(linterConfig, fieldPath.Child(init.Name()))...)

			validatedLinters.Insert(init.Name())
		}
	}

	fieldErrors = append(fieldErrors, validateUnusedLinters(lintersCfg, validatedLinters, fieldPath)...)

	return fieldErrors
}

// getEnabledInitializers returns the initializers that are enabled by the config.
// It returns a list of initializers that are enabled by the config.
func (r *registry) getEnabledInitializers(cfg config.Linters) []initializer.AnalyzerInitializer {
	enabled := sets.New(cfg.Enable...)
	disabled := sets.New(cfg.Disable...)

	allEnabled := enabled.Len() == 1 && enabled.Has(config.Wildcard)
	allDisabled := disabled.Len() == 1 && disabled.Has(config.Wildcard)

	initializers := []initializer.AnalyzerInitializer{}

	for _, init := range r.initializers {
		if !disabled.Has(init.Name()) && (allEnabled || enabled.Has(init.Name()) || !allDisabled && init.Default()) {
			initializers = append(initializers, init)
		}
	}

	return initializers
}

// getLinterTypedConfig returns the typed config for a linter.
func getLinterTypedConfig(ci initializer.ConfigurableAnalyzerInitializer, lintersCfg config.LintersConfig) (any, error) {
	rawConfig, ok := getConfigByName(ci.Name(), lintersCfg)
	if !ok {
		return ci.ConfigType(), nil
	}

	encodedConfig, err := yaml.Marshal(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("error encoding config for linter %q: %w", ci.Name(), err)
	}

	linterConfig := ci.ConfigType()

	if err := yaml.Unmarshal(encodedConfig, linterConfig); err != nil {
		return nil, fmt.Errorf("error reading config for linter %q: %w", ci.Name(), err)
	}

	return linterConfig, nil
}

// getConfigByName returns the config for a linter by name.
// It returns the config as a byte slice and a boolean indicating if the config was found.
// It also supports backwards compatibility with early configuration.
// We use to have camelCased config names, but now it is all lowercase matched on the linter name.
// TODO(@JoelSpeed): Remove the strings.ToLower in a future release with a release note about the change.
func getConfigByName(name string, lintersCfg config.LintersConfig) (any, bool) {
	rawConfig, ok := lintersCfg[name]
	if ok {
		return rawConfig, true
	}

	// Hack to allow backwards compatibility with early configuration.
	// We use to have camelCased config names, but now it is all lowercase matched on the linter name.
	// TODO(@JoelSpeed): Remove this in a future release with a release note about the change.
	for k, v := range lintersCfg {
		if strings.ToLower(k) == name {
			return v, true
		}
	}

	return nil, false
}

// validateUnusedLinters validates that all linters in the config are enabled.
// It returns a list of errors for each linter that is not enabled.
func validateUnusedLinters(lintersCfg config.LintersConfig, validatedLinters sets.Set[string], fieldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}

	for name := range lintersCfg {
		// Hack to allow backwards compatibility with early configuration.
		// We use to have camelCased config names, but now it is all lowercase matched on the linter name.
		// TODO(@JoelSpeed): Remove the strings.ToLower in a future release with a release note about the change.
		if !validatedLinters.Has(strings.ToLower(name)) {
			fieldErrors = append(fieldErrors, field.Invalid(fieldPath.Child(name), nil, "linter is not enabled"))
		}
	}

	return fieldErrors
}

// isConfigurable determines whether or not the initializer expects to be provided a config.
// When true, the initializer should also match the ConfigurableAnalyzerInitializer interface.
func isConfigurable(init initializer.AnalyzerInitializer) (initializer.ConfigurableAnalyzerInitializer, bool) {
	ci, ok := init.(initializer.ConfigurableAnalyzerInitializer)
	return ci, ok
}
