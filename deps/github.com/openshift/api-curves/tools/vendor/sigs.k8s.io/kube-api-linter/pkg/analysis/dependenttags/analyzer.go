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
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"

	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
)

// analyzer implements the dependenttags linter.
type analyzer struct {
	cfg Config
}

// newAnalyzer creates a new analyzer.
func newAnalyzer(cfg Config) *analysis.Analyzer {
	// Register markers from configuration
	for _, rule := range cfg.Rules {
		markers.DefaultRegistry().Register(rule.Identifier)

		for _, dep := range rule.DependsOn {
			markers.DefaultRegistry().Register(dep)
		}
	}

	a := &analyzer{
		cfg: cfg,
	}

	return &analysis.Analyzer{
		Name:     name,
		Doc:      "Enforces dependencies between markers.",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer, markers.Analyzer},
	}
}

// run is the main function for the analyzer.
func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string) {
		if field.Doc == nil {
			return
		}

		fieldMarkers := utils.TypeAwareMarkerCollectionForField(pass, markersAccess, field)

		for _, rule := range a.cfg.Rules {
			if _, ok := fieldMarkers[rule.Identifier]; ok {
				switch rule.Type {
				case DependencyTypeAny:
					handleAny(pass, field, rule, fieldMarkers, qualifiedFieldName)
				case DependencyTypeAll:
					handleAll(pass, field, rule, fieldMarkers, qualifiedFieldName)
				default:
					panic(fmt.Sprintf("unknown dependency type %s", rule.Type))
				}
			}
		}
	})

	return nil, nil //nolint:nilnil
}
func handleAll(pass *analysis.Pass, field *ast.Field, rule Rule, fieldMarkers markers.MarkerSet, qualifiedFieldName string) {
	missing := make([]string, 0, len(rule.DependsOn))

	for _, dependent := range rule.DependsOn {
		if _, depOk := fieldMarkers[dependent]; !depOk {
			missing = append(missing, fmt.Sprintf("+%s", dependent))
		}
	}

	if len(missing) > 0 {
		pass.Reportf(field.Pos(), "field %s with marker +%s is missing required marker(s): %s", qualifiedFieldName, rule.Identifier, strings.Join(missing, ", "))
	}
}

func handleAny(pass *analysis.Pass, field *ast.Field, rule Rule, fieldMarkers markers.MarkerSet, qualifiedFieldName string) {
	found := false

	for _, dependent := range rule.DependsOn {
		if _, depOk := fieldMarkers[dependent]; depOk {
			found = true
			break
		}
	}

	if !found {
		dependsOn := make([]string, len(rule.DependsOn))
		for i, d := range rule.DependsOn {
			dependsOn[i] = fmt.Sprintf("+%s", d)
		}

		pass.Reportf(field.Pos(), "field %s with marker +%s requires at least one of the following markers, but none were found: %s", qualifiedFieldName, rule.Identifier, strings.Join(dependsOn, ", "))
	}
}
