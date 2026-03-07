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
package minlength

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
	name = "minlength"
)

// Analyzer is the analyzer for the minlength package.
// It checks that strings and arrays have minimum lengths and minimum items respectively.
var Analyzer = &analysis.Analyzer{
	Name:     name,
	Doc:      "Checks that all strings formatted fields are marked with a minimum length, and that arrays are marked with min items, maps are marked with min properties, and structs that do not have required fields are marked with min properties",
	Run:      run,
	Requires: []*analysis.Analyzer{inspector.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, markersAccess markershelper.Markers, qualifiedFieldName string) {
		checkField(pass, field, markersAccess, qualifiedFieldName)
	})

	return nil, nil //nolint:nilnil
}

func checkField(pass *analysis.Pass, field *ast.Field, markersAccess markershelper.Markers, qualifiedFieldName string) {
	prefix := fmt.Sprintf("field %s", qualifiedFieldName)

	checkTypeExpr(pass, field.Type, field, nil, markersAccess, prefix, markers.KubebuilderMinLengthMarker, needsStringMinLength)
}

func checkIdent(pass *analysis.Pass, ident *ast.Ident, node ast.Node, aliases []*ast.TypeSpec, markersAccess markershelper.Markers, prefix, marker string, needsMaxLength func(markershelper.MarkerSet) bool) {
	if utils.IsBasicType(pass, ident) { // Built-in type
		checkString(pass, ident, node, aliases, markersAccess, prefix, marker, needsMaxLength)

		return
	}

	tSpec, ok := utils.LookupTypeSpec(pass, ident)
	if !ok {
		return
	}

	checkTypeSpec(pass, tSpec, node, append(aliases, tSpec), markersAccess, fmt.Sprintf("%s type", prefix), marker, needsMaxLength)
}

func checkString(pass *analysis.Pass, ident *ast.Ident, node ast.Node, aliases []*ast.TypeSpec, markersAccess markershelper.Markers, prefix, marker string, needsMinLength func(markershelper.MarkerSet) bool) {
	if ident.Name != "string" {
		return
	}

	markers := getCombinedMarkers(markersAccess, node, aliases)

	if needsMinLength(markers) {
		pass.Reportf(node.Pos(), "%s must have a minimum length, add %s marker", prefix, marker)
	}
}

func checkTypeSpec(pass *analysis.Pass, tSpec *ast.TypeSpec, node ast.Node, aliases []*ast.TypeSpec, markersAccess markershelper.Markers, prefix, marker string, needsMinLength func(markershelper.MarkerSet) bool) {
	if tSpec.Name == nil {
		return
	}

	typeName := tSpec.Name.Name
	prefix = fmt.Sprintf("%s %s", prefix, typeName)

	checkTypeExpr(pass, tSpec.Type, node, aliases, markersAccess, prefix, marker, needsMinLength)
}

func checkTypeExpr(pass *analysis.Pass, typeExpr ast.Expr, node ast.Node, aliases []*ast.TypeSpec, markersAccess markershelper.Markers, prefix, marker string, needsMinLength func(markershelper.MarkerSet) bool) {
	switch typ := typeExpr.(type) {
	case *ast.Ident:
		checkIdent(pass, typ, node, aliases, markersAccess, prefix, marker, needsMinLength)
	case *ast.StarExpr:
		checkTypeExpr(pass, typ.X, node, aliases, markersAccess, prefix, marker, needsMinLength)
	case *ast.ArrayType:
		checkArrayType(pass, typ, node, aliases, markersAccess, prefix)
	case *ast.MapType:
		checkMapType(pass, node, aliases, markersAccess, prefix)
	case *ast.StructType:
		checkStructType(pass, typ, node, aliases, markersAccess, prefix)
	}
}

func checkArrayType(pass *analysis.Pass, arrayType *ast.ArrayType, node ast.Node, aliases []*ast.TypeSpec, markersAccess markershelper.Markers, prefix string) {
	if arrayType.Elt != nil {
		if ident, ok := arrayType.Elt.(*ast.Ident); ok {
			if ident.Name == "byte" {
				// byte slices are a special case as they are treated as strings.
				// Pretend the ident is a string so that checkString can process it as expected.
				i := &ast.Ident{
					NamePos: ident.NamePos,
					Name:    "string",
				}
				checkString(pass, i, node, aliases, markersAccess, prefix, markers.KubebuilderMinLengthMarker, needsStringMinLength)

				return
			}

			checkArrayElementIdent(pass, ident, node, aliases, markersAccess, fmt.Sprintf("%s array element", prefix))
		}
	}

	markerSet := getCombinedMarkers(markersAccess, node, aliases)

	if !markerSet.Has(markers.KubebuilderMinItemsMarker) {
		pass.Reportf(node.Pos(), "%s must have a minimum items, add %s marker", prefix, markers.KubebuilderMinItemsMarker)
	}
}

