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

package dependenttags

import (
	"fmt"

	"golang.org/x/tools/go/analysis"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/initializer"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/registry"
)

const (
	name = "dependenttags"
)

func init() {
	registry.DefaultRegistry().RegisterLinter(Initializer())
}

// Initializer returns the AnalyzerInitializer for this Analyzer so that it can be added to the registry.
func Initializer() initializer.ConfigurableAnalyzerInitializer {
	return initializer.NewConfigurableInitializer(
		name,
		initAnalyzer,
		false,
		validateConfig,
	)
}

// initAnalyzer returns the initialized Analyzer.
func initAnalyzer(cfg *Config) (*analysis.Analyzer, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	return newAnalyzer(*cfg), nil
}

// validateConfig validates the linter configuration.
func validateConfig(cfg *Config, fldPath *field.Path) field.ErrorList {
	var errs field.ErrorList
	if cfg == nil {
		return errs
	}

	rulesPath := fldPath.Child("rules")

	if len(cfg.Rules) == 0 {
		errs = append(errs, field.Invalid(rulesPath, cfg.Rules, "rules cannot be empty"))
	}

	for i, rule := range cfg.Rules {
		if rule.Identifier == "" {
			errs = append(errs, field.Invalid(rulesPath.Index(i).Child("identifier"), rule.Identifier, "identifier marker cannot be empty"))
		}

		if len(rule.DependsOn) == 0 {
			errs = append(errs, field.Invalid(rulesPath.Index(i).Child("dependsOn"), rule.DependsOn, "dependsOn list cannot be empty"))
		}

		if rule.Type == "" {
			errs = append(errs, field.Required(rulesPath.Index(i).Child("type"), fmt.Sprintf("type must be explicitly set to '%s' or '%s'", DependencyTypeAll, DependencyTypeAny)))
		} else {
			switch rule.Type {
			case DependencyTypeAll, DependencyTypeAny:
				// valid
			default:
				errs = append(errs, field.Invalid(rulesPath.Index(i).Child("type"), rule.Type, fmt.Sprintf("type must be '%s' or '%s'", DependencyTypeAll, DependencyTypeAny)))
			}
		}
	}

	return errs
}
