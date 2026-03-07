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
package statussubresource

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	markershelper "sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
	"sigs.k8s.io/kube-api-linter/pkg/markers"
)

const (
	name = "statussubresource"

	statusJSONTag = "status"
)

// Analyzer is a analyzer for the statussubresource package.
var Analyzer = &analysis.Analyzer{
	Name:     name,
	Doc:      "Checks that a type marked with kubebuilder:object:root:=true and containing a status field is marked with kubebuilder:subresource:status",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer, markershelper.Analyzer, extractjsontags.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	markersAccess, ok := pass.ResultOf[markershelper.Analyzer].(markershelper.Markers)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetMarkers
	}

	jsonTags, ok := pass.ResultOf[extractjsontags.Analyzer].(extractjsontags.StructFieldTags)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetJSONTags
	}

	// Filter to type specs so we can get the names of types
	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return
		}

		// we only care about struct types
		sTyp, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return
		}

		// no identifier on the type
		if typeSpec.Name == nil {
			return
		}

		structMarkers := markersAccess.StructMarkers(sTyp)
		checkStruct(pass, sTyp, typeSpec.Name.Name, structMarkers, jsonTags)
	})

	return nil, nil //nolint:nilnil
}

func checkStruct(pass *analysis.Pass, sTyp *ast.StructType, name string, structMarkers markershelper.MarkerSet, jsonTags extractjsontags.StructFieldTags) {
	if sTyp == nil {
		return
	}

	if !structMarkers.HasWithValue(formatKubeBuilderMarkerWithValue(markers.KubebuilderRootMarker, "true")) {
		return
	}

	// Skip Kubernetes List types as they follow a different pattern
	// and don't use the status subresource
	if utils.IsKubernetesListType(sTyp, name) {
		return
	}

	hasStatusSubresourceMarker := structMarkers.Has(markers.KubebuilderStatusSubresourceMarker)
	hasStatusField := hasStatusField(sTyp, jsonTags)

	// Both present or both absent is acceptable
	if hasStatusSubresourceMarker == hasStatusField {
		return
	}

	// Marker present but no status field
	if hasStatusSubresourceMarker {
		pass.Reportf(sTyp.Pos(), "root object type %q is marked to enable the status subresource with marker %q but has no status field", name, markers.KubebuilderStatusSubresourceMarker)
		return
	}

	// Status field present but no marker - suggest autofix
	pass.Report(analysis.Diagnostic{
		Pos:     sTyp.Pos(),
		Message: fmt.Sprintf("root object type %q has a status field but does not have the marker %q to enable the status subresource", name, markers.KubebuilderStatusSubresourceMarker),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "should add the kubebuilder:subresource:status marker",
				TextEdits: []analysis.TextEdit{
					{
						// sTyp.Pos() is the beginning of the 'struct' keyword. Subtract
						// the length of the struct name + 7 (2 for spaces surrounding type name, 4 for the 'type' keyword,
						// and 1 for the newline) to position at the end of the line above the struct
						// definition.
						Pos: sTyp.Pos() - token.Pos(len(name)+7),
						// prefix with a newline to ensure we aren't appending to a previous comment
						NewText: []byte("\n// +kubebuilder:subresource:status"),
					},
				},
			},
		},
	})
}

func hasStatusField(sTyp *ast.StructType, jsonTags extractjsontags.StructFieldTags) bool {
	if sTyp == nil || sTyp.Fields == nil || sTyp.Fields.List == nil {
		return false
	}

	for _, field := range sTyp.Fields.List {
		info := jsonTags.FieldTags(field)
		if info.Name == statusJSONTag {
			return true
		}
	}

	return false
}

func formatKubeBuilderMarkerWithValue(marker, value string) string {
	return fmt.Sprintf("%s:=%s", marker, value)
}
