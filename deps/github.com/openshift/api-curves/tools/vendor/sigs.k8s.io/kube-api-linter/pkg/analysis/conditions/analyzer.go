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
package conditions

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
)

const (
	name = "conditions"

	listTypeMarkerID      = "listType"
	listMapKeyMarkerID    = "listMapKey"
	patchStrategyMarkerID = "patchStrategy"
	patchMergeKeyMarkerID = "patchMergeKey"

	listTypeMap        = "listType=map"
	listMapKeyType     = "listMapKey=type"
	patchStrategy      = "patchStrategy"
	patchStrategyMerge = "patchStrategy=merge"
	patchMergeKey      = "patchMergeKey"
	patchMergeKeyType  = "patchMergeKey=type"
	optional           = "optional"

	expectedJSONTag     = "json:\"conditions,omitempty\""
	expectedPatchTag    = "patchStrategy:\"merge\" patchMergeKey:\"type\""
	expectedProtobufTag = "protobuf:\"bytes,%d,rep,name=conditions\""
)

func init() {
	markers.DefaultRegistry().Register(
		listTypeMarkerID,
		listMapKeyMarkerID,
		patchStrategyMarkerID,
		patchMergeKeyMarkerID,
		optional,
	)
}

type analyzer struct {
	isFirstField     ConditionsFirstField
	useProtobuf      ConditionsUseProtobuf
	usePatchStrategy ConditionsUsePatchStrategy
}

// newAnalyzer creates a new analyzer.
func newAnalyzer(cfg *ConditionsConfig) *analysis.Analyzer {
	if cfg == nil {
		cfg = &ConditionsConfig{}
	}

	defaultConfig(cfg)

	a := &analyzer{
		isFirstField:     cfg.IsFirstField,
		useProtobuf:      cfg.UseProtobuf,
		usePatchStrategy: cfg.UsePatchStrategy,
	}

	return &analysis.Analyzer{
		Name:     name,
		Doc:      `Checks that all conditions type fields conform to the required conventions.`,
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer, markers.Analyzer, extractjsontags.Analyzer},
	}
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	markersAccess, ok := pass.ResultOf[markers.Analyzer].(markers.Markers)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetMarkers
	}

	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		tSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return
		}

		sTyp, ok := tSpec.Type.(*ast.StructType)
		if !ok {
			return
		}

		if sTyp.Fields == nil {
			return
		}

		structName := tSpec.Name.Name

		for i, field := range sTyp.Fields.List {
			fieldMarkers := markersAccess.FieldMarkers(field)

			a.checkField(pass, i, field, fieldMarkers, structName)
		}
	})

	return nil, nil //nolint:nilnil
}

func (a *analyzer) checkField(pass *analysis.Pass, index int, field *ast.Field, fieldMarkers markers.MarkerSet, structName string) {
	if !fieldIsCalledConditions(field) {
		return
	}

	if !isSliceMetaV1Condition(field) {
		pass.Reportf(field.Pos(), "Conditions field in %s must be a slice of metav1.Condition", structName)
		return
	}

	checkFieldMarkers(pass, field, fieldMarkers, a.usePatchStrategy, structName)
	a.checkFieldTags(pass, index, field, structName)

	if a.isFirstField == ConditionsFirstFieldWarn && index != 0 {
		pass.Reportf(field.Pos(), "Conditions field in %s must be the first field in the struct", structName)
	}
}

func checkFieldMarkers(pass *analysis.Pass, field *ast.Field, fieldMarkers markers.MarkerSet, usePatchStrategy ConditionsUsePatchStrategy, structName string) {
	missingMarkers := []string{}
	additionalMarkers := []markers.Marker{}

	if !fieldMarkers.HasWithValue(listTypeMap) {
		missingMarkers = append(missingMarkers, listTypeMap)
	}

	if !fieldMarkers.HasWithValue(listMapKeyType) {
		missingMarkers = append(missingMarkers, listMapKeyType)
	}

	patchMissingMarkers, patchAdditionalMarkers := checkPatchStrategyMarkers(fieldMarkers, usePatchStrategy)
	missingMarkers = append(missingMarkers, patchMissingMarkers...)
	additionalMarkers = append(additionalMarkers, patchAdditionalMarkers...)

	if !fieldMarkers.Has(optional) {
		missingMarkers = append(missingMarkers, optional)
	}

	if len(missingMarkers) != 0 {
		reportMissingMarkers(pass, field, missingMarkers, usePatchStrategy, structName)
	}

	if len(additionalMarkers) != 0 {
		reportAdditionalMarkers(pass, field, additionalMarkers, structName)
	}
}

