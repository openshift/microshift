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
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// TypeChecker is an interface for checking types in an AST.
// It is used to check the underlying, built-in types of fields within structs, and raw type declarations.
// It is up to the implementer to provide a function that will be called when a built-in type is found,
// and to provide the logic for providing analysis about this type.
type TypeChecker interface {
	CheckNode(pass *analysis.Pass, node ast.Node)
}

// NewTypeChecker returns a new TypeChecker with the provided checkFunc.
func NewTypeChecker(isTypeFunc func(pass *analysis.Pass, ident ast.Expr) bool, checkFunc func(pass *analysis.Pass, expr ast.Expr, node ast.Node, prefix string)) TypeChecker {
	return &typeChecker{
		isTypeFunc: isTypeFunc,
		checkFunc:  checkFunc,
	}
}

type typeChecker struct {
	isTypeFunc func(pass *analysis.Pass, expr ast.Expr) bool
	checkFunc  func(pass *analysis.Pass, expr ast.Expr, node ast.Node, prefix string)
}

// CheckNode checks the provided node for built-in types.
// It will iterate through fields within structs, and raw type declarations.
// Calling the checkFunc when a built-in type is found.
func (t *typeChecker) CheckNode(pass *analysis.Pass, node ast.Node) {
	switch n := node.(type) {
	case *ast.StructType:
		if n.Fields == nil {
			return
		}

		for _, field := range n.Fields.List {
			t.checkField(pass, field)
		}
	case *ast.Field:
		t.checkField(pass, n)
	case *ast.TypeSpec:
		t.checkTypeSpec(pass, n, node, "type")
	}
}

func (t *typeChecker) checkField(pass *analysis.Pass, field *ast.Field) {
	fieldName := GetQualifiedFieldName(pass, field)
	if fieldName == "" {
		return
	}

	prefix := fmt.Sprintf("field %s", fieldName)

	t.checkTypeExpr(pass, field.Type, field, prefix)
}

func (t *typeChecker) checkTypeSpec(pass *analysis.Pass, tSpec *ast.TypeSpec, node ast.Node, prefix string) {
	if tSpec.Name == nil {
		return
	}

	typeName := tSpec.Name.Name
	prefix = fmt.Sprintf("%s %s", prefix, typeName)

	t.checkTypeExpr(pass, tSpec.Type, node, prefix)
}

func (t *typeChecker) checkTypeExpr(pass *analysis.Pass, typeExpr ast.Expr, node ast.Node, prefix string) {
	if t.isTypeFunc(pass, typeExpr) {
		t.checkFunc(pass, typeExpr, node, prefix)
		return
	}

	switch typ := typeExpr.(type) {
	case *ast.Ident:
		t.checkIdent(pass, typ, node, prefix)
	case *ast.StarExpr:
		t.checkTypeExpr(pass, typ.X, node, fmt.Sprintf("%s pointer", prefix))
	case *ast.ArrayType:
		t.checkTypeExpr(pass, typ.Elt, node, fmt.Sprintf("%s array element", prefix))
	case *ast.MapType:
		t.checkTypeExpr(pass, typ.Key, node, fmt.Sprintf("%s map key", prefix))
		t.checkTypeExpr(pass, typ.Value, node, fmt.Sprintf("%s map value", prefix))
	case *ast.IndexExpr:
		t.checkTypeExpr(pass, typ.X, node, prefix)
	}
}

// checkIdent calls the checkFunc with the ident, when we have hit a built-in type.
// If the ident is not a built in, we look at the underlying type until we hit a built-in type.
func (t *typeChecker) checkIdent(pass *analysis.Pass, ident *ast.Ident, node ast.Node, prefix string) {
	tSpec, ok := LookupTypeSpec(pass, ident)
	if !ok {
		return
	}

	// The field is using a type alias, check if the alias is an int.
	t.checkTypeSpec(pass, tSpec, node, fmt.Sprintf("%s type", prefix))
}
