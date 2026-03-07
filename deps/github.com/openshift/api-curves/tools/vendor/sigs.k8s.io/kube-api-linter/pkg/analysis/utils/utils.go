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
package utils

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	markershelper "sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/markers"
)

const stringTypeName = "string"

// IsBasicType checks if the type of the given identifier is a basic type.
// Basic types are types like int, string, bool, etc.
func IsBasicType(pass *analysis.Pass, expr ast.Expr) bool {
	_, ok := pass.TypesInfo.TypeOf(expr).(*types.Basic)
	return ok
}

// IsStringType checks if the type of the given expression is a string type..
func IsStringType(pass *analysis.Pass, expr ast.Expr) bool {
	// In case the expr is a pointer.
	underlying := getUnderlyingType(expr)

	ident, ok := underlying.(*ast.Ident)
	if !ok {
		return false
	}

	if ident.Name == stringTypeName {
		return true
	}

	// Is either an alias or another basic type, try to look up the alias.
	tSpec, ok := LookupTypeSpec(pass, ident)
	if !ok {
		// Basic type and not a string.
		return false
	}

	return IsStringType(pass, tSpec.Type)
}

// IsStructType checks if the given expression is a struct type.
func IsStructType(pass *analysis.Pass, expr ast.Expr) bool {
	underlying := getUnderlyingType(expr)

	if _, ok := underlying.(*ast.StructType); ok {
		return true
	}

	// Where there's an ident, recurse to find the underlying type.
	if ident, ok := underlying.(*ast.Ident); ok {
		typeSpec, ok := LookupTypeSpec(pass, ident)
		if !ok {
			return false
		}

		return IsStructType(pass, typeSpec.Type)
	}

	return false
}

// IsStarExpr checks if the expression is a pointer type.
// If it is, it returns the expression inside the pointer.
func IsStarExpr(expr ast.Expr) (bool, ast.Expr) {
	if ptrType, ok := expr.(*ast.StarExpr); ok {
		return true, ptrType.X
	}

	return false, expr
}

// IsPointer checks if the expression is a pointer.
func IsPointer(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}

// IsPointerType checks if the expression is a pointer type.
// This is for types that are always implemented as pointers and therefore should
// not be the underlying type of a star expr.
func IsPointerType(pass *analysis.Pass, expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.StarExpr, *ast.MapType, *ast.ArrayType:
		return true
	case *ast.Ident:
		// If the ident is a type alias, keep checking until we find the underlying type.
		typeSpec, ok := LookupTypeSpec(pass, t)
		if !ok {
			return false
		}

		return IsPointerType(pass, typeSpec.Type)
	default:
		return false
	}
}

// LookupTypeSpec is used to search for the type spec of a given identifier.
// It will first check to see if the ident has an Obj, and if so, it will return the type spec
// from the Obj. If the Obj is nil, it will search through the files in the package to find the
// type spec that matches the identifier's position.
func LookupTypeSpec(pass *analysis.Pass, ident *ast.Ident) (*ast.TypeSpec, bool) {
	if ident.Obj != nil && ident.Obj.Decl != nil {
		// The identifier has an Obj, we can use it to find the type spec.
		if tSpec, ok := ident.Obj.Decl.(*ast.TypeSpec); ok {
			return tSpec, true
		}
	}

	namedType, ok := pass.TypesInfo.TypeOf(ident).(*types.Named)
	if !ok {
		return nil, false
	}

	if !isInPassPackage(pass, namedType) {
		// The identifier is not in the pass package, we can't find the type spec.
		return nil, false
	}

	tokenFile, astFile := getFilesForType(pass, ident)

	if astFile == nil {
		// We couldn't match the token.File to the ast.File.
		return nil, false
	}

	for n := range ast.Preorder(astFile) {
		tSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			continue
		}

		// Token files are 1-based, while ast files are 0-based.
		// We need to adjust the position to match the token file.
		filePos := tSpec.Pos() - astFile.FileStart + token.Pos(tokenFile.Base())

		if filePos == namedType.Obj().Pos() {
			return tSpec, true
		}
	}

	return nil, false
}

// FieldName returns the name of the field. If the field has a name, it returns that name.
// If the field is embedded and it can be converted to an identifier, it returns the name of the identifier.
// If it doesn't have a name and can't be converted to an identifier, it returns an empty string.
func FieldName(field *ast.Field) string {
	if len(field.Names) > 0 && field.Names[0] != nil {
		return field.Names[0].Name
	}

	switch typ := field.Type.(type) {
	case *ast.Ident:
		return typ.Name
	case *ast.StarExpr:
		if ident, ok := typ.X.(*ast.Ident); ok {
			return ident.Name
		}
	}

	return ""
}

// GetStructName returns the name of the struct that the field is in.
func GetStructName(pass *analysis.Pass, field *ast.Field) string {
	_, astFile := getFilesForField(pass, field)
	if astFile == nil {
		return ""
	}

	return GetStructNameFromFile(astFile, field)
}

