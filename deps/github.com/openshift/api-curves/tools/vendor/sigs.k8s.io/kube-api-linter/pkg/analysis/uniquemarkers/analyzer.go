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
package uniquemarkers

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"k8s.io/apimachinery/pkg/util/sets"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
	markersconsts "sigs.k8s.io/kube-api-linter/pkg/markers"
)

const name = "uniquemarkers"

func init() {
	for _, uniqueMarker := range defaultUniqueMarkers() {
		markers.DefaultRegistry().Register(uniqueMarker.Identifier)
	}
}

type analyzer struct {
	uniqueMarkers []UniqueMarker
}

func newAnalyzer(cfg *UniqueMarkersConfig) *analysis.Analyzer {
	if cfg == nil {
		cfg = &UniqueMarkersConfig{}
	}

	a := &analyzer{
		uniqueMarkers: append(defaultUniqueMarkers(), cfg.CustomMarkers...),
	}

	return &analysis.Analyzer{
		Name:     name,
		Doc:      "Check that all markers that should be unique on a field/type are only present once",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer},
	}
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, _ extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string) {
		checkField(pass, field, markersAccess, a.uniqueMarkers, qualifiedFieldName)
	})

	inspect.InspectTypeSpec(func(typeSpec *ast.TypeSpec, markersAccess markers.Markers) {
		checkType(pass, typeSpec, markersAccess, a.uniqueMarkers)
	})

	return nil, nil //nolint:nilnil
}

func checkField(pass *analysis.Pass, field *ast.Field, markersAccess markers.Markers, uniqueMarkers []UniqueMarker, qualifiedFieldName string) {
	if field == nil || len(field.Names) == 0 {
		return
	}

	markers := utils.TypeAwareMarkerCollectionForField(pass, markersAccess, field)
	check(markers, uniqueMarkers, reportField(pass, field, qualifiedFieldName))
}

func checkType(pass *analysis.Pass, typeSpec *ast.TypeSpec, markersAccess markers.Markers, uniqueMarkers []UniqueMarker) {
	if typeSpec == nil {
		return
	}

	markers := markersAccess.TypeMarkers(typeSpec)
	check(markers, uniqueMarkers, reportType(pass, typeSpec))
}

func check(markerSet markers.MarkerSet, uniqueMarkers []UniqueMarker, reportFunc func(id string)) {
	for _, marker := range uniqueMarkers {
		marks := markerSet.Get(marker.Identifier)
		markSet := sets.New[string]()

		for _, mark := range marks {
			id := constructIdentifier(mark, marker.Attributes...)

			if markSet.Has(id) {
				reportFunc(id)
				continue
			}

			markSet.Insert(id)
		}
	}
}

// constructIdentifier returns a string that serves as a unique identifier for a
// marker based on the provided attributes that should be unique for the marker.
func constructIdentifier(marker markers.Marker, attributes ...string) string {
	// if there are no unique attributes, the unique identifier is just
	// the base marker identifier
	if len(attributes) == 0 {
		return marker.Identifier
	}

	switch marker.Type {
	case markers.MarkerTypeDeclarativeValidation:
		// If a marker doesn't specify the attribute, we should assume it is equivalent
		// to the empty string ("") so that we can still key on uniqueness of other attributes
		// effectively.
		id := fmt.Sprintf("%s(", marker.Identifier)

		for _, attr := range attributes {
			if attr == "" {
				id += marker.Arguments[attr]
				continue
			}

			id += fmt.Sprintf("%s: %s,", attr, marker.Arguments[attr])
		}

		id = strings.TrimSuffix(id, ",")
		id += ")"

		return id
	case markers.MarkerTypeKubebuilder:
		// If a marker doesn't specify the attribute, we should assume it is equivalent
		// to the empty string ("") so that we can still key on uniqueness of other attributes
		// effectively.
		id := fmt.Sprintf("%s:", marker.Identifier)
		for _, attr := range attributes {
			id += fmt.Sprintf("%s=%s,", attr, marker.Arguments[attr])
		}

		id = strings.TrimSuffix(id, ",")

		return id
	default:
		// programmer error
		panic(fmt.Sprintf("unknown marker format %s", marker.Type))
	}
}

