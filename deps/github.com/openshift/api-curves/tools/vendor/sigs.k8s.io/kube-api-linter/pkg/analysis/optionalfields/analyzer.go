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
package optionalfields

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	markershelper "sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils/serialization"
	"sigs.k8s.io/kube-api-linter/pkg/markers"
)

const (
	name = "optionalfields"
)

func init() {
	markershelper.DefaultRegistry().Register(
		markers.OptionalMarker,
		markers.RequiredMarker,
		markers.KubebuilderOptionalMarker,
		markers.KubebuilderRequiredMarker,
		markers.KubebuilderMinItemsMarker,
		markers.KubebuilderMinLengthMarker,
		markers.KubebuilderMinPropertiesMarker,
		markers.KubebuilderMinimumMarker,
		markers.KubebuilderEnumMarker,
	)
}

type analyzer struct {
	serializationCheck serialization.SerializationCheck
}

// newAnalyzer creates a new analyzer.
func newAnalyzer(cfg *OptionalFieldsConfig) *analysis.Analyzer {
	if cfg == nil {
		cfg = &OptionalFieldsConfig{}
	}

	defaultConfig(cfg)

	serializationCheck := serialization.New(&serialization.Config{
		Pointers: serialization.PointersConfig{
			Policy:     serialization.PointersPolicy(cfg.Pointers.Policy),
			Preference: serialization.PointersPreference(cfg.Pointers.Preference),
		},
		OmitEmpty: serialization.OmitEmptyConfig{
			Policy: serialization.OmitEmptyPolicy(cfg.OmitEmpty.Policy),
		},
		OmitZero: serialization.OmitZeroConfig{
			Policy: serialization.OmitZeroPolicy(cfg.OmitZero.Policy),
		},
	})

	a := &analyzer{
		serializationCheck: serializationCheck,
	}

	return &analysis.Analyzer{
		Name: name,
		Doc: `Checks all optional fields comply with the configured policy.
		Depending on the configuration, this may include checking for the presence of the omitempty tag or
		whether the field is a pointer.
		For structs, this includes checking that if the field is marked as optional, it should be a pointer when it has omitempty.
		Where structs include required fields, they must be a pointer when they themselves are optional.
		`,
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer, extractjsontags.Analyzer},
	}
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markershelper.Markers, qualifiedFieldName string) {
		a.checkField(pass, field, markersAccess, jsonTagInfo, qualifiedFieldName)
	})

	return nil, nil //nolint:nilnil
}

func (a *analyzer) checkField(pass *analysis.Pass, field *ast.Field, markersAccess markershelper.Markers, jsonTags extractjsontags.FieldTagInfo, qualifiedFieldName string) {
	if field == nil || len(field.Names) == 0 {
		return
	}

	if !utils.IsFieldOptional(field, markersAccess) {
		// The field is not marked optional, so we don't need to check it.
		return
	}

	if field.Type == nil {
		// The field has no type? We can't check if it's a pointer.
		return
	}

	a.serializationCheck.Check(pass, field, markersAccess, jsonTags, qualifiedFieldName)
}

func defaultConfig(cfg *OptionalFieldsConfig) {
	if cfg.Pointers.Policy == "" {
		cfg.Pointers.Policy = OptionalFieldsPointerPolicySuggestFix
	}

	if cfg.Pointers.Preference == "" {
		cfg.Pointers.Preference = OptionalFieldsPointerPreferenceAlways
	}

	if cfg.OmitEmpty.Policy == "" {
		cfg.OmitEmpty.Policy = OptionalFieldsOmitEmptyPolicySuggestFix
	}

	if cfg.OmitZero.Policy == "" {
		cfg.OmitZero.Policy = OptionalFieldsOmitZeroPolicySuggestFix
	}
}
