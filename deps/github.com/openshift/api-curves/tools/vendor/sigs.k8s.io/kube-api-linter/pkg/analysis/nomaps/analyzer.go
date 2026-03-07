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
package nomaps

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
)

const (
	name = "nomaps"
)

type analyzer struct {
	policy NoMapsPolicy
}

// newAnalyzer creates a new analyzer.
func newAnalyzer(cfg *NoMapsConfig) *analysis.Analyzer {
	if cfg == nil {
		cfg = &NoMapsConfig{}
	}

	defaultConfig(cfg)

	a := &analyzer{
		policy: cfg.Policy,
	}

	return &analysis.Analyzer{
		Name:     name,
		Doc:      "Checks for usage of map types. Maps are discouraged apart from `map[string]string` which is used for labels and annotations. Use a list of named objects instead.",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer},
	}
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	typeChecker := utils.NewTypeChecker(isMap, a.checkMap)

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, _ markers.Markers, _ string) {
		typeChecker.CheckNode(pass, field)
	})

	inspect.InspectTypeSpec(func(typeSpec *ast.TypeSpec, _ markers.Markers) {
		typeChecker.CheckNode(pass, typeSpec)
	})

	return nil, nil //nolint:nilnil
}

func isMap(pass *analysis.Pass, expr ast.Expr) bool {
	_, ok := expr.(*ast.MapType)

	return ok
}

func (a *analyzer) checkMap(pass *analysis.Pass, expr ast.Expr, node ast.Node, prefix string) {
	mapType, ok := expr.(*ast.MapType)
	if !ok {
		return
	}

	switch a.policy {
	case NoMapsEnforce:
		report(pass, node.Pos(), prefix)
	case NoMapsAllowStringToStringMaps:
		if !isStringToStringMap(pass, mapType) {
			report(pass, node.Pos(), prefix)
		}
	case NoMapsIgnore:
		if !isBasicMap(pass, mapType) {
			report(pass, node.Pos(), prefix)
		}
	}
}

func isStringToStringMap(pass *analysis.Pass, mapType *ast.MapType) bool {
	return utils.IsStringType(pass, mapType.Key) && utils.IsStringType(pass, mapType.Value)
}

func isBasicMap(pass *analysis.Pass, mapType *ast.MapType) bool {
	return utils.IsBasicType(pass, mapType.Key) && utils.IsBasicType(pass, mapType.Value)
}

func report(pass *analysis.Pass, pos token.Pos, fieldName string) {
	pass.Report(analysis.Diagnostic{
		Pos:     pos,
		Message: fmt.Sprintf("%s should not use a map type, use a list type with a unique name/identifier instead", fieldName),
	})
}

func defaultConfig(cfg *NoMapsConfig) {
	if cfg.Policy == "" {
		cfg.Policy = NoMapsAllowStringToStringMaps
	}
}
