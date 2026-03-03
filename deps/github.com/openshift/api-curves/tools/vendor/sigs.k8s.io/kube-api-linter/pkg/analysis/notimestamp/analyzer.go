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
package notimestamp

import (
	"errors"
	"fmt"

	"golang.org/x/tools/go/analysis"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/initializer"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/namingconventions"
)

const (
	name = "notimestamp"
	doc  = "Suggest the usage of the term 'time' over 'timestamp'"
)

var errUnexpectedInitializerType = errors.New("expected namingconventions.Initializer() to be of type initializer.ConfigurableAnalyzerInitializer, but was not")

func newAnalyzer() *analysis.Analyzer {
	cfg := &namingconventions.Config{
		Conventions: []namingconventions.Convention{
			{
				Name:             "notimestamp",
				ViolationMatcher: "(?i)timestamp",
				Operation:        namingconventions.OperationReplacement,
				Message:          "prefer use of the term 'time' over 'timestamp'",
				Replacement:      "Time",
			},
		},
	}

	configInit, ok := namingconventions.Initializer().(initializer.ConfigurableAnalyzerInitializer)
	if !ok {
		panic(fmt.Errorf("getting initializer: %w", errUnexpectedInitializerType))
	}

	errs := configInit.ValidateConfig(cfg, field.NewPath("notimestamp"))
	if err := errs.ToAggregate(); err != nil {
		panic(fmt.Errorf("notimestamp linter has an invalid namingconventions configuration: %w", err))
	}

	analyzer, err := configInit.Init(cfg)
	if err != nil {
		panic(fmt.Errorf("notimestamp linter encountered an error initializing wrapped namingconventions analyzer: %w", err))
	}

	analyzer.Name = name
	analyzer.Doc = doc

	return analyzer
}
