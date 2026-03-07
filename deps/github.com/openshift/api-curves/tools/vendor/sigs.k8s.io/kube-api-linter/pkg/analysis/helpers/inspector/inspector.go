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
package inspector

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	astinspector "golang.org/x/tools/go/ast/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
	markersconsts "sigs.k8s.io/kube-api-linter/pkg/markers"
)

// Inspector is an interface that allows for the inspection of fields in structs.
type Inspector interface {
	// InspectFields is a function that iterates over fields in structs.
	InspectFields(func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string))

	// InspectFieldsIncludingListTypes is a function that iterates over fields in structs, including list types.
	InspectFieldsIncludingListTypes(func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string))

	// InspectTypeSpec is a function that inspects the type spec and calls the provided inspectTypeSpec function.
	InspectTypeSpec(func(typeSpec *ast.TypeSpec, markersAccess markers.Markers))
}

// inspector implements the Inspector interface.
type inspector struct {
	inspector *astinspector.Inspector
	jsonTags  extractjsontags.StructFieldTags
	markers   markers.Markers
}

// newInspector creates a new inspector.
func newInspector(astinspector *astinspector.Inspector, jsonTags extractjsontags.StructFieldTags, markers markers.Markers) Inspector {
	return &inspector{
		inspector: astinspector,
		jsonTags:  jsonTags,
		markers:   markers,
	}
}

// InspectFields iterates over fields in structs, ignoring any struct that is not a type declaration, and any field that is ignored and
// therefore would not be included in the CRD spec.
// For the remaining fields, it calls the provided inspectField function to apply analysis logic.
func (i *inspector) InspectFields(inspectField func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string)) {
	i.inspectFields(inspectField, true)
}

// InspectFieldsIncludingListTypes iterates over fields in structs, including list types.
// Unlike InspectFields, this method does not skip fields in list type structs.
func (i *inspector) InspectFieldsIncludingListTypes(inspectField func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string)) {
	i.inspectFields(inspectField, false)
}

// inspectFields is a shared implementation for field iteration.
// The skipListTypes parameter controls whether list type structs should be skipped.
func (i *inspector) inspectFields(inspectField func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string), skipListTypes bool) {
	// Filter to fields so that we can iterate over fields in a struct.
	nodeFilter := []ast.Node{
		(*ast.Field)(nil),
	}

	i.inspector.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) (proceed bool) {
		if !push {
			return false
		}

		field, ok := n.(*ast.Field)
		if !ok || !i.shouldProcessField(stack, skipListTypes) {
			return ok
		}

		if i.shouldSkipField(field) {
			return false
		}

		var structName string

		qualifiedFieldName := utils.FieldName(field)
		if qualifiedFieldName == "" {
			qualifiedFieldName = types.ExprString(field.Type)
		}

		// The 0th node in the stack is the *ast.File.
		file, ok := stack[0].(*ast.File)
		if ok {
			structName = utils.GetStructNameFromFile(file, field)
		}

		if structName != "" {
			qualifiedFieldName = fmt.Sprintf("%s.%s", structName, qualifiedFieldName)
		}

		i.processFieldWithRecovery(field, qualifiedFieldName, inspectField)

		return true
	})
}

// shouldProcessField checks if the field should be processed.
// The skipListTypes parameter controls whether list type structs should be skipped.
func (i *inspector) shouldProcessField(stack []ast.Node, skipListTypes bool) bool {
	if len(stack) < 3 {
		return false
	}

	// The 0th node in the stack is the *ast.File.
	// The 1st node in the stack is the *ast.GenDecl.
	decl, ok := stack[1].(*ast.GenDecl)
	if !ok || decl.Tok != token.TYPE {
		// Make sure that we don't inspect structs within a function or non-type declarations.
		return false
	}

	structType, ok := stack[len(stack)-3].(*ast.StructType)
	if !ok {
		// Not in a struct.
		return false
	}

	if skipListTypes && utils.IsKubernetesListType(structType, "") {
		// Skip list types if requested.
		return false
	}

	return true
}

// shouldSkipField checks if a field should be skipped.
func (i *inspector) shouldSkipField(field *ast.Field) bool {
	tagInfo := i.jsonTags.FieldTags(field)
	if tagInfo.Ignored {
		return true
	}

	markerSet := i.markers.FieldMarkers(field)

	return isSchemalessType(markerSet)
}

// processFieldWithRecovery processes a field with panic recovery.
func (i *inspector) processFieldWithRecovery(field *ast.Field, qualifiedFieldName string, inspectField func(field *ast.Field, jsonTagInfo extractjsontags.FieldTagInfo, markersAccess markers.Markers, qualifiedFieldName string)) {
	tagInfo := i.jsonTags.FieldTags(field)

	defer func() {
		if r := recover(); r != nil {
			// If the inspectField function panics, we recover and log information that will help identify the issue.
			debug := printDebugInfo(field)
			panic(fmt.Sprintf("%s %v", debug, r)) // Re-panic to propagate the error.
		}
	}()

	inspectField(field, tagInfo, i.markers, qualifiedFieldName)
}

// InspectTypeSpec inspects the type spec and calls the provided inspectTypeSpec function.
func (i *inspector) InspectTypeSpec(inspectTypeSpec func(typeSpec *ast.TypeSpec, markersAccess markers.Markers)) {
	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	i.inspector.Preorder(nodeFilter, func(n ast.Node) {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return
		}

		inspectTypeSpec(typeSpec, i.markers)
	})
}

func isSchemalessType(markerSet markers.MarkerSet) bool {
	// Check if the field is marked as schemaless.
	schemalessMarker := markerSet.Get(markersconsts.KubebuilderSchemaLessMarker)
	return len(schemalessMarker) > 0
}

// printDebugInfo prints debug information about the field that caused a panic during inspection.
// This function is designed to allow us to help identify which fields are causing issues during inspection.
func printDebugInfo(field *ast.Field) string {
	var debug string

	debug += fmt.Sprintf("Panic observed while inspecting field: %v (type: %v)\n", utils.FieldName(field), field.Type)

	return debug
}
