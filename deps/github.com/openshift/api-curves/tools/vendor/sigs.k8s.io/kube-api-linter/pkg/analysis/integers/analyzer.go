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
package integers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
)

const name = "integers"

// Analyzer is the analyzer for the integers package.
// It checks that no struct fields or type aliases are `int`, or unsigned integers.
var Analyzer = &analysis.Analyzer{
	Name:     name,
	Doc:      "All integers should be explicit about their size, int32 and int64 should be used over plain int. Unsigned ints are not allowed.",
	Run:      run,
	Requires: []*analysis.Analyzer{inspector.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	typeChecker := utils.NewTypeChecker(utils.IsBasicType, checkIntegers)

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, _ markers.Markers, _ string) {
		typeChecker.CheckNode(pass, field)
	})

	inspect.InspectTypeSpec(func(typeSpec *ast.TypeSpec, markersAccess markers.Markers) {
		typeChecker.CheckNode(pass, typeSpec)
	})

	return nil, nil //nolint:nilnil
}

// checkIntegers looks for known type of integers that do not match the allowed `int32` or `int64` requirements.
func checkIntegers(pass *analysis.Pass, expr ast.Expr, node ast.Node, prefix string) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return
	}

	switch ident.Name {
	case "int32", "int64":
		// Valid cases
	case "int", "int8", "int16":
		pass.Reportf(node.Pos(), "%s should not use an int, int8 or int16. Use int32 or int64 depending on bounding requirements", prefix)
	case "uint", "uint8", "uint16", "uint32", "uint64":
		pass.Reportf(node.Pos(), "%s should not use unsigned integers, use only int32 or int64 and apply validation to ensure the value is positive", prefix)
	}
}
