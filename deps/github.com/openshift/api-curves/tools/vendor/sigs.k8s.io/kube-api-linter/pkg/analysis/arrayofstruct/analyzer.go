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
package arrayofstruct

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

const name = "arrayofstruct"

// Analyzer is the analyzer for the arrayofstruct package.
// It checks that arrays containing structs have at least one required field.
var Analyzer = &analysis.Analyzer{
	Name:     name,
	Doc:      "Arrays containing structs must have at least one required field to prevent ambiguous YAML representations",
	Run:      run,
	Requires: []*analysis.Analyzer{inspector.Analyzer},
}

func init() {
	markershelper.DefaultRegistry().Register(markers.KubebuilderExactlyOneOf)
}

func run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markershelper.Markers, qualifiedFieldName string) {
		checkField(pass, field, markersAccess, qualifiedFieldName)
	})

	return nil, nil //nolint:nilnil
}

func checkField(pass *analysis.Pass, field *ast.Field, markersAccess markershelper.Markers, qualifiedFieldName string) {
	// Get the element type of the array
	elementType := getArrayElementType(pass, field)
	if elementType == nil {
		return
	}

	// Check if this is an array of objects (not primitives)
	if !isObjectType(pass, elementType) {
		return
	}

	// Handle pointer types (e.g., []*MyStruct)
	if starExpr, ok := elementType.(*ast.StarExpr); ok {
		elementType = starExpr.X
	}

	// Get the struct type definition
	structType := getStructType(pass, elementType)
	if structType == nil {
		return
	}

	// Check if the struct has union markers that satisfy the required constraint
	if hasExactlyOneOfMarker(structType, markersAccess) {
		// ExactlyOneOf marker enforces that exactly one field is set,
		// so we don't need to report an error
		return
	}

	// Check if at least one field in the struct has a required marker
	if hasRequiredField(structType, markersAccess) {
		return
	}

	reportArrayOfStructIssue(pass, field, qualifiedFieldName)
}

// getArrayElementType extracts the element type from an array field.
// Returns nil if the field is not an array.
func getArrayElementType(pass *analysis.Pass, field *ast.Field) ast.Expr {
	switch fieldType := field.Type.(type) {
	case *ast.ArrayType:
		return fieldType.Elt
	case *ast.Ident:
		// For type aliases to arrays, we need to resolve the underlying type
		typeSpec, ok := utils.LookupTypeSpec(pass, fieldType)
		if !ok {
			return nil
		}

		arrayType, ok := typeSpec.Type.(*ast.ArrayType)
		if !ok {
			return nil
		}

		return arrayType.Elt
	default:
		return nil
	}
}

// reportArrayOfStructIssue reports a diagnostic for an array of structs without required fields.
func reportArrayOfStructIssue(pass *analysis.Pass, field *ast.Field, qualifiedFieldName string) {
	message := fmt.Sprintf("%s is an array of structs, but the struct has no required fields. At least one field should be marked as required to prevent ambiguous YAML configurations", qualifiedFieldName)
	pass.Report(analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: message,
	})
}

// isObjectType checks if the given expression represents an object type (not a primitive).
func isObjectType(pass *analysis.Pass, expr ast.Expr) bool {
	switch et := expr.(type) {
	case *ast.StructType:
		// Inline struct definition
		return true
	case *ast.Ident:
		// Check if it's a basic type
		if utils.IsBasicType(pass, et) {
			return false
		}
		// It's a named type, check if it's a struct
		typeSpec, ok := utils.LookupTypeSpec(pass, et)
		if !ok {
			// Might be from another package, assume it's an object
			return true
		}
		// Recursively check the underlying type
		return isObjectType(pass, typeSpec.Type)
	case *ast.StarExpr:
		// Pointer to something, check what it points to
		return isObjectType(pass, et.X)
	case *ast.SelectorExpr:
		// Type from another package, we can't inspect it
		// Return false to be conservative and skip checking these fields
		return false
	default:
		return false
	}
}

// getStructType resolves the given expression to a struct type,
// following type aliases and handling inline structs.
func getStructType(pass *analysis.Pass, expr ast.Expr) *ast.StructType {
	switch et := expr.(type) {
	case *ast.StructType:
		// Inline struct definition
		return et
	case *ast.Ident:
		// Check if it's a basic type - exit condition for recursion
		if utils.IsBasicType(pass, et) {
			return nil
		}

		// Named struct type or type alias
		typeSpec, ok := utils.LookupTypeSpec(pass, et)
		if !ok {
			// This might be a type from another package or a built-in type
			// In this case, we can't inspect it, so we return nil
			return nil
		}

		// Check if the typeSpec.Type is a struct
		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
			return structType
		}

		// If not a struct, it might be an alias to another type
		// Recursively resolve it
		return getStructType(pass, typeSpec.Type)
	case *ast.SelectorExpr:
		// Type from another package, we can't inspect it
		return nil
	default:
		return nil
	}
}

// hasRequiredField checks if at least one field in the struct has a required marker.
func hasRequiredField(structType *ast.StructType, markersAccess markershelper.Markers) bool {
	if structType.Fields == nil {
		return false
	}

	for _, field := range structType.Fields.List {
		if utils.IsFieldRequired(field, markersAccess) {
			return true
		}
	}

	return false
}

// hasExactlyOneOfMarker checks if the struct has an ExactlyOneOf marker,
// which satisfies the required field constraint by ensuring exactly one field is set.
func hasExactlyOneOfMarker(structType *ast.StructType, markersAccess markershelper.Markers) bool {
	if structType == nil {
		return false
	}

	// Use StructMarkers to get the set of markers on the struct
	markerSet := markersAccess.StructMarkers(structType)

	return markerSet.Has(markers.KubebuilderExactlyOneOf)
}