func checkPatchStrategyMarkers(fieldMarkers markers.MarkerSet, usePatchStrategy ConditionsUsePatchStrategy) ([]string, []markers.Marker) {
	missingMarkers := []string{}
	additionalMarkers := []markers.Marker{}

	switch usePatchStrategy {
	case ConditionsUsePatchStrategySuggestFix, ConditionsUsePatchStrategyWarn:
		if !fieldMarkers.HasWithValue(patchStrategyMerge) {
			missingMarkers = append(missingMarkers, patchStrategyMerge)
		}

		if !fieldMarkers.HasWithValue(patchMergeKeyType) {
			missingMarkers = append(missingMarkers, patchMergeKeyType)
		}
	case ConditionsUsePatchStrategyIgnore:
		// If it's there, we don't care.
	case ConditionsUsePatchStrategyForbid:
		if fieldMarkers.HasWithValue(patchStrategyMerge) {
			additionalMarkers = append(additionalMarkers, fieldMarkers[patchStrategy]...)
		}

		if fieldMarkers.HasWithValue(patchMergeKeyType) {
			additionalMarkers = append(additionalMarkers, fieldMarkers[patchMergeKey]...)
		}
	default:
		panic("unexpected usePatchStrategy value")
	}

	return missingMarkers, additionalMarkers
}

func reportMissingMarkers(pass *analysis.Pass, field *ast.Field, missingMarkers []string, usePatchStrategy ConditionsUsePatchStrategy, structName string) {
	suggestedFixes := []analysis.SuggestedFix{}

	// If patch strategy is warn, and the only markers in the list are patchStrategy and patchMergeKeyType, we don't need to suggest a fix.
	if usePatchStrategy != ConditionsUsePatchStrategyWarn || slices.ContainsFunc(missingMarkers, func(marker string) bool {
		switch marker {
		case patchStrategyMerge, patchMergeKeyType:
			return false
		default:
			return true
		}
	}) {
		suggestedFixes = []analysis.SuggestedFix{
			{
				Message: "Add missing markers",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     field.Pos(),
						End:     token.NoPos,
						NewText: getNewMarkers(missingMarkers),
					},
				},
			},
		}
	}

	pass.Report(analysis.Diagnostic{
		Pos:            field.Pos(),
		End:            field.End(),
		Message:        "Conditions field in " + structName + " is missing the following markers: " + strings.Join(missingMarkers, ", "),
		SuggestedFixes: suggestedFixes,
	})
}

func reportAdditionalMarkers(pass *analysis.Pass, field *ast.Field, additionalMarkers []markers.Marker, structName string) {
	suggestedFixes := []analysis.SuggestedFix{}
	additionalMarkerValues := []string{}

	for _, marker := range additionalMarkers {
		additionalMarkerValues = append(additionalMarkerValues, marker.String())

		suggestedFixes = append(suggestedFixes, analysis.SuggestedFix{
			Message: fmt.Sprintf("Remove additional marker %s", marker.String()),
			TextEdits: []analysis.TextEdit{
				{
					Pos:     marker.Pos,
					End:     marker.End + 1, // Add 1 to position to include the new line
					NewText: nil,
				},
			},
		})
	}

	pass.Report(analysis.Diagnostic{
		Pos:            field.Pos(),
		End:            field.End(),
		Message:        fmt.Sprintf("Conditions field in %s has the following additional markers: %s", structName, strings.Join(additionalMarkerValues, ", ")),
		SuggestedFixes: suggestedFixes,
	})
}

func getNewMarkers(missingMarkers []string) []byte {
	var out string

	for _, marker := range missingMarkers {
		out += "// +" + marker + "\n"
	}

	return []byte(out)
}

func (a *analyzer) checkFieldTags(pass *analysis.Pass, index int, field *ast.Field, structName string) {
	if field.Tag == nil {
		expectedTag := getExpectedTag(a.usePatchStrategy, a.useProtobuf, a.isFirstField, index)

		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			End:     field.End(),
			Message: fmt.Sprintf("Conditions field in %s is missing tags, should be: %s", structName, expectedTag),
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: fmt.Sprintf("Add missing tags: %s", expectedTag),
					TextEdits: []analysis.TextEdit{
						{
							Pos:     field.End(),
							End:     token.NoPos,
							NewText: []byte(expectedTag),
						},
					},
				},
			},
		})

		return
	}

	asExpected, shouldFix := tagIsAsExpected(field.Tag.Value, a.usePatchStrategy, a.useProtobuf, a.isFirstField, index)
	if !asExpected {
		expectedTag := getExpectedTag(a.usePatchStrategy, a.useProtobuf, a.isFirstField, index)

		if !shouldFix {
			pass.Reportf(field.Tag.ValuePos, "Conditions field in %s has incorrect tags, should be: %s", structName, expectedTag)
		} else {
			pass.Report(analysis.Diagnostic{
				Pos:     field.Tag.ValuePos,
				End:     field.Tag.End(),
				Message: fmt.Sprintf("Conditions field in %s has incorrect tags, should be: %s", structName, expectedTag),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: fmt.Sprintf("Update tags to: %s", expectedTag),
						TextEdits: []analysis.TextEdit{
							{
								Pos:     field.Tag.ValuePos,
								End:     field.Tag.End(),
								NewText: []byte(expectedTag),
							},
						},
					},
				},
			})
		}
	}
}

