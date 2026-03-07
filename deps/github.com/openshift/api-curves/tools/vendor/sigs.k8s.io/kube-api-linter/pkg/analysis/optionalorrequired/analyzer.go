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
package optionalorrequired

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	markershelper "sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/markers"
)

const (
	name = "optionalorrequired"
)

func init() {
	markershelper.DefaultRegistry().Register(
		markers.OptionalMarker,
		markers.RequiredMarker,
		markers.KubebuilderOptionalMarker,
		markers.KubebuilderRequiredMarker,
		markers.K8sOptionalMarker,
		markers.K8sRequiredMarker,
	)
}

type analyzer struct {
	primaryOptionalMarker   string
	secondaryOptionalMarker string

	primaryRequiredMarker   string
	secondaryRequiredMarker string
}

// newAnalyzer creates a new analyzer with the given configuration.
func newAnalyzer(cfg *OptionalOrRequiredConfig) *analysis.Analyzer {
	if cfg == nil {
		cfg = &OptionalOrRequiredConfig{}
	}

	defaultConfig(cfg)

	a := &analyzer{}

	switch cfg.PreferredOptionalMarker {
	case markers.OptionalMarker:
		a.primaryOptionalMarker = markers.OptionalMarker
		a.secondaryOptionalMarker = markers.KubebuilderOptionalMarker
	case markers.KubebuilderOptionalMarker:
		a.primaryOptionalMarker = markers.KubebuilderOptionalMarker
		a.secondaryOptionalMarker = markers.OptionalMarker
	}

	switch cfg.PreferredRequiredMarker {
	case markers.RequiredMarker:
		a.primaryRequiredMarker = markers.RequiredMarker
		a.secondaryRequiredMarker = markers.KubebuilderRequiredMarker
	case markers.KubebuilderRequiredMarker:
		a.primaryRequiredMarker = markers.KubebuilderRequiredMarker
		a.secondaryRequiredMarker = markers.RequiredMarker
	}

	return &analysis.Analyzer{
		Name:     name,
		Doc:      "Checks that all struct fields are marked either with the optional or required markers.",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer},
	}
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markershelper.Markers, qualifiedFieldName string) {
		a.checkField(pass, field, markersAccess.FieldMarkers(field), jsonTagInfo, qualifiedFieldName)
	})

	inspect.InspectTypeSpec(func(typeSpec *ast.TypeSpec, markersAccess markershelper.Markers) {
		a.checkTypeSpec(pass, typeSpec, markersAccess)
	})

	return nil, nil //nolint:nilnil
}

//nolint:cyclop
func (a *analyzer) checkField(pass *analysis.Pass, field *ast.Field, fieldMarkers markershelper.MarkerSet, fieldTagInfo extractjsontags.FieldTagInfo, qualifiedFieldName string) {
	if fieldTagInfo.Inline {
		// Inline fields would have no effect if they were marked as optional/required.
		return
	}

	prefix := "field %s"
	if len(field.Names) == 0 || field.Names[0] == nil {
		prefix = "embedded field %s"
	}

	prefix = fmt.Sprintf(prefix, qualifiedFieldName)

	hasPrimaryOptional := fieldMarkers.Has(a.primaryOptionalMarker)
	hasPrimaryRequired := fieldMarkers.Has(a.primaryRequiredMarker)

	hasSecondaryOptional := fieldMarkers.Has(a.secondaryOptionalMarker)
	hasSecondaryRequired := fieldMarkers.Has(a.secondaryRequiredMarker)

	hasEitherOptional := hasPrimaryOptional || hasSecondaryOptional
	hasEitherRequired := hasPrimaryRequired || hasSecondaryRequired

	hasBothOptional := hasPrimaryOptional && hasSecondaryOptional
	hasBothRequired := hasPrimaryRequired && hasSecondaryRequired

	a.checkK8sMarkers(pass, field, fieldMarkers, prefix, hasEitherOptional, hasEitherRequired)

	switch {
	case hasEitherOptional && hasEitherRequired:
		pass.Reportf(field.Pos(), "%s must not be marked as both optional and required", prefix)
	case hasSecondaryOptional:
		marker := fieldMarkers[a.secondaryOptionalMarker]
		if hasBothOptional {
			pass.Report(reportShouldRemoveSecondaryMarker(field, marker, a.primaryOptionalMarker, a.secondaryOptionalMarker, prefix))
		} else {
			pass.Report(reportShouldReplaceSecondaryMarker(field, marker, a.primaryOptionalMarker, a.secondaryOptionalMarker, prefix))
		}
	case hasSecondaryRequired:
		marker := fieldMarkers[a.secondaryRequiredMarker]
		if hasBothRequired {
			pass.Report(reportShouldRemoveSecondaryMarker(field, marker, a.primaryRequiredMarker, a.secondaryRequiredMarker, prefix))
		} else {
			pass.Report(reportShouldReplaceSecondaryMarker(field, marker, a.primaryRequiredMarker, a.secondaryRequiredMarker, prefix))
		}
	case hasPrimaryOptional || hasPrimaryRequired:
		// This is the correct state.
	default:
		pass.Reportf(field.Pos(), "%s must be marked as %s or %s", prefix, a.primaryOptionalMarker, a.primaryRequiredMarker)
	}
}

