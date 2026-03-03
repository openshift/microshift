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
package duplicatemarkers

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
)

const (
	name = "duplicatemarkers"
)

// Analyzer is the analyzer for the duplicatemarkers package.
// It checks for duplicate markers on struct fields.
var Analyzer = &analysis.Analyzer{
	Name:     name,
	Doc:      "Check for duplicate markers on defined types and struct fields.",
	Run:      run,
	Requires: []*analysis.Analyzer{inspector.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string) {
		checkField(pass, field, markersAccess, qualifiedFieldName)
	})

	inspect.InspectTypeSpec(func(typeSpec *ast.TypeSpec, markersAccess markers.Markers) {
		checkTypeSpec(pass, typeSpec, markersAccess)
	})

	return nil, nil //nolint:nilnil
}

func checkField(pass *analysis.Pass, field *ast.Field, markersAccess markers.Markers, qualifiedFieldName string) {
	if field == nil || len(field.Names) == 0 {
		return
	}

	markerSet := utils.TypeAwareMarkerCollectionForField(pass, markersAccess, field)

	seen := markers.NewMarkerSet()

	for _, marker := range markerSet.UnsortedList() {
		if !seen.HasWithValue(marker.String()) {
			seen.Insert(marker)
			continue
		}

		report(pass, field.Pos(), qualifiedFieldName, marker)
	}
}

func checkTypeSpec(pass *analysis.Pass, typeSpec *ast.TypeSpec, markersAccess markers.Markers) {
	if typeSpec == nil {
		return
	}

	typeMarkers := markersAccess.TypeMarkers(typeSpec)

	markerSet := markers.NewMarkerSet()

	for _, marker := range typeMarkers.UnsortedList() {
		if !markerSet.HasWithValue(marker.String()) {
			markerSet.Insert(marker)
			continue
		}

		report(pass, typeSpec.Pos(), typeSpec.Name.Name, marker)
	}
}

func report(pass *analysis.Pass, pos token.Pos, fieldName string, marker markers.Marker) {
	pass.Report(analysis.Diagnostic{
		Pos:     pos,
		Message: fmt.Sprintf("%s has duplicated markers %s", fieldName, marker),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: fmt.Sprintf("Remove duplicated marker %s", marker),
				TextEdits: []analysis.TextEdit{
					{
						Pos: marker.Pos,
						// To remove the duplicated marker, we need to remove the whole line.
						End: marker.End + 1,
					},
				},
			},
		},
	})
}
