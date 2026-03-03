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
package forbiddenmarkers

import (
	"fmt"
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
)

const name = "forbiddenmarkers"

type analyzer struct {
	forbiddenMarkers []Marker
}

// NewAnalyzer creates a new analysis.Analyzer for the forbiddenmarkers
// linter based on the provided config.ForbiddenMarkersConfig.
func newAnalyzer(cfg *Config) *analysis.Analyzer {
	a := &analyzer{
		forbiddenMarkers: cfg.Markers,
	}

	analyzer := &analysis.Analyzer{
		Name:     name,
		Doc:      "Check that no forbidden markers are present on types and fields.",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer},
	}

	for _, marker := range a.forbiddenMarkers {
		markers.DefaultRegistry().Register(marker.Identifier)
	}

	return analyzer
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string) {
		checkField(pass, field, markersAccess, a.forbiddenMarkers, qualifiedFieldName)
	})

	inspect.InspectTypeSpec(func(typeSpec *ast.TypeSpec, markersAccess markers.Markers) {
		checkType(pass, typeSpec, markersAccess, a.forbiddenMarkers)
	})

	return nil, nil //nolint:nilnil
}

func checkField(pass *analysis.Pass, field *ast.Field, markersAccess markers.Markers, forbiddenMarkers []Marker, qualifiedFieldName string) {
	if field == nil || len(field.Names) == 0 {
		return
	}

	markers := utils.TypeAwareMarkerCollectionForField(pass, markersAccess, field)
	check(markers, forbiddenMarkers, reportField(pass, field, qualifiedFieldName))
}

func checkType(pass *analysis.Pass, typeSpec *ast.TypeSpec, markersAccess markers.Markers, forbiddenMarkers []Marker) {
	if typeSpec == nil {
		return
	}

	markers := markersAccess.TypeMarkers(typeSpec)
	check(markers, forbiddenMarkers, reportType(pass, typeSpec))
}

func check(markerSet markers.MarkerSet, forbiddenMarkers []Marker, reportFunc func(marker markers.Marker)) {
	for _, marker := range forbiddenMarkers {
		marks := markerSet.Get(marker.Identifier)
		for _, mark := range marks {
			if len(marker.RuleSets) == 0 {
				reportFunc(mark)
				continue
			}

			for _, ruleSet := range marker.RuleSets {
				if markerMatchesAttributeRules(mark, ruleSet.Attributes...) {
					reportFunc(mark)
				}
			}
		}
	}
}

func markerMatchesAttributeRules(marker markers.Marker, attrRules ...MarkerAttribute) bool {
	for _, attrRule := range attrRules {
		// if the marker doesn't contain the attribute for a specified rule it fails the AND
		// operation.
		val, ok := marker.Arguments[attrRule.Name]
		if !ok {
			return false
		}

		// if the value doesn't match one of the forbidden ones, this marker is not forbidden
		if len(attrRule.Values) > 0 && !slices.Contains(attrRule.Values, val) {
			return false
		}
	}

	return true
}

func reportField(pass *analysis.Pass, field *ast.Field, qualifiedFieldName string) func(marker markers.Marker) {
	return func(marker markers.Marker) {
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("field %s has forbidden marker %q", qualifiedFieldName, marker.String()),
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: fmt.Sprintf("remove forbidden marker %q", marker.String()),
					TextEdits: []analysis.TextEdit{
						{
							Pos: marker.Pos,
							End: marker.End,
						},
					},
				},
			},
		})
	}
}

func reportType(pass *analysis.Pass, typeSpec *ast.TypeSpec) func(marker markers.Marker) {
	return func(marker markers.Marker) {
		pass.Report(analysis.Diagnostic{
			Pos:     typeSpec.Pos(),
			Message: fmt.Sprintf("type %s has forbidden marker %q", typeSpec.Name, marker.String()),
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: fmt.Sprintf("remove forbidden marker %q", marker.String()),
					TextEdits: []analysis.TextEdit{
						{
							Pos: marker.Pos,
							End: marker.End,
						},
					},
				},
			},
		})
	}
}