func reportField(pass *analysis.Pass, field *ast.Field, qualifiedFieldName string) func(id string) {
	return func(id string) {
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf("field %s has multiple definitions of marker %s when only a single definition should exist", qualifiedFieldName, id),
		})
	}
}

func reportType(pass *analysis.Pass, typeSpec *ast.TypeSpec) func(id string) {
	return func(id string) {
		pass.Report(analysis.Diagnostic{
			Pos:     typeSpec.Pos(),
			Message: fmt.Sprintf("type %s has multiple definitions of marker %s when only a single definition should exist", typeSpec.Name, id),
		})
	}
}

//nolint:funlen
func defaultUniqueMarkers() []UniqueMarker {
	return []UniqueMarker{
		// Basic unique markers
		// ------
		{
			Identifier: markersconsts.DefaultMarker,
		},
		// ------

		// Kubebuilder-specific unique markers
		// ------
		{
			Identifier: markersconsts.KubebuilderDefaultMarker,
		},
		{
			Identifier: markersconsts.KubebuilderExampleMarker,
		},
		{
			Identifier: markersconsts.KubebuilderEnumMarker,
		},
		{
			Identifier: markersconsts.KubebuilderExclusiveMaximumMarker,
		},
		{
			Identifier: markersconsts.KubebuilderExclusiveMinimumMarker,
		},
		{
			Identifier: markersconsts.KubebuilderFormatMarker,
		},
		{
			Identifier: markersconsts.KubebuilderMaxItemsMarker,
		},
		{
			Identifier: markersconsts.KubebuilderMaxLengthMarker,
		},
		{
			Identifier: markersconsts.KubebuilderMaxPropertiesMarker,
		},
		{
			Identifier: markersconsts.KubebuilderMaximumMarker,
		},
		{
			Identifier: markersconsts.KubebuilderMinItemsMarker,
		},
		{
			Identifier: markersconsts.KubebuilderMinLengthMarker,
		},
		{
			Identifier: markersconsts.KubebuilderMinPropertiesMarker,
		},
		{
			Identifier: markersconsts.KubebuilderMinimumMarker,
		},
		{
			Identifier: markersconsts.KubebuilderMultipleOfMarker,
		},
		{
			Identifier: markersconsts.KubebuilderPatternMarker,
		},
		{
			Identifier: markersconsts.KubebuilderTypeMarker,
		},
		{
			Identifier: markersconsts.KubebuilderUniqueItemsMarker,
		},

		{
			Identifier: markersconsts.KubebuilderItemsEnumMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsFormatMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsMaxLengthMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsMaxItemsMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsMaxPropertiesMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsMaximumMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsMinLengthMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsMinItemsMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsMinPropertiesMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsMinimumMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsExclusiveMaximumMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsExclusiveMinimumMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsMultipleOfMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsPatternMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsTypeMarker,
		},
		{
			Identifier: markersconsts.KubebuilderItemsUniqueItemsMarker,
		},
		{
			Identifier: markersconsts.KubebuilderXValidationMarker,
			Attributes: []string{
				"rule",
			},
		},
		{
			Identifier: markersconsts.KubebuilderItemsXValidationMarker,
			Attributes: []string{
				"rule",
			},
		},
		// ------

		// K8s-specific unique markers
		// ------
		{
			Identifier: markersconsts.K8sFormatMarker,
		},
		{
			Identifier: markersconsts.K8sMinLengthMarker,
		},
		{
			Identifier: markersconsts.K8sMaxLengthMarker,
		},
		{
			Identifier: markersconsts.K8sMinItemsMarker,
		},
		{
			Identifier: markersconsts.K8sMaxItemsMarker,
		},
		{
			Identifier: markersconsts.K8sMinimumMarker,
		},
		{
			Identifier: markersconsts.K8sMaximumMarker,
		},
		{
			Identifier: markersconsts.K8sListTypeMarker,
		},
		// ------
	}
}
