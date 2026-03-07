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
package conflictingmarkers

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"k8s.io/apimachinery/pkg/util/sets"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
)

const name = "conflictingmarkers"

type analyzer struct {
	conflictSets []ConflictSet
}

func newAnalyzer(cfg *ConflictingMarkersConfig) *analysis.Analyzer {
	if cfg == nil {
		cfg = &ConflictingMarkersConfig{}
	}

	// Register markers from configuration
	for _, conflictSet := range cfg.Conflicts {
		for _, set := range conflictSet.Sets {
			for _, markerID := range set {
				markers.DefaultRegistry().Register(markerID)
			}
		}
	}

	a := &analyzer{
		conflictSets: cfg.Conflicts,
	}

	return &analysis.Analyzer{
		Name:     name,
		Doc:      "Check that fields do not have conflicting markers from mutually exclusive sets",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer},
	}
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string) {
		checkField(pass, field, markersAccess, a.conflictSets, qualifiedFieldName)
	})

	return nil, nil //nolint:nilnil
}

func checkField(pass *analysis.Pass, field *ast.Field, markersAccess markers.Markers, conflictSets []ConflictSet, qualifiedFieldName string) {
	if field == nil || len(field.Names) == 0 {
		return
	}

	markers := utils.TypeAwareMarkerCollectionForField(pass, markersAccess, field)

	for _, conflictSet := range conflictSets {
		checkConflict(pass, field, markers, conflictSet, qualifiedFieldName)
	}
}

func checkConflict(pass *analysis.Pass, field *ast.Field, markers markers.MarkerSet, conflictSet ConflictSet, qualifiedFieldName string) {
	// Track which sets have markers present
	conflictingMarkers := make([]sets.Set[string], 0)

	for _, set := range conflictSet.Sets {
		foundMarkers := sets.New[string]()

		for _, markerID := range set {
			if markers.Has(markerID) {
				foundMarkers.Insert(markerID)
			}
		}
		// Only add the set if it has at least one marker
		if foundMarkers.Len() > 0 {
			conflictingMarkers = append(conflictingMarkers, foundMarkers)
		}
	}

	// If two or more sets have markers, report the conflict
	if len(conflictingMarkers) >= 2 {
		reportConflict(pass, field, conflictSet, conflictingMarkers, qualifiedFieldName)
	}
}

func reportConflict(pass *analysis.Pass, field *ast.Field, conflictSet ConflictSet, conflictingMarkers []sets.Set[string], qualifiedFieldName string) {
	// Build a descriptive message showing which sets conflict
	setDescriptions := make([]string, 0, len(conflictingMarkers))

	for _, set := range conflictingMarkers {
		markersList := sets.List(set)
		setDescriptions = append(setDescriptions, fmt.Sprintf("%v", markersList))
	}

	message := fmt.Sprintf("field %s has conflicting markers: %s: {%s}. %s",
		qualifiedFieldName,
		conflictSet.Name,
		strings.Join(setDescriptions, ", "),
		conflictSet.Description)

	pass.Report(analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: message,
	})
}
