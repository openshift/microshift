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
package defaults

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	markershelper "sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
	"sigs.k8s.io/kube-api-linter/pkg/markers"
)

const (
	name = "defaults"
)

func init() {
	markershelper.DefaultRegistry().Register(
		markers.DefaultMarker,
		markers.KubebuilderDefaultMarker,
		markers.K8sDefaultMarker,
		markers.OptionalMarker,
		markers.KubebuilderOptionalMarker,
		markers.K8sOptionalMarker,
		markers.RequiredMarker,
		markers.KubebuilderRequiredMarker,
		markers.K8sRequiredMarker,
	)
}

type analyzer struct {
	preferredDefaultMarker string
	secondaryDefaultMarker string
	omitEmptyPolicy        OmitEmptyPolicy
	omitZeroPolicy         OmitZeroPolicy
}

// newAnalyzer creates a new analyzer with the given configuration.
func newAnalyzer(cfg *DefaultsConfig) *analysis.Analyzer {
	if cfg == nil {
		cfg = &DefaultsConfig{}
	}

	defaultConfig(cfg)

	a := &analyzer{
		omitEmptyPolicy: cfg.OmitEmpty.Policy,
		omitZeroPolicy:  cfg.OmitZero.Policy,
	}

	switch cfg.PreferredDefaultMarker {
	case markers.DefaultMarker:
		a.preferredDefaultMarker = markers.DefaultMarker
		a.secondaryDefaultMarker = markers.KubebuilderDefaultMarker
	case markers.KubebuilderDefaultMarker:
		a.preferredDefaultMarker = markers.KubebuilderDefaultMarker
		a.secondaryDefaultMarker = markers.DefaultMarker
	}

	return &analysis.Analyzer{
		Name: name,
		Doc: `Checks that fields with default markers are configured correctly.
Fields with default markers (+default, +kubebuilder:default, or +k8s:default) should also be marked as optional and not required.
Additionally, fields with default markers should have "omitempty" or "omitzero" in their json tags to ensure that the default values are applied correctly during serialization and deserialization.
`,
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer},
	}
}

func defaultConfig(cfg *DefaultsConfig) {
	if cfg.PreferredDefaultMarker == "" {
		cfg.PreferredDefaultMarker = markers.DefaultMarker
	}

	if cfg.OmitEmpty.Policy == "" {
		cfg.OmitEmpty.Policy = OmitEmptyPolicySuggestFix
	}

	if cfg.OmitZero.Policy == "" {
		cfg.OmitZero.Policy = OmitZeroPolicySuggestFix
	}
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markershelper.Markers, qualifiedFieldName string) {
		a.checkField(pass, field, jsonTagInfo, markersAccess, qualifiedFieldName)
	})

	return nil, nil //nolint:nilnil
}

func (a *analyzer) checkField(pass *analysis.Pass, field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markershelper.Markers, qualifiedFieldName string) {
	if field == nil || len(field.Names) == 0 {
		return
	}

	fieldMarkers := markersAccess.FieldMarkers(field)

	// Check for any default marker (+default, +kubebuilder:default, or +k8s:default)
	hasPreferredDefault := fieldMarkers.Has(a.preferredDefaultMarker)
	hasSecondaryDefault := fieldMarkers.Has(a.secondaryDefaultMarker)
	hasK8sDefault := fieldMarkers.Has(markers.K8sDefaultMarker)

	hasAnyDefault := hasPreferredDefault || hasSecondaryDefault || hasK8sDefault

	if !hasAnyDefault {
		return
	}

	// Check +k8s:default marker (for declarative validation, separate from preferred/secondary)
	// If +k8s:default is present but neither +default nor +kubebuilder:default is present, suggest adding the preferred one
	a.checkK8sDefault(pass, field, fieldMarkers, qualifiedFieldName, hasPreferredDefault || hasSecondaryDefault)

	// Check secondary marker usage
	// If both preferred and secondary exist, suggest removing secondary
	// If only secondary exists, suggest replacing with preferred
	if hasSecondaryDefault {
		hasBothDefaults := hasPreferredDefault && hasSecondaryDefault
		a.checkSecondaryDefault(pass, field, fieldMarkers, qualifiedFieldName, hasBothDefaults)
	}

	a.checkDefaultNotRequired(pass, field, markersAccess, qualifiedFieldName)

	a.checkDefaultOptional(pass, field, markersAccess, qualifiedFieldName)

	a.checkDefaultOmitEmptyOrOmitZero(pass, field, jsonTagInfo, qualifiedFieldName)
}