// GetStructNameFromFile returns the name of the struct that the field is in.
func GetStructNameFromFile(file *ast.File, field *ast.Field) string {
	var (
		structName string
		found      bool
	)

	ast.Inspect(file, func(n ast.Node) bool {
		if found {
			return false
		}

		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structName = typeSpec.Name.Name

		if structType.Fields == nil {
			return true
		}

		if slices.Contains(structType.Fields.List, field) {
			found = true
			return false
		}

		return true
	})

	if found {
		return structName
	}

	return ""
}

// GetQualifiedFieldName returns the qualified field name.
func GetQualifiedFieldName(pass *analysis.Pass, field *ast.Field) string {
	fieldName := FieldName(field)
	structName := GetStructName(pass, field)

	return fmt.Sprintf("%s.%s", structName, fieldName)
}

func getFilesForField(pass *analysis.Pass, field *ast.Field) (*token.File, *ast.File) {
	tokenFile := pass.Fset.File(field.Pos())
	for _, astFile := range pass.Files {
		if astFile.FileStart == token.Pos(tokenFile.Base()) {
			return tokenFile, astFile
		}
	}

	return tokenFile, nil
}

func getFilesForType(pass *analysis.Pass, ident *ast.Ident) (*token.File, *ast.File) {
	namedType, ok := pass.TypesInfo.TypeOf(ident).(*types.Named)
	if !ok {
		return nil, nil
	}

	tokenFile := pass.Fset.File(namedType.Obj().Pos())

	for _, astFile := range pass.Files {
		if astFile.FileStart == token.Pos(tokenFile.Base()) {
			return tokenFile, astFile
		}
	}

	return tokenFile, nil
}

func isInPassPackage(pass *analysis.Pass, namedType *types.Named) bool {
	return namedType.Obj().Pkg() != nil && namedType.Obj().Pkg() == pass.Pkg
}

// TypeAwareMarkerCollectionForField collects the markers for a given field into a single markers.MarkerSet.
// If the field has a type that is not a basic type (i.e a custom type) then it will also gather any markers from
// the type and include them in the markers.MarkerSet that is returned.
// It will look through *ast.StarExpr to the underlying type.
// Markers on the type will always come before markers on the field in the list of markers for an identifier.
func TypeAwareMarkerCollectionForField(pass *analysis.Pass, markersAccess markershelper.Markers, field *ast.Field) markershelper.MarkerSet {
	markers := markersAccess.FieldMarkers(field)

	var underlyingType ast.Expr

	switch t := field.Type.(type) {
	case *ast.Ident:
		underlyingType = t
	case *ast.StarExpr:
		underlyingType = t.X
	default:
		return markers
	}

	ident, ok := underlyingType.(*ast.Ident)
	if !ok {
		return markers
	}

	if IsBasicType(pass, ident) {
		return markers
	}

	typeSpec, ok := LookupTypeSpec(pass, ident)
	if !ok {
		return markers
	}

	typeMarkers := markersAccess.TypeMarkers(typeSpec)
	typeMarkers.Insert(markers.UnsortedList()...)

	return typeMarkers
}

// IsArrayTypeOrAlias checks if the field type is an array type or an alias to an array type.
func IsArrayTypeOrAlias(pass *analysis.Pass, field *ast.Field) bool {
	if _, ok := field.Type.(*ast.ArrayType); ok {
		return true
	}

	if ident, ok := field.Type.(*ast.Ident); ok {
		typeOf := pass.TypesInfo.TypeOf(ident)
		if typeOf == nil {
			return false
		}

		return isArrayType(typeOf)
	}

	return false
}

// IsObjectList checks if the field represents a list of objects (not primitives).
func IsObjectList(pass *analysis.Pass, field *ast.Field) bool {
	if arrayType, ok := field.Type.(*ast.ArrayType); ok {
		return inspectType(pass, arrayType.Elt)
	}

	if ident, ok := field.Type.(*ast.Ident); ok {
		typeOf := pass.TypesInfo.TypeOf(ident)
		if typeOf == nil {
			return false
		}

		return isObjectListFromType(typeOf)
	}

	return false
}

// IsByteArray checks if the field type is a byte array or an alias to a byte array.
func IsByteArray(pass *analysis.Pass, field *ast.Field) bool {
	if arrayType, ok := field.Type.(*ast.ArrayType); ok {
		if ident, ok := arrayType.Elt.(*ast.Ident); ok && types.Identical(pass.TypesInfo.TypeOf(ident), types.Typ[types.Byte]) {
			return true
		}
	}

	if ident, ok := field.Type.(*ast.Ident); ok {
		typeOf := pass.TypesInfo.TypeOf(ident)
		if typeOf == nil {
			return false
		}

		switch typeOf := typeOf.(type) {
		case *types.Alias:
			if sliceType, ok := typeOf.Underlying().(*types.Slice); ok {
				return types.Identical(sliceType.Elem(), types.Typ[types.Byte])
			}
		case *types.Named:
			if sliceType, ok := typeOf.Underlying().(*types.Slice); ok {
				return types.Identical(sliceType.Elem(), types.Typ[types.Byte])
			}
		}
	}

	return false
}

