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
package ssatags

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"

	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
	kubebuildermarkers "sigs.k8s.io/kube-api-linter/pkg/markers"
)

const name = "ssatags"

const (
	listTypeAtomic = "atomic"
	listTypeSet    = "set"
	listTypeMap    = "map"
)

type analyzer struct {
	listTypeSetUsage SSATagsListTypeSetUsage
}

func newAnalyzer(cfg *SSATagsConfig) *analysis.Analyzer {
	if cfg == nil {
		cfg = &SSATagsConfig{}
	}

	defaultConfig(cfg)

	a := &analyzer{
		listTypeSetUsage: cfg.ListTypeSetUsage,
	}

	return &analysis.Analyzer{
		Name:     name,
		Doc:      "Check that all array types in the API have a listType tag and the usage of the tags is correct",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer, extractjsontags.Analyzer},
	}
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string) {
		a.checkField(pass, field, markersAccess, qualifiedFieldName)
	})

	return nil, nil //nolint:nilnil
}

func (a *analyzer) checkField(pass *analysis.Pass, field *ast.Field, markersAccess markers.Markers, qualifiedFieldName string) {
	if !utils.IsArrayTypeOrAlias(pass, field) {
		return
	}

	fieldMarkers := utils.TypeAwareMarkerCollectionForField(pass, markersAccess, field)
	if fieldMarkers == nil {
		return
	}

	// If the field is a byte array, we cannot use listType markers with it.
	if utils.IsByteArray(pass, field) {
		listTypeMarkers := fieldMarkers.Get(kubebuildermarkers.KubebuilderListTypeMarker)
		for _, marker := range listTypeMarkers {
			pass.Report(analysis.Diagnostic{
				Pos:     field.Pos(),
				Message: fmt.Sprintf("%s is a byte array, which does not support the listType marker. Remove the listType marker", qualifiedFieldName),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: fmt.Sprintf("Remove listType marker from %s", qualifiedFieldName),
						TextEdits: []analysis.TextEdit{
							{
								Pos:     marker.Pos,
								End:     marker.End + 1,
								NewText: []byte(""),
							},
						},
					},
				},
			})
		}

		return
	}

	listTypeMarkers := fieldMarkers.Get(kubebuildermarkers.KubebuilderListTypeMarker)

	if len(listTypeMarkers) == 0 {
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("%s should have a listType marker for proper Server-Side Apply behavior (atomic, set, or map)", qualifiedFieldName),
		})

		return
	}

	for _, marker := range listTypeMarkers {
		listType := marker.Payload.Value

		a.checkListTypeMarker(pass, listType, field, qualifiedFieldName)

		if listType == listTypeMap {
			a.checkListTypeMap(pass, fieldMarkers, field, qualifiedFieldName)
		}

		if listType == listTypeSet {
			a.checkListTypeSet(pass, field, qualifiedFieldName)
		}
	}
}

func (a *analyzer) checkListTypeMarker(pass *analysis.Pass, listType string, field *ast.Field, qualifiedFieldName string) {
	if !validListType(listType) {
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("%s has invalid listType %q, must be one of: atomic, set, map", qualifiedFieldName, listType),
		})

		return
	}
}

func (a *analyzer) checkListTypeMap(pass *analysis.Pass, fieldMarkers markers.MarkerSet, field *ast.Field, qualifiedFieldName string) {
	listMapKeyMarkers := fieldMarkers.Get(kubebuildermarkers.KubebuilderListMapKeyMarker)

	isObjectList := utils.IsObjectList(pass, field)

	if !isObjectList {
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("%s with listType=map can only be used for object lists, not primitive lists", qualifiedFieldName),
		})

		return
	}

	if len(listMapKeyMarkers) == 0 {
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("%s with listType=map must have at least one listMapKey marker", qualifiedFieldName),
		})

		return
	}

	a.validateListMapKeys(pass, field, listMapKeyMarkers, qualifiedFieldName)
}

func (a *analyzer) checkListTypeSet(pass *analysis.Pass, field *ast.Field, qualifiedFieldName string) {
	if a.listTypeSetUsage == SSATagsListTypeSetUsageIgnore {
		return
	}

	isObjectList := utils.IsObjectList(pass, field)
	if !isObjectList {
		return
	}

	diagnostic := analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: fmt.Sprintf("%s with listType=set is not recommended due to Server-Side Apply compatibility issues. Consider using listType=%s or listType=%s instead", qualifiedFieldName, listTypeAtomic, listTypeMap),
	}

	pass.Report(diagnostic)
}

