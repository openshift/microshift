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
package nonullable

import (
	"errors"
	"fmt"

	"golang.org/x/tools/go/analysis"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/forbiddenmarkers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/initializer"
)

const (
	name = "nonullable"
	doc  = "Check that nullable marker is not present on any types or fields."
)

var errUnexpectedInitializerType = errors.New("expected forbiddenmarkers.Initializer() to be of type initializer.ConfigurableAnalyzerInitializer, but was not")

func newAnalyzer() *analysis.Analyzer {
	cfg := &forbiddenmarkers.Config{
		Markers: []forbiddenmarkers.Marker{
			{
				Identifier: "nullable",
			},
		},
	}

	configInit, ok := forbiddenmarkers.Initializer().(initializer.ConfigurableAnalyzerInitializer)
	if !ok {
		panic(fmt.Errorf("getting initializer: %w", errUnexpectedInitializerType))
	}

	errs := configInit.ValidateConfig(cfg, field.NewPath("nullable"))
	if err := errs.ToAggregate(); err != nil {
		panic(fmt.Errorf("nonullable linter has an invalid forbiddenmarkers configuration: %w", err))
	}

	analyzer, err := configInit.Init(cfg)
	if err != nil {
		panic(fmt.Errorf("nonullable linter encountered an error initializing wrapped forbiddenmarkers analyzer: %w", err))
	}

	analyzer.Name = name
	analyzer.Doc = doc

	return analyzer
}
