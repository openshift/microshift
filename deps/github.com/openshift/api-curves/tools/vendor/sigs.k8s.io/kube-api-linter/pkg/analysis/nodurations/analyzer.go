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

package nodurations

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
)

const name = "nodurations"

// Analyzer is the analyzer for the nodurations package.
// It checks that no struct field is of a type either time.Duration or metav1.Duration.
var Analyzer = &analysis.Analyzer{
	Name:     name,
	Doc:      "Duration types should not be used, to avoid the need for clients to implement go duration parsing. Instead, use integer based fields with the unit in the field name.",
	Run:      run,
	Requires: []*analysis.Analyzer{inspector.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	typeChecker := utils.NewTypeChecker(isDurationType, checkDuration)

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, _ markers.Markers, _ string) {
		typeChecker.CheckNode(pass, field)
	})

	inspect.InspectTypeSpec(func(typeSpec *ast.TypeSpec, markersAccess markers.Markers) {
		typeChecker.CheckNode(pass, typeSpec)
	})

	return nil, nil //nolint:nilnil
}

func isDurationType(pass *analysis.Pass, expr ast.Expr) bool {
	typ, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkg, ok := typ.X.(*ast.Ident)
	if !ok {
		return false
	}

	if typ.X == nil || (pkg.Name != "time" && pkg.Name != "metav1") {
		return false
	}

	if typ.Sel == nil || typ.Sel.Name != "Duration" {
		return false
	}

	return true
}

func checkDuration(pass *analysis.Pass, expr ast.Expr, node ast.Node, prefix string) {
	pass.Reportf(node.Pos(), "%s should not use a Duration. Use an integer type with units in the name to avoid the need for clients to implement Go style duration parsing.", prefix)
}
