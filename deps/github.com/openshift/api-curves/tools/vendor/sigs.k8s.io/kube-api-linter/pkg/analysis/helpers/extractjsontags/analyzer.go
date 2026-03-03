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
package extractjsontags

import (
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"k8s.io/utils/set"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
)

const (
	omitEmpty = "omitempty"
	omitZero  = "omitzero"
)

// StructFieldTags is used to find information about
// json tags on fields within struct.
type StructFieldTags interface {
	FieldTags(*ast.Field) FieldTagInfo
}

type structFieldTags struct {
	fieldTags map[*ast.Field]FieldTagInfo
}

func newStructFieldTags() StructFieldTags {
	return &structFieldTags{
		fieldTags: make(map[*ast.Field]FieldTagInfo),
	}
}

func (s *structFieldTags) insertFieldTagInfo(field *ast.Field, tagInfo FieldTagInfo) {
	s.fieldTags[field] = tagInfo
}

// FieldTags find the tag information for the named field within the given struct.
func (s *structFieldTags) FieldTags(field *ast.Field) FieldTagInfo {
	return s.fieldTags[field]
}

// Analyzer is the analyzer for the jsontags package.
// It checks that all struct fields in an API are tagged with json tags.
var Analyzer = &analysis.Analyzer{
	Name:       "extractjsontags",
	Doc:        "Iterates over all fields in structs and extracts their json tags.",
	Run:        run,
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	ResultType: reflect.TypeOf(newStructFieldTags()),
}

func run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	// Filter to structs so that we can iterate over fields in a struct.
	nodeFilter := []ast.Node{
		(*ast.Field)(nil),
	}

	results, ok := newStructFieldTags().(*structFieldTags)
	if !ok {
		return nil, kalerrors.ErrCouldNotCreateStructFieldTags
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		field, ok := n.(*ast.Field)
		if !ok {
			return
		}

		results.insertFieldTagInfo(field, extractTagInfo(field.Tag))
	})

	return results, nil
}

//nolint:cyclop
func extractTagInfo(tag *ast.BasicLit) FieldTagInfo {
	if tag == nil || tag.Value == "" {
		return FieldTagInfo{Missing: true}
	}

	rawTag, err := strconv.Unquote(tag.Value)
	if err != nil {
		// This means the way AST is treating tags has changed.
		panic(err)
	}

	tagValue, ok := reflect.StructTag(rawTag).Lookup("json")
	if !ok {
		return FieldTagInfo{Missing: true}
	}

	if tagValue == "" {
		return FieldTagInfo{}
	}

	pos := tag.Pos() + token.Pos(strings.Index(tag.Value, tagValue))
	end := pos + token.Pos(len(tagValue))

	tagValues := strings.Split(tagValue, ",")

	if len(tagValues) == 1 && tagValues[0] == "-" {
		return FieldTagInfo{
			Ignored:  true,
			RawValue: tagValue,
			Pos:      pos,
			End:      end,
		}
	}

	if len(tagValues) == 2 && tagValues[0] == "" && tagValues[1] == "inline" {
		return FieldTagInfo{
			Inline:   true,
			RawValue: tagValue,
			Pos:      pos,
			End:      end,
		}
	}

	tagName := tagValues[0]
	tagSet := set.New[string](tagValues...)

	return FieldTagInfo{
		Name:      tagName,
		OmitEmpty: tagSet.Has(omitEmpty),
		OmitZero:  tagSet.Has(omitZero),
		RawValue:  tagValue,
		Pos:       pos,
		End:       end,
	}
}

// FieldTagInfo contains information about a field's json tag.
// This is used to pass information about a field's json tag between analyzers.
type FieldTagInfo struct {
	// Name is the name of the field extracted from the json tag.
	Name string

	// Ignored is true if the field is ignored by json package.
	Ignored bool

	// OmitEmpty is true if the field has the omitempty option in the json tag.
	OmitEmpty bool

	// OmitZero is true if the field has the omitzero option in the json tag.
	OmitZero bool

	// Inline is true if the field has the inline option in the json tag.
	Inline bool

	// Missing is true when the field had no json tag.
	Missing bool

	// RawValue is the raw value from the json tag.
	RawValue string

	// Pos marks the starting position of the json tag value.
	Pos token.Pos

	// End marks the end of the json tag value.
	End token.Pos
}
