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
package base

import (
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/registry"
	"sigs.k8s.io/kube-api-linter/pkg/config"
	"sigs.k8s.io/kube-api-linter/pkg/validation"
)

func init() {
	register.Plugin("kubeapilinter", New)
}

// New creates a new golangci-lint plugin based on the KAL analyzers.
func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[config.GolangCIConfig](settings)
	if err != nil {
		return nil, fmt.Errorf("error decoding settings: %w", err)
	}

	return &GolangCIPlugin{config: s}, nil
}

// GolangCIPlugin constructs a new plugin for the golangci-lint
// plugin pattern.
// This allows golangci-lint to build a version of itself, containing
// all of the analyzers included in KAL.
type GolangCIPlugin struct {
	config config.GolangCIConfig
}

// BuildAnalyzers returns all of the analyzers to run, based on the configuration.
func (f *GolangCIPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	if err := validation.ValidateGolangCIConfig(f.config, field.NewPath("")); err != nil {
		return nil, fmt.Errorf("error in KAL configuration: %w", err)
	}

	analyzers, err := registry.DefaultRegistry().InitializeLinters(f.config.Linters, f.config.LintersConfig)
	if err != nil {
		return nil, fmt.Errorf("error initializing analyzers: %w", err)
	}

	return analyzers, nil
}

// GetLoadMode implements the golangci-lint plugin interface.
func (f *GolangCIPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