func (a *analyzer) checkK8sMarkers(pass *analysis.Pass, field *ast.Field, fieldMarkers markershelper.MarkerSet, prefix string, hasEitherOptional, hasEitherRequired bool) {
	hasK8sOptional := fieldMarkers.Has(markers.K8sOptionalMarker)
	hasK8sRequired := fieldMarkers.Has(markers.K8sRequiredMarker)

	if hasK8sOptional && hasK8sRequired {
		pass.Reportf(field.Pos(), "%s must not be marked as both %s and %s", prefix, markers.K8sOptionalMarker, markers.K8sRequiredMarker)
	}

	if hasK8sOptional && hasEitherRequired {
		pass.Reportf(field.Pos(), "%s must not be marked as both %s and %s", prefix, markers.K8sOptionalMarker, markers.RequiredMarker)
	}

	if hasK8sRequired && hasEitherOptional {
		pass.Reportf(field.Pos(), "%s must not be marked as both %s and %s", prefix, markers.OptionalMarker, markers.K8sRequiredMarker)
	}
}

func reportShouldReplaceSecondaryMarker(field *ast.Field, marker []markershelper.Marker, primaryMarker, secondaryMarker, prefix string) analysis.Diagnostic {
	textEdits := make([]analysis.TextEdit, len(marker))

	for i, m := range marker {
		if i == 0 {
			textEdits[i] = analysis.TextEdit{
				Pos:     m.Pos,
				End:     m.End,
				NewText: fmt.Appendf(nil, "// +%s", primaryMarker),
			}

			continue
		}

		textEdits[i] = analysis.TextEdit{
			Pos:     m.Pos,
			End:     m.End + 1, // Add 1 to position to include the new line
			NewText: nil,
		}
	}

	return analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: fmt.Sprintf("%s should use marker %s instead of %s", prefix, primaryMarker, secondaryMarker),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message:   fmt.Sprintf("should replace `%s` with `%s`", secondaryMarker, primaryMarker),
				TextEdits: textEdits,
			},
		},
	}
}

func reportShouldRemoveSecondaryMarker(field *ast.Field, marker []markershelper.Marker, primaryMarker, secondaryMarker, prefix string) analysis.Diagnostic {
	textEdits := make([]analysis.TextEdit, len(marker))

	for i, m := range marker {
		textEdits[i] = analysis.TextEdit{
			Pos:     m.Pos,
			End:     m.End + 1, // Add 1 to position to include the new line
			NewText: nil,
		}
	}

	return analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: fmt.Sprintf("%s should use only the marker %s, %s is not required", prefix, primaryMarker, secondaryMarker),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message:   fmt.Sprintf("should remove `// +%s`", secondaryMarker),
				TextEdits: textEdits,
			},
		},
	}
}

func (a *analyzer) checkTypeSpec(pass *analysis.Pass, typeSpec *ast.TypeSpec, markersAccess markershelper.Markers) {
	name := typeSpec.Name.Name
	set := markersAccess.TypeMarkers(typeSpec)

	for _, marker := range set.UnsortedList() {
		switch marker.Identifier {
		case a.primaryOptionalMarker, a.secondaryOptionalMarker, a.primaryRequiredMarker, a.secondaryRequiredMarker, markers.K8sOptionalMarker, markers.K8sRequiredMarker:
			pass.Report(analysis.Diagnostic{
				Pos:     typeSpec.Pos(),
				Message: fmt.Sprintf("type %s should not be marked as %s", name, marker.String()),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: fmt.Sprintf("should remove `// +%s`", marker.String()),
						TextEdits: []analysis.TextEdit{
							{
								Pos:     marker.Pos,
								End:     marker.End,
								NewText: nil,
							},
						},
					},
				},
			})
		}
	}
}

func defaultConfig(cfg *OptionalOrRequiredConfig) {
	if cfg.PreferredOptionalMarker == "" {
		cfg.PreferredOptionalMarker = markers.OptionalMarker
	}

	if cfg.PreferredRequiredMarker == "" {
		cfg.PreferredRequiredMarker = markers.RequiredMarker
	}
}