// checkK8sDefault checks for +k8s:default marker usage.
// +k8s:default is for declarative validation and is separate from preferred/secondary default markers.
// If the field has +k8s:default but doesn't have +default or +kubebuilder:default, we suggest adding the preferred one.
func (a *analyzer) checkK8sDefault(pass *analysis.Pass, field *ast.Field, fieldMarkers markershelper.MarkerSet, qualifiedFieldName string, hasOtherDefault bool) {
	if !fieldMarkers.Has(markers.K8sDefaultMarker) {
		return
	}

	// If the field already has +default or +kubebuilder:default, +k8s:default is acceptable alongside them
	// (e.g., in K/K where both are needed during transition period)
	if hasOtherDefault {
		return
	}

	// If only +k8s:default is present, suggest adding the preferred default marker
	k8sDefaultMarkers := fieldMarkers.Get(markers.K8sDefaultMarker)
	for _, marker := range k8sDefaultMarkers {
		payloadValue := marker.Payload.Value
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("field %s has +%s but should also have +%s marker", qualifiedFieldName, markers.K8sDefaultMarker, a.preferredDefaultMarker),
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: fmt.Sprintf("add +%s=%s", a.preferredDefaultMarker, payloadValue),
					TextEdits: []analysis.TextEdit{
						{
							Pos:     marker.Pos,
							End:     marker.Pos,
							NewText: fmt.Appendf(nil, "// +%s=%s\n\t", a.preferredDefaultMarker, payloadValue),
						},
					},
				},
			},
		})
	}
}

func (a *analyzer) checkSecondaryDefault(pass *analysis.Pass, field *ast.Field, fieldMarkers markershelper.MarkerSet, qualifiedFieldName string, hasBothDefaults bool) {
	secondaryDefaultMarkers := fieldMarkers.Get(a.secondaryDefaultMarker)

	if hasBothDefaults {
		// Both preferred and secondary markers exist - suggest removing secondary
		pass.Report(reportShouldRemoveSecondaryMarker(field, secondaryDefaultMarkers, a.preferredDefaultMarker, a.secondaryDefaultMarker, qualifiedFieldName))
		return
	}
	// Only secondary marker exists - suggest replacing with preferred
	pass.Report(reportShouldReplaceSecondaryMarker(field, secondaryDefaultMarkers, a.preferredDefaultMarker, a.secondaryDefaultMarker, qualifiedFieldName))
}

func reportShouldReplaceSecondaryMarker(field *ast.Field, markers []markershelper.Marker, preferredMarker, secondaryMarker, qualifiedFieldName string) analysis.Diagnostic {
	textEdits := make([]analysis.TextEdit, len(markers))

	for i, marker := range markers {
		if i == 0 {
			textEdits[i] = analysis.TextEdit{
				Pos:     marker.Pos,
				End:     marker.End,
				NewText: fmt.Appendf(nil, "// +%s=%s", preferredMarker, marker.Payload.Value),
			}

			continue
		}

		textEdits[i] = analysis.TextEdit{
			Pos:     marker.Pos,
			End:     marker.End + 1, // Add 1 to position to include the new line
			NewText: nil,
		}
	}

	return analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: fmt.Sprintf("field %s should use +%s marker instead of +%s", qualifiedFieldName, preferredMarker, secondaryMarker),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message:   fmt.Sprintf("replace +%s with +%s", secondaryMarker, preferredMarker),
				TextEdits: textEdits,
			},
		},
	}
}

func reportShouldRemoveSecondaryMarker(field *ast.Field, markers []markershelper.Marker, preferredMarker, secondaryMarker, qualifiedFieldName string) analysis.Diagnostic {
	textEdits := make([]analysis.TextEdit, len(markers))

	for i, marker := range markers {
		textEdits[i] = analysis.TextEdit{
			Pos:     marker.Pos,
			End:     marker.End + 1, // Add 1 to position to include the new line
			NewText: nil,
		}
	}

	return analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: fmt.Sprintf("field %s should use only the marker +%s, +%s is not required", qualifiedFieldName, preferredMarker, secondaryMarker),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message:   fmt.Sprintf("remove +%s", secondaryMarker),
				TextEdits: textEdits,
			},
		},
	}
}

func (a *analyzer) checkDefaultNotRequired(pass *analysis.Pass, field *ast.Field, markersAccess markershelper.Markers, qualifiedFieldName string) {
	if utils.IsFieldRequired(field, markersAccess) {
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("field %s has a default value but is marked as required, which is contradictory", qualifiedFieldName),
		})
	}
}

func (a *analyzer) checkDefaultOptional(pass *analysis.Pass, field *ast.Field, markersAccess markershelper.Markers, qualifiedFieldName string) {
	// If the field is required, we've already reported that issue in checkDefaultNotRequired
	if utils.IsFieldRequired(field, markersAccess) {
		return
	}

	if !utils.IsFieldOptional(field, markersAccess) {
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("field %s has a default value but is not marked as optional", qualifiedFieldName),
		})
	}
}

