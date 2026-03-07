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
package nofloats

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
)

const name = "nofloats"

// Analyzer is the analyzer for the nofloats package.
// It checks that no struct fields are `float`.
var Analyzer = &analysis.Analyzer{
	Name:     name,
	Doc:      "Float values cannot be reliably round-tripped without changing and have varying precisions and representations across languages and architectures.",
	Run:      run,
	Requires: []*analysis.Analyzer{inspector.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	typeChecker := utils.NewTypeChecker(utils.IsBasicType, checkFloat)

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, _ markers.Markers, _ string) {
		typeChecker.CheckNode(pass, field)
	})

	inspect.InspectTypeSpec(func(typeSpec *ast.TypeSpec, markersAccess markers.Markers) {
		typeChecker.CheckNode(pass, typeSpec)
	})

	return nil, nil //nolint:nilnil
}

func checkFloat(pass *analysis.Pass, expr ast.Expr, node ast.Node, prefix string) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return
	}

	if ident.Name == "float32" || ident.Name == "float64" {
		pass.Reportf(node.Pos(), "%s should not use a float value because they cannot be reliably round-tripped.", prefix)
	}
}
