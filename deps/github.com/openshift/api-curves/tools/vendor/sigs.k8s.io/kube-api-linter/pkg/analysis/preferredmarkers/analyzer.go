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
package preferredmarkers

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
)

const name = "preferredmarkers"

type analyzer struct {
	// equivalentToPreferred maps equivalent marker identifiers to their preferred identifiers
	equivalentToPreferred map[string]string
}

// newAnalyzer creates a new analysis.Analyzer for the preferredmarkers
// linter based on the provided Config.
func newAnalyzer(cfg *Config) *analysis.Analyzer {
	a := &analyzer{
		equivalentToPreferred: make(map[string]string),
	}

	// Build the mapping from equivalent identifiers to preferred identifiers
	for _, marker := range cfg.Markers {
		for _, equivalent := range marker.EquivalentIdentifiers {
			a.equivalentToPreferred[equivalent.Identifier] = marker.PreferredIdentifier
		}
	}

	analyzer := &analysis.Analyzer{
		Name:     name,
		Doc:      "Check that preferred markers are used instead of equivalent markers.",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer},
	}

	// Register all equivalent identifiers so they can be parsed.
	// Note: The marker registry is thread-safe and idempotent, so it's safe
	// to register the same marker multiple times or from concurrent goroutines.
	for equivalent := range a.equivalentToPreferred {
		markers.DefaultRegistry().Register(equivalent)
	}

	return analyzer
}

// run is the main analysis function that inspects all types and fields in the package.
func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string) {
		checkField(pass, field, markersAccess, a.equivalentToPreferred, qualifiedFieldName)
	})

	inspect.InspectTypeSpec(func(typeSpec *ast.TypeSpec, markersAccess markers.Markers) {
		checkType(pass, typeSpec, markersAccess, a.equivalentToPreferred)
	})

	return nil, nil //nolint:nilnil
}

// checkField validates a single struct field for marker usage.
// Only checks markers directly on the field, not inherited from type aliases,
// since inherited markers are already reported at the type level.
func checkField(pass *analysis.Pass, field *ast.Field, markersAccess markers.Markers, equivalentToPreferred map[string]string, qualifiedFieldName string) {
	if field == nil || len(field.Names) == 0 {
		return
	}

	markerSet := markersAccess.FieldMarkers(field)
	check(markerSet, equivalentToPreferred, func(marks []markers.Marker, preferredIdentifier string, preferredExists bool) {
		reportMarkers(pass, marks, preferredIdentifier, qualifiedFieldName, field.Pos(), "field", preferredExists)
	})
}

// checkType validates a single type definition for marker usage.
func checkType(pass *analysis.Pass, typeSpec *ast.TypeSpec, markersAccess markers.Markers, equivalentToPreferred map[string]string) {
	if typeSpec == nil {
		return
	}

	markerSet := markersAccess.TypeMarkers(typeSpec)
	check(markerSet, equivalentToPreferred, func(marks []markers.Marker, preferredIdentifier string, preferredExists bool) {
		reportMarkers(pass, marks, preferredIdentifier, typeSpec.Name.Name, typeSpec.Pos(), "type", preferredExists)
	})
}

// check examines a set of markers for equivalent identifiers that should be replaced.
func check(markerSet markers.MarkerSet, equivalentToPreferred map[string]string, reportFunc func(markers []markers.Marker, preferredIdentifier string, preferredExists bool)) {
	// Group markers by their preferred identifier to handle duplicates correctly
	preferredToMarkers := make(map[string][]markers.Marker)

	for equivalentIdentifier, preferredIdentifier := range equivalentToPreferred {
		marks := markerSet.Get(equivalentIdentifier)
		if len(marks) > 0 {
			preferredToMarkers[preferredIdentifier] = append(preferredToMarkers[preferredIdentifier], marks...)
		}
	}

	// Sort preferred identifiers for deterministic reporting
	preferredIdentifiers := make([]string, 0, len(preferredToMarkers))
	for preferredIdentifier := range preferredToMarkers {
		preferredIdentifiers = append(preferredIdentifiers, preferredIdentifier)
	}

	sort.Strings(preferredIdentifiers)

	// Report each group of markers
	for _, preferredIdentifier := range preferredIdentifiers {
		marks := preferredToMarkers[preferredIdentifier]
		// Check if the preferred marker already exists
		preferredExists := len(markerSet.Get(preferredIdentifier)) > 0
		reportFunc(marks, preferredIdentifier, preferredExists)
	}
}

// formatMarkerList formats a list of markers as a sorted, quoted, comma-separated string.
// For example, [marker1, marker2] becomes `"marker1", "marker2"`.
func formatMarkerList(marks []markers.Marker) string {
	names := make([]string, len(marks))
	for i, m := range marks {
		names[i] = fmt.Sprintf("%q", m.Identifier)
	}

	sort.Strings(names)

	return strings.Join(names, ", ")
}

