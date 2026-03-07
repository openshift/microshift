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
package statusoptional

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"

	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	markershelper "sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
	"sigs.k8s.io/kube-api-linter/pkg/markers"
)

const (
	name = "statusoptional"

	statusJSONTag = "status"
)

func init() {
	markershelper.DefaultRegistry().Register(
		markers.OptionalMarker,
		markers.KubebuilderOptionalMarker,
		markers.K8sOptionalMarker,
		markers.RequiredMarker,
		markers.KubebuilderRequiredMarker,
		markers.K8sRequiredMarker,
	)
}

type analyzer struct {
	preferredOptionalMarker string
}

// newAnalyzer creates a new analyzer.
func newAnalyzer(preferredOptionalMarker string) *analysis.Analyzer {
	if preferredOptionalMarker == "" {
		preferredOptionalMarker = markers.OptionalMarker
	}

	a := &analyzer{
		preferredOptionalMarker: preferredOptionalMarker,
	}

	return &analysis.Analyzer{
		Name:     name,
		Doc:      "Checks that all first-level children fields within status struct are marked as optional",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer, extractjsontags.Analyzer},
	}
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	jsonTags, ok := pass.ResultOf[extractjsontags.Analyzer].(extractjsontags.StructFieldTags)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetJSONTags
	}

	inspect.InspectFields(func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markershelper.Markers, _ string) {
		if jsonTagInfo.Name != statusJSONTag {
			return
		}

		statusStructType := getStructFromField(pass, field)
		a.checkStatusStruct(pass, statusStructType, markersAccess, jsonTags)
	})

	return nil, nil //nolint:nilnil
}

func (a *analyzer) checkStatusStruct(pass *analysis.Pass, statusType *ast.StructType, markersAccess markershelper.Markers, jsonTags extractjsontags.StructFieldTags) {
	if statusType == nil || statusType.Fields == nil || statusType.Fields.List == nil {
		return
	}

	// Check each child field of the status struct
	for _, childField := range statusType.Fields.List {
		fieldName := utils.FieldName(childField)
		jsonTagInfo := jsonTags.FieldTags(childField)

		switch {
		case fieldName == "", jsonTagInfo.Ignored:
			// Skip fields that are ignored or have no name
		case jsonTagInfo.Inline:
			if len(childField.Names) > 0 {
				// Inline fields should not have names
				continue
			}
			// Check embedded structs recursively
			a.checkStatusStruct(pass, getStructFromField(pass, childField), markersAccess, jsonTags)
		default:
			// Check if the field has the required optional markers
			a.checkFieldOptionalMarker(pass, childField, fieldName, markersAccess)
		}
	}
}

// checkFieldOptionalMarker checks if a field has the required optional markers.
// If the field has a required marker, it will be replaced with the preferred optional marker.
// If the field does not have an optional marker, it will be added.
func (a *analyzer) checkFieldOptionalMarker(pass *analysis.Pass, field *ast.Field, fieldName string, markersAccess markershelper.Markers) {
	fieldMarkers := markersAccess.FieldMarkers(field)

	// Check if the field has either the optional or kubebuilder:validation:Optional marker
	if hasOptionalMarker(fieldMarkers) {
		return
	}

	// Check if the field has required markers that need to be replaced
	if hasRequiredMarker(fieldMarkers) {
		a.reportAndReplaceRequiredMarkers(pass, field, fieldName, fieldMarkers)
	} else {
		// Report the error and suggest a fix to add the optional marker
		a.reportAndAddOptionalMarker(pass, field, fieldName)
	}
}

// hasOptionalMarker checks if a field has any optional marker.
func hasOptionalMarker(fieldMarkers markershelper.MarkerSet) bool {
	return fieldMarkers.Has(markers.OptionalMarker) ||
		fieldMarkers.Has(markers.KubebuilderOptionalMarker) ||
		fieldMarkers.Has(markers.K8sOptionalMarker)
}

// hasRequiredMarker checks if a field has any required marker.
func hasRequiredMarker(fieldMarkers markershelper.MarkerSet) bool {
	return fieldMarkers.Has(markers.RequiredMarker) ||
		fieldMarkers.Has(markers.KubebuilderRequiredMarker) ||
		fieldMarkers.Has(markers.K8sRequiredMarker)
}

// reportAndReplaceRequiredMarkers reports an error and suggests replacing required markers with optional ones.
func (a *analyzer) reportAndReplaceRequiredMarkers(pass *analysis.Pass, field *ast.Field, fieldName string, fieldMarkers markershelper.MarkerSet) {
	textEdits := createMarkerRemovalEdits(fieldMarkers)

	// Add the preferred optional marker at the beginning of the field
	textEdits = append(textEdits, analysis.TextEdit{
		Pos:     field.Pos(),
		NewText: fmt.Appendf(nil, "// +%s\n", a.preferredOptionalMarker),
	})

	pass.Report(analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: fmt.Sprintf("status field %q must be marked as optional, not required", fieldName),
		SuggestedFixes: []analysis.SuggestedFix{{
			Message:   fmt.Sprintf("replace required marker(s) with %s", a.preferredOptionalMarker),
			TextEdits: textEdits,
		}},
	})
}

// reportAndAddOptionalMarker reports an error and suggests adding an optional marker.
// TODO: consolidate the logic for removing markers with other linters.
func (a *analyzer) reportAndAddOptionalMarker(pass *analysis.Pass, field *ast.Field, fieldName string) {
	pass.Report(analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: fmt.Sprintf("status field %q must be marked as optional", fieldName),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "add the optional marker",
				TextEdits: []analysis.TextEdit{
					{
						// Position at the beginning of the line of the field
						Pos: field.Pos(),
						// Insert the marker before the field
						NewText: fmt.Appendf(nil, "// +%s\n", a.preferredOptionalMarker),
					},
				},
			},
		},
	})
}

// createMarkerRemovalEdits creates text edits to remove required markers.
// TODO: consolidate the logic for removing markers with other linters.
func createMarkerRemovalEdits(fieldMarkers markershelper.MarkerSet) []analysis.TextEdit {
	var textEdits []analysis.TextEdit

	// Handle standard required markers
	if fieldMarkers.Has(markers.RequiredMarker) {
		for _, marker := range fieldMarkers[markers.RequiredMarker] {
			textEdits = append(textEdits, analysis.TextEdit{
				Pos:     marker.Pos,
				End:     marker.End + 1,
				NewText: []byte(""),
			})
		}
	}

	// Handle kubebuilder required markers
	if fieldMarkers.Has(markers.KubebuilderRequiredMarker) {
		for _, marker := range fieldMarkers[markers.KubebuilderRequiredMarker] {
			textEdits = append(textEdits, analysis.TextEdit{
				Pos:     marker.Pos,
				End:     marker.End + 1,
				NewText: []byte(""),
			})
		}
	}

	// Handle k8s required markers
	if fieldMarkers.Has(markers.K8sRequiredMarker) {
		for _, marker := range fieldMarkers[markers.K8sRequiredMarker] {
			textEdits = append(textEdits, analysis.TextEdit{
				Pos:     marker.Pos,
				End:     marker.End + 1,
				NewText: []byte(""),
			})
		}
	}

	return textEdits
}

// getStructFromField extracts the struct type from an AST Field.
func getStructFromField(pass *analysis.Pass, field *ast.Field) *ast.StructType {
	ident, ok := field.Type.(*ast.Ident)
	if !ok {
		return nil
	}

	typeSpec, ok := utils.LookupTypeSpec(pass, ident)
	if !ok {
		return nil
	}

	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return nil
	}

	return structType
}