func (a *analyzer) checkDefaultOmitEmptyOrOmitZero(pass *analysis.Pass, field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, qualifiedFieldName string) {
	if jsonTagInfo.Inline || jsonTagInfo.Ignored {
		return
	}

	hasOmitEmpty := jsonTagInfo.OmitEmpty
	hasOmitZero := jsonTagInfo.OmitZero

	// Check if the field is a pointer type - pointers don't need omitzero because nil is their zero value
	isPointer, _ := utils.IsStarExpr(field.Type)
	isStruct := !isPointer && utils.IsStructType(pass, field.Type)

	// For struct types (but not pointers), we prefer omitzero over omitempty.
	// When omitzero is present, omitempty is not needed (modernize linter would complain).
	if isStruct && a.omitZeroPolicy != OmitZeroPolicyForbid {
		if !hasOmitZero {
			a.reportMissingOmitZero(pass, field, jsonTagInfo, qualifiedFieldName, hasOmitEmpty)
		}

		return
	}

	// Check omitempty for non-struct types (only if policy is not Ignore)
	if a.omitEmptyPolicy != OmitEmptyPolicyIgnore && !hasOmitEmpty {
		a.reportMissingOmitEmpty(pass, field, jsonTagInfo, qualifiedFieldName)
	}
}

func (a *analyzer) reportMissingOmitEmpty(pass *analysis.Pass, field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, qualifiedFieldName string) {
	suggestedTag := fmt.Sprintf("%s,omitempty", jsonTagInfo.RawValue)
	message := fmt.Sprintf("add omitempty to the json tag of field %s", qualifiedFieldName)

	switch a.omitEmptyPolicy {
	case OmitEmptyPolicySuggestFix:
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("field %s has a default value but does not have omitempty in its json tag", qualifiedFieldName),
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: message,
					TextEdits: []analysis.TextEdit{
						{
							Pos:     jsonTagInfo.Pos,
							End:     jsonTagInfo.End,
							NewText: []byte(suggestedTag),
						},
					},
				},
			},
		})
	case OmitEmptyPolicyWarn:
		pass.Reportf(field.Pos(), "field %s has a default value but does not have omitempty in its json tag", qualifiedFieldName)
	case OmitEmptyPolicyIgnore:
		// Unreachable: this function is only called when the policy is not Ignore.
		return
	}
}

func (a *analyzer) reportMissingOmitZero(pass *analysis.Pass, field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, qualifiedFieldName string, hasOmitEmpty bool) {
	// For struct types, we prefer omitzero over omitempty.
	// If the field has omitempty, we replace it with omitzero.
	// If the field doesn't have omitempty, we just add omitzero.
	// We never add both omitempty and omitzero together (modernize linter would complain).
	var suggestedTag string

	var message string

	if hasOmitEmpty {
		// Replace omitempty with omitzero
		suggestedTag = replaceOmitEmptyWithOmitZero(jsonTagInfo.RawValue)
		message = fmt.Sprintf("replace omitempty with omitzero in the json tag of field %s", qualifiedFieldName)
	} else {
		// Just add omitzero
		suggestedTag = fmt.Sprintf("%s,omitzero", jsonTagInfo.RawValue)
		message = fmt.Sprintf("add omitzero to the json tag of field %s", qualifiedFieldName)
	}

	switch a.omitZeroPolicy {
	case OmitZeroPolicySuggestFix:
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("field %s has a default value but does not have omitzero in its json tag", qualifiedFieldName),
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: message,
					TextEdits: []analysis.TextEdit{
						{
							Pos:     jsonTagInfo.Pos,
							End:     jsonTagInfo.End,
							NewText: []byte(suggestedTag),
						},
					},
				},
			},
		})
	case OmitZeroPolicyWarn:
		pass.Reportf(field.Pos(), "field %s has a default value but does not have omitzero in its json tag", qualifiedFieldName)
	case OmitZeroPolicyForbid:
		// Unreachable: this function is only called when the policy is not Forbid.
		return
	}
}

// replaceOmitEmptyWithOmitZero replaces omitempty with omitzero in the json tag value.
func replaceOmitEmptyWithOmitZero(rawValue string) string {
	// rawValue is like "fieldName,omitempty" or "fieldName,omitempty,inline"
	// We need to replace "omitempty" with "omitzero"
	parts := strings.Split(rawValue, ",")
	for i, part := range parts {
		if part == "omitempty" {
			parts[i] = "omitzero"
		}
	}

	return strings.Join(parts, ",")
}