// buildTextEdits generates the text edits to fix equivalent markers.
// If preferredExists is true, all markers are deleted. Otherwise, the first
// marker is replaced with the preferred identifier and the rest are deleted.
func buildTextEdits(marks []markers.Marker, preferredIdentifier string, preferredExists bool) []analysis.TextEdit {
	// Sort markers by position to ensure deterministic text edits
	sort.Slice(marks, func(i, j int) bool {
		return marks[i].Pos < marks[j].Pos
	})

	edits := make([]analysis.TextEdit, 0, len(marks))

	// If the preferred marker doesn't exist, replace the first equivalent marker
	if !preferredExists {
		edits = append(edits, analysis.TextEdit{
			Pos:     marks[0].Pos,
			End:     marks[0].End,
			NewText: []byte(buildReplacementMarker(marks[0], preferredIdentifier)),
		})
		marks = marks[1:] // Process remaining markers for deletion
	}

	// Delete remaining markers to avoid duplicates
	// Note: We add 1 to the end position to include the newline character,
	// which removes the entire line and prevents blank lines in the output.
	// This works correctly for most cases. At end of file without a trailing
	// newline, the go/analysis framework handles the extra position gracefully.
	for _, mark := range marks {
		edits = append(edits, analysis.TextEdit{
			Pos:     mark.Pos,
			End:     mark.End + 1, // +1 to include the newline character
			NewText: []byte(""),
		})
	}

	return edits
}

// reportMarkers generates a diagnostic report for markers that should be
// replaced. This function handles the common logic for both field and type
// reporting.
func reportMarkers(pass *analysis.Pass, marks []markers.Marker, preferredIdentifier, elementName string, pos token.Pos, elementType string, preferredExists bool) {
	if len(marks) == 0 {
		return
	}

	markerWord := "marker"
	if len(marks) > 1 {
		markerWord = "markers"
	}

	message := fmt.Sprintf("%s %s uses %s %s, should use preferred marker %q instead",
		elementType, elementName, markerWord, formatMarkerList(marks), preferredIdentifier)

	fixMessage := "remove equivalent markers"
	if !preferredExists {
		fixMessage = fmt.Sprintf("replace with %q", preferredIdentifier)
	}

	pass.Report(analysis.Diagnostic{
		Pos:     pos,
		Message: message,
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message:   fixMessage,
				TextEdits: buildTextEdits(marks, preferredIdentifier, preferredExists),
			},
		},
	})
}

// buildReplacementMarker constructs the replacement marker text with the
// preferred identifier while preserving the structure from the original marker.
// It handles different marker types (DeclarativeValidation vs Kubebuilder) and
// properly reconstructs arguments and payloads.
func buildReplacementMarker(marker markers.Marker, preferredIdentifier string) string {
	// Determine the target marker type based on the preferred identifier
	// DeclarativeValidation markers start with "k8s:", others are Kubebuilder style
	targetType := markers.MarkerTypeKubebuilder
	if strings.HasPrefix(preferredIdentifier, "k8s:") {
		targetType = markers.MarkerTypeDeclarativeValidation
	}

	switch targetType {
	case markers.MarkerTypeDeclarativeValidation:
		return buildDeclarativeValidationMarker(marker, preferredIdentifier)
	case markers.MarkerTypeKubebuilder:
		return buildKubebuilderMarker(marker, preferredIdentifier)
	default:
		// Fallback for unknown marker types (should not happen in practice)
		return "// +" + preferredIdentifier
	}
}

// getSortedKeys returns sorted keys from a map for deterministic output.
func getSortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// buildDeclarativeValidationMarker reconstructs a DeclarativeValidation marker.
// Format: // +identifier(argName: argValue, ...)={payload.Value || reconstruct(payload.Marker)}.
func buildDeclarativeValidationMarker(marker markers.Marker, preferredIdentifier string) string {
	result := "// +" + preferredIdentifier

	// Add arguments if present
	if len(marker.Arguments) > 0 {
		parts := make([]string, 0, len(marker.Arguments))

		for _, key := range getSortedKeys(marker.Arguments) {
			if key == "" {
				parts = append(parts, marker.Arguments[key])
			} else {
				parts = append(parts, fmt.Sprintf("%s: %s", key, marker.Arguments[key]))
			}
		}

		result += "(" + strings.Join(parts, ", ") + ")"
	}

	// Add payload if present
	if marker.Payload.Value != "" {
		result += "=" + marker.Payload.Value
	} else if marker.Payload.Marker != nil {
		// Nested marker - reconstruct without "// +" prefix
		nested := buildReplacementMarker(*marker.Payload.Marker, marker.Payload.Marker.Identifier)
		result += "=" + strings.TrimPrefix(nested, "// +")
	}

	return result
}

// buildKubebuilderMarker reconstructs a Kubebuilder marker.
// Format with arguments: // +identifier:argOne="valueOne",argTwo="valueTwo".
// Format without arguments: // +identifier={payload.Value}.
func buildKubebuilderMarker(marker markers.Marker, preferredIdentifier string) string {
	result := "// +" + preferredIdentifier

	// Handle case with arguments
	if len(marker.Arguments) > 0 {
		parts := make([]string, 0, len(marker.Arguments))
		for _, key := range getSortedKeys(marker.Arguments) {
			parts = append(parts, fmt.Sprintf("%s=%s", key, marker.Arguments[key]))
		}

		return result + ":" + strings.Join(parts, ",")
	}

	// Handle case without arguments but with payload
	if marker.Payload.Value != "" {
		result += "=" + marker.Payload.Value
	}

	return result
}