func getExpectedTag(usePatchStrategy ConditionsUsePatchStrategy, useProtobuf ConditionsUseProtobuf, isFirstField ConditionsFirstField, index int) string {
	expectedTag := fmt.Sprintf("`%s", expectedJSONTag)

	if usePatchStrategy == ConditionsUsePatchStrategySuggestFix || usePatchStrategy == ConditionsUsePatchStrategyWarn {
		expectedTag += fmt.Sprintf(" %s", expectedPatchTag)
	}

	if useProtobuf == ConditionsUseProtobufSuggestFix || useProtobuf == ConditionsUseProtobufWarn {
		expectedTag += fmt.Sprintf(" %s", getExpectedProtobufTag(isFirstField, index))
	}

	expectedTag += "`"

	return expectedTag
}

func getExpectedProtobufTag(isFirstField ConditionsFirstField, index int) string {
	i := 1
	if isFirstField == ConditionsFirstFieldIgnore {
		i = index + 1
	}

	return fmt.Sprintf(expectedProtobufTag, i)
}

func tagIsAsExpected(tag string, usePatchStrategy ConditionsUsePatchStrategy, useProtobuf ConditionsUseProtobuf, isFirstField ConditionsFirstField, index int) (bool, bool) {
	patchTagCorrect, patchShouldSuggestFix := patchStrategyTagIsAsExpected(tag, usePatchStrategy)
	protoTagCorrect, protoShouldSuggestFix := protobufTagIsAsExpected(tag, useProtobuf, isFirstField, index)

	return patchTagCorrect && protoTagCorrect, patchShouldSuggestFix || protoShouldSuggestFix
}

func patchStrategyTagIsAsExpected(tag string, usePatchStrategy ConditionsUsePatchStrategy) (bool, bool) {
	switch usePatchStrategy {
	case ConditionsUsePatchStrategySuggestFix:
		return strings.Contains(tag, expectedPatchTag), true
	case ConditionsUsePatchStrategyWarn:
		return strings.Contains(tag, expectedPatchTag), false
	case ConditionsUsePatchStrategyIgnore:
		return true, false
	case ConditionsUsePatchStrategyForbid:
		return !strings.Contains(tag, expectedPatchTag), true
	default:
		panic("unexpected usePatchStrategy value")
	}
}

func protobufTagIsAsExpected(tag string, useProtobuf ConditionsUseProtobuf, isFirstField ConditionsFirstField, index int) (bool, bool) {
	switch useProtobuf {
	case ConditionsUseProtobufSuggestFix:
		return strings.Contains(tag, getExpectedProtobufTag(isFirstField, index)), true
	case ConditionsUseProtobufWarn:
		return strings.Contains(tag, getExpectedProtobufTag(isFirstField, index)), false
	case ConditionsUseProtobufIgnore:
		return true, false
	case ConditionsUseProtobufForbid:
		return !strings.Contains(tag, getExpectedProtobufTag(isFirstField, index)), true
	default:
		panic("unexpected useProtobuf value")
	}
}

func fieldIsCalledConditions(field *ast.Field) bool {
	if field == nil {
		return false
	}

	return len(field.Names) != 0 && field.Names[0] != nil && field.Names[0].Name == "Conditions"
}

func isSliceMetaV1Condition(field *ast.Field) bool {
	if field == nil {
		return false
	}

	// Field is not an array type.
	arr, ok := field.Type.(*ast.ArrayType)
	if !ok {
		return false
	}

	// Array element is not imported.
	selector, ok := arr.Elt.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	// Array element is not imported from metav1.
	if selector.X == nil || pkg.Name != "metav1" {
		return false
	}

	// Array element is not a metav1.Condition.
	if selector.Sel == nil || selector.Sel.Name != "Condition" {
		return false
	}

	return true
}

func defaultConfig(cfg *ConditionsConfig) {
	if cfg.IsFirstField == "" {
		cfg.IsFirstField = ConditionsFirstFieldWarn
	}

	if cfg.UseProtobuf == "" {
		cfg.UseProtobuf = ConditionsUseProtobufSuggestFix
	}

	if cfg.UsePatchStrategy == "" {
		cfg.UsePatchStrategy = ConditionsUsePatchStrategySuggestFix
	}
}