func checkArrayElementIdent(pass *analysis.Pass, ident *ast.Ident, node ast.Node, aliases []*ast.TypeSpec, markersAccess markershelper.Markers, prefix string) {
	if ident.Obj == nil { // Built-in type
		checkString(pass, ident, node, aliases, markersAccess, prefix, markers.KubebuilderItemsMinLengthMarker, needsItemsMinLength)

		return
	}

	tSpec, ok := ident.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		return
	}

	// If the array element wasn't directly a string, allow a string alias to be used
	// with either the items style markers or the on alias style markers.
	checkTypeSpec(pass, tSpec, node, append(aliases, tSpec), markersAccess, fmt.Sprintf("%s type", prefix), markers.KubebuilderMinLengthMarker, func(ms markershelper.MarkerSet) bool {
		return needsStringMinLength(ms) && needsItemsMinLength(ms)
	})
}

func checkMapType(pass *analysis.Pass, node ast.Node, aliases []*ast.TypeSpec, markersAccess markershelper.Markers, prefix string) {
	markerSet := getCombinedMarkers(markersAccess, node, aliases)

	if !markerSet.Has(markers.KubebuilderMinPropertiesMarker) {
		pass.Reportf(node.Pos(), "%s must have a minimum properties, add %s marker", prefix, markers.KubebuilderMinPropertiesMarker)
	}
}

func checkStructType(pass *analysis.Pass, structType *ast.StructType, node ast.Node, aliases []*ast.TypeSpec, markersAccess markershelper.Markers, prefix string) {
	markerSet := getCombinedMarkers(markersAccess, node, aliases)

	minProperties, err := utils.GetMinProperties(markerSet)
	if err != nil {
		pass.Reportf(node.Pos(), "could not get min properties for struct: %v", err)
		return
	}

	if minProperties != nil {
		// There's already a min properties specified.
		return
	}

	// Check if the struct has union markers that satisfy the required constraint
	if markerSet.Has(markers.KubebuilderExactlyOneOf) || markerSet.Has(markers.KubebuilderAtLeastOneOfMarker) {
		// ExactlyOneOf / AtLeastOneOf markers enforce that at least one field is required,
		// this means that `{}` is not valid.
		return
	}

	for _, field := range structType.Fields.List {
		if utils.IsFieldRequired(field, markersAccess) {
			// The struct has at least one required field,
			// this means that `{}` is not valid.
			return
		}
	}

	// The field does not have a min properties, and does not have any required fields.
	pass.Reportf(node.Pos(), "%s must have a minimum properties, add %s marker", prefix, markers.KubebuilderMinPropertiesMarker)
}

func getCombinedMarkers(markersAccess markershelper.Markers, node ast.Node, aliases []*ast.TypeSpec) markershelper.MarkerSet {
	base := markershelper.NewMarkerSet(getMarkers(markersAccess, node).UnsortedList()...)

	for _, a := range aliases {
		base.Insert(getMarkers(markersAccess, a).UnsortedList()...)
	}

	return base
}

func getMarkers(markersAccess markershelper.Markers, node ast.Node) markershelper.MarkerSet {
	switch t := node.(type) {
	case *ast.Field:
		return markersAccess.FieldMarkers(t)
	case *ast.TypeSpec:
		return markersAccess.TypeMarkers(t)
	}

	return nil
}

// needsMinLength returns true if the field needs a minimum length.
// Fields do not need a minimum length if they are already marked with a minimum length,
// or if they are an enum, or if they are a date, date-time or duration.
func needsStringMinLength(markerSet markershelper.MarkerSet) bool {
	switch {
	case markerSet.Has(markers.KubebuilderMinLengthMarker),
		markerSet.Has(markers.KubebuilderEnumMarker),
		markerSet.HasWithValue(kubebuilderFormatWithValue("date")),
		markerSet.HasWithValue(kubebuilderFormatWithValue("date-time")),
		markerSet.HasWithValue(kubebuilderFormatWithValue("duration")):
		return false
	}

	return true
}

func needsItemsMinLength(markerSet markershelper.MarkerSet) bool {
	switch {
	case markerSet.Has(markers.KubebuilderItemsMinLengthMarker),
		markerSet.Has(markers.KubebuilderItemsEnumMarker),
		markerSet.HasWithValue(kubebuilderItemsFormatWithValue("date")),
		markerSet.HasWithValue(kubebuilderItemsFormatWithValue("date-time")),
		markerSet.HasWithValue(kubebuilderItemsFormatWithValue("duration")):
		return false
	}

	return true
}

func kubebuilderFormatWithValue(value string) string {
	return fmt.Sprintf("%s:=%s", markers.KubebuilderFormatMarker, value)
}

func kubebuilderItemsFormatWithValue(value string) string {
	return fmt.Sprintf("%s:=%s", markers.KubebuilderItemsFormatMarker, value)
}
