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
package serialization

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
)

// reportShouldAddPointer adds an analysis diagnostic that explains that a pointer should be added.
// Where the pointer policy is suggest fix, it also adds a suggested fix to add the pointer.
func reportShouldAddPointer(pass *analysis.Pass, field *ast.Field, pointerPolicy PointersPolicy, fieldName, messageFmt string, args ...any) {
	switch pointerPolicy {
	case PointersPolicySuggestFix:
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf(messageFmt, args...),
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "should make the field a pointer",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     field.Pos() + token.Pos(len(fieldName)+1),
							NewText: []byte("*"),
						},
					},
				},
			},
		})
	case PointersPolicyWarn:
		pass.Reportf(field.Pos(), messageFmt, args...)
	default:
		panic(fmt.Sprintf("unknown pointer policy: %s", pointerPolicy))
	}
}

// reportShouldRemovePointer adds an analysis diagnostic that explains that a pointer should be removed.
// Where the pointer policy is suggest fix, it also adds a suggested fix to remove the pointer.
func reportShouldRemovePointer(pass *analysis.Pass, field *ast.Field, pointerPolicy PointersPolicy, fieldName, messageFmt string, args ...any) {
	switch pointerPolicy {
	case PointersPolicySuggestFix:
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf(messageFmt, args...),
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "should remove the pointer",
					TextEdits: []analysis.TextEdit{
						{
							Pos: field.Pos() + token.Pos(len(fieldName)+1),
							End: field.Pos() + token.Pos(len(fieldName)+2),
						},
					},
				},
			},
		})
	case PointersPolicyWarn:
		pass.Reportf(field.Pos(), messageFmt, args...)
	default:
		panic(fmt.Sprintf("unknown pointer policy: %s", pointerPolicy))
	}
}

// reportShouldAddOmitEmpty adds an analysis diagnostic that explains that an omitempty tag should be added.
func reportShouldAddOmitEmpty(pass *analysis.Pass, field *ast.Field, omitEmptyPolicy OmitEmptyPolicy, qualifiedFieldName, messageFmt string, fieldTagInfo extractjsontags.FieldTagInfo) {
	switch omitEmptyPolicy {
	case OmitEmptyPolicySuggestFix:
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf(messageFmt, qualifiedFieldName),
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: fmt.Sprintf("should add 'omitempty' to the field tag for field %s", qualifiedFieldName),
					TextEdits: []analysis.TextEdit{
						{
							Pos:     fieldTagInfo.Pos + token.Pos(len(fieldTagInfo.Name)),
							NewText: []byte(",omitempty"),
						},
					},
				},
			},
		})
	case OmitEmptyPolicyWarn:
		pass.Reportf(field.Pos(), messageFmt, qualifiedFieldName)
	case OmitEmptyPolicyIgnore:
		// Do nothing, as the policy is to ignore the missing omitempty tag.
	default:
		panic(fmt.Sprintf("unknown omit empty policy: %s", omitEmptyPolicy))
	}
}

// reportShouldAddOmitZero adds an analysis diagnostic that explains that an omitzero tag should be added.
func reportShouldAddOmitZero(pass *analysis.Pass, field *ast.Field, omitZeroPolicy OmitZeroPolicy, qualifiedFieldName, messageFmt string, fieldTagInfo extractjsontags.FieldTagInfo) {
	switch omitZeroPolicy {
	case OmitZeroPolicySuggestFix:
		pass.Report(analysis.Diagnostic{
			Pos:     field.Pos(),
			Message: fmt.Sprintf(messageFmt, qualifiedFieldName),
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: fmt.Sprintf("should add 'omitzero' to the field tag for field %s", qualifiedFieldName),
					TextEdits: []analysis.TextEdit{
						{
							Pos:     fieldTagInfo.Pos + token.Pos(len(fieldTagInfo.Name)),
							NewText: []byte(",omitzero"),
						},
					},
				},
			},
		})
	case OmitZeroPolicyWarn:
		pass.Reportf(field.Pos(), messageFmt, qualifiedFieldName)
	case OmitZeroPolicyForbid:
		// Do nothing, as the policy is to forbid the missing omitzero tag.
	default:
		panic(fmt.Sprintf("unknown omit zero policy: %s", omitZeroPolicy))
	}
}

// reportShouldRemoveOmitZero adds an analysis diagnostic that explains that an omitzero tag should be removed.
func reportShouldRemoveOmitZero(pass *analysis.Pass, field *ast.Field, qualifiedFieldName string, jsonTags extractjsontags.FieldTagInfo) {
	omitZeroPos := jsonTags.Pos + token.Pos(strings.Index(jsonTags.RawValue, ",omitzero"))

	pass.Report(analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: fmt.Sprintf("field %s has the omitzero tag, but by policy is not allowed. The omitzero tag should be removed.", qualifiedFieldName),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "should remove the omitzero tag",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     omitZeroPos,
						End:     omitZeroPos + token.Pos(len(",omitzero")),
						NewText: nil, // Clear the omitzero tag.
					},
				},
			},
		},
	})
}