func isArrayType(t types.Type) bool {
	if aliasType, ok := t.(*types.Alias); ok {
		return isArrayType(aliasType.Underlying())
	}

	if namedType, ok := t.(*types.Named); ok {
		return isArrayType(namedType.Underlying())
	}

	if _, ok := t.(*types.Slice); ok {
		return true
	}

	return false
}

func isObjectListFromType(t types.Type) bool {
	if aliasType, ok := t.(*types.Alias); ok {
		return isObjectListFromType(aliasType.Underlying())
	}

	if namedType, ok := t.(*types.Named); ok {
		return isObjectListFromType(namedType.Underlying())
	}

	if sliceType, ok := t.(*types.Slice); ok {
		return !isTypeBasic(sliceType.Elem())
	}

	return false
}

func inspectType(pass *analysis.Pass, expr ast.Expr) bool {
	switch elementType := expr.(type) {
	case *ast.Ident:
		return !isBasicOrAliasToBasic(pass, elementType)
	case *ast.StarExpr:
		return inspectType(pass, elementType.X)
	case *ast.ArrayType:
		return inspectType(pass, elementType.Elt)
	case *ast.SelectorExpr:
		return true
	}

	return false
}

func isBasicOrAliasToBasic(pass *analysis.Pass, ident *ast.Ident) bool {
	typeOf := pass.TypesInfo.TypeOf(ident)
	if typeOf == nil {
		return false
	}

	return isTypeBasic(typeOf)
}

func isTypeBasic(t types.Type) bool {
	// Direct basic type
	if _, ok := t.(*types.Basic); ok {
		return true
	}

	// Handle type aliases (type T = U)
	if aliasType, ok := t.(*types.Alias); ok {
		return isTypeBasic(aliasType.Underlying())
	}

	// Handle defined types (type T U)
	if namedType, ok := t.(*types.Named); ok {
		return isTypeBasic(namedType.Underlying())
	}

	return false
}

// GetMinProperties returns the value of the minimum properties marker.
// It returns a nil value when the marker is not present, and an error
// when the marker is present, but malformed.
func GetMinProperties(markerSet markershelper.MarkerSet) (*int, error) {
	minProperties, err := getMarkerNumericValueByName[int](markerSet, markers.KubebuilderMinPropertiesMarker)
	if err != nil && !errors.Is(err, errMarkerMissingValue) {
		return nil, fmt.Errorf("invalid format for minimum properties marker: %w", err)
	}

	return minProperties, nil
}

// IsKubernetesListType checks if a struct is a Kubernetes List type.
// A Kubernetes List type has:
// - Name ending with "List" (only checked if name is provided)
// - Exactly 3 fields: TypeMeta, ListMeta, and Items (slice type)
//
// The name parameter is optional and can be an empty string. When empty, only
// the structural pattern (3 fields: TypeMeta, ListMeta, Items) is checked without
// validating the type name suffix. This is useful for generic field inspection
// where the type name may not be readily available.
//
// Example:
//
//	type FooList struct {
//	    metav1.TypeMeta `json:",inline"`
//	    metav1.ListMeta `json:"metadata,omitempty"`
//	    Items           []Foo `json:"items"`
//	}
func IsKubernetesListType(sTyp *ast.StructType, name string) bool {
	if sTyp == nil || sTyp.Fields == nil || sTyp.Fields.List == nil {
		return false
	}

	// Check name suffix if name is provided
	if name != "" && !strings.HasSuffix(name, "List") {
		return false
	}

	// Must have exactly 3 fields
	if len(sTyp.Fields.List) != 3 {
		return false
	}

	return hasListFields(sTyp.Fields.List)
}

// hasListFields checks if the field list contains TypeMeta, ListMeta, and Items.
func hasListFields(fields []*ast.Field) bool {
	hasTypeMeta := false
	hasListMeta := false
	hasItems := false

	for _, field := range fields {
		typeName := getFieldTypeName(field)

		// Check for TypeMeta (embedded or named)
		if typeName == "TypeMeta" {
			hasTypeMeta = true
			continue
		}

		// Check for ListMeta (embedded or named)
		if typeName == "ListMeta" {
			hasListMeta = true
			continue
		}

		// Check for Items field (must be named "Items" and be a slice type)
		if len(field.Names) > 0 && field.Names[0].Name == "Items" {
			if _, ok := field.Type.(*ast.ArrayType); ok {
				hasItems = true
			}
		}
	}

	return hasTypeMeta && hasListMeta && hasItems
}

// getFieldTypeName returns the type name of a field, handling both embedded fields
// and named fields with simple or qualified identifiers (e.g., TypeMeta or metav1.TypeMeta).
func getFieldTypeName(field *ast.Field) string {
	switch t := field.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.Sel.Name
	}

	return ""
}
