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

// Package main is meant to be compiled as a plugin for golangci-lint, see
// https://golangci-lint.run/plugins/go-plugins/.
package main

import (
	"fmt"

	"golang.org/x/tools/go/analysis"
	pluginbase "sigs.k8s.io/kube-api-linter/pkg/plugin/base"

	// Import the default linters.
	// DO NOT ADD DIRECTLY TO THIS FILE.
	_ "sigs.k8s.io/kube-api-linter/pkg/registration"
)

// New API, see https://github.com/golangci/golangci-lint/pull/3887.
func New(pluginSettings any) ([]*analysis.Analyzer, error) {
	plugin, err := pluginbase.New(pluginSettings)
	if err != nil {
		return nil, fmt.Errorf("error creating plugin: %w", err)
	}

	analyzers, err := plugin.BuildAnalyzers()
	if err != nil {
		return nil, fmt.Errorf("error building analyzers: %w", err)
	}

	return analyzers, nil
}