func (a *analyzer) validateListMapKeys(pass *analysis.Pass, field *ast.Field, listMapKeyMarkers []markers.Marker, qualifiedFieldName string) {
	jsonTags, ok := pass.ResultOf[extractjsontags.Analyzer].(extractjsontags.StructFieldTags)
	if !ok {
		return
	}

	structFields := a.getStructFieldsFromField(pass, jsonTags, field)
	if structFields == nil {
		return
	}

	for _, marker := range listMapKeyMarkers {
		keyName := marker.Payload.Value
		if keyName == "" {
			continue
		}

		if !a.hasFieldWithJSONTag(structFields, jsonTags, keyName) {
			pass.Report(analysis.Diagnostic{
				Pos:     field.Pos(),
				Message: fmt.Sprintf("%s listMapKey %q does not exist as a field in the struct", qualifiedFieldName, keyName),
			})
		}
	}
}

func (a *analyzer) getStructFieldsFromField(pass *analysis.Pass, jsonTags extractjsontags.StructFieldTags, field *ast.Field) *ast.FieldList {
	var elementType ast.Expr

	// Get the element type from array or field type
	if arrayType, ok := field.Type.(*ast.ArrayType); ok {
		elementType = arrayType.Elt
	} else {
		elementType = field.Type
	}

	return a.getStructFieldsFromExpr(pass, jsonTags, elementType)
}

func (a *analyzer) getStructFieldsFromExpr(pass *analysis.Pass, jsonTags extractjsontags.StructFieldTags, expr ast.Expr) *ast.FieldList {
	switch elementType := expr.(type) {
	case *ast.Ident:
		typeSpec, ok := utils.LookupTypeSpec(pass, elementType)
		if !ok {
			return nil
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return nil
		}

		return flattenStructFields(pass, jsonTags, structType.Fields)
	case *ast.StarExpr:
		return a.getStructFieldsFromExpr(pass, jsonTags, elementType.X)
	case *ast.SelectorExpr:
		return nil
	}

	return nil
}

func (a *analyzer) hasFieldWithJSONTag(fields *ast.FieldList, jsonTags extractjsontags.StructFieldTags, fieldName string) bool {
	if fields == nil {
		return false
	}

	for _, field := range fields.List {
		tagInfo := jsonTags.FieldTags(field)

		if tagInfo.Name == fieldName {
			return true
		}
	}

	return false
}

func validListType(listType string) bool {
	switch listType {
	case listTypeAtomic, listTypeSet, listTypeMap:
		return true
	default:
		return false
	}
}

func defaultConfig(cfg *SSATagsConfig) {
	if cfg.ListTypeSetUsage == "" {
		cfg.ListTypeSetUsage = SSATagsListTypeSetUsageWarn
	}
}

// flattenStructFields flattens a struct's fields by looking for embedded structs and promoting their fields to the top level.
// This allows us to correctly check listMapKey markers where the map key is a member of the embedded struct.
func flattenStructFields(pass *analysis.Pass, jsonTags extractjsontags.StructFieldTags, fields *ast.FieldList) *ast.FieldList {
	if fields == nil {
		return nil
	}

	flattenedFields := &ast.FieldList{}

	for _, field := range fields.List {
		tagInfo := jsonTags.FieldTags(field)
		if len(field.Names) > 0 || tagInfo.Name != "" {
			// Field is not embedded, it has an explicit name.
			flattenedFields.List = append(flattenedFields.List, field)
			continue
		}

		ident, ok := field.Type.(*ast.Ident)
		if !ok {
			flattenedFields.List = append(flattenedFields.List, field)
			continue
		}

		typeSpec, ok := utils.LookupTypeSpec(pass, ident)
		if !ok {
			flattenedFields.List = append(flattenedFields.List, field)
			continue
		}

		embeddedStruct, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			flattenedFields.List = append(flattenedFields.List, field)
			continue
		}

		flattenedFields.List = append(flattenedFields.List, flattenStructFields(pass, jsonTags, embeddedStruct.Fields).List...)
	}

	return flattenedFields
}
