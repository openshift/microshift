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
package namingconventions

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	kalerrors "sigs.k8s.io/kube-api-linter/pkg/analysis/errors"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/extractjsontags"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/inspector"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/helpers/markers"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/utils"
)

const name = "namingconventions"

type analyzer struct {
	conventions []Convention
}

// newAnalyzer creates a new analysis.Analyzer for the namingconventions
// linter based on the provided config.
func newAnalyzer(cfg *Config) *analysis.Analyzer {
	a := &analyzer{
		conventions: cfg.Conventions,
	}

	analyzer := &analysis.Analyzer{
		Name:     name,
		Doc:      "Enforces naming conventions on fields",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspector.Analyzer, extractjsontags.Analyzer},
	}

	return analyzer
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	inspect, ok := pass.ResultOf[inspector.Analyzer].(inspector.Inspector)
	if !ok {
		return nil, kalerrors.ErrCouldNotGetInspector
	}

	inspect.InspectFields(func(field *ast.Field, jsonTags extractjsontags.FieldTagInfo, _ markers.Markers, qualifiedFieldName string) {
		checkField(pass, field, jsonTags, qualifiedFieldName, a.conventions...)
	})

	return nil, nil //nolint:nilnil
}

func checkField(pass *analysis.Pass, field *ast.Field, tagInfo extractjsontags.FieldTagInfo, qualifiedFieldName string, conventions ...Convention) {
	if field == nil || len(field.Names) == 0 {
		return
	}

	fieldName := utils.FieldName(field)

	for _, convention := range conventions {
		// regexp.MustCompile will panic if the regular expression doesn't compile.
		// This should be reasonable as any regular expressions in naming conventions
		// will have already been validated to compile during the configuration validation stage.
		matcher := regexp.MustCompile(convention.ViolationMatcher)

		if !matcher.MatchString(fieldName) && !matcher.MatchString(tagInfo.Name) {
			continue
		}

		switch convention.Operation {
		case OperationInform:
			reportConventionWithSuggestedFixes(pass, field, convention, qualifiedFieldName)

		case OperationDropField:
			reportDropField(pass, field, convention, qualifiedFieldName)

		case OperationDrop:
			reportDrop(pass, field, tagInfo, convention, matcher, qualifiedFieldName)

		case OperationReplacement:
			reportReplace(pass, field, tagInfo, convention, matcher, qualifiedFieldName)
		}
	}
}

func reportConventionWithSuggestedFixes(pass *analysis.Pass, field *ast.Field, convention Convention, qualifiedFieldName string, suggestedFixes ...analysis.SuggestedFix) {
	pass.Report(analysis.Diagnostic{
		Pos:            field.Pos(),
		Message:        fmt.Sprintf("naming convention %q: field %s: %s", convention.Name, qualifiedFieldName, convention.Message),
		SuggestedFixes: suggestedFixes,
	})
}

func reportDropField(pass *analysis.Pass, field *ast.Field, convention Convention, qualifiedFieldName string) {
	suggestedFixes := []analysis.SuggestedFix{
		{
			Message: "remove the field",
			TextEdits: []analysis.TextEdit{
				{
					Pos:     field.Pos(),
					NewText: []byte(""),
					End:     field.End(),
				},
			},
		},
	}

	reportConventionWithSuggestedFixes(pass, field, convention, qualifiedFieldName, suggestedFixes...)
}

func reportDrop(pass *analysis.Pass, field *ast.Field, tagInfo extractjsontags.FieldTagInfo, convention Convention, matcher *regexp.Regexp, qualifiedFieldName string) {
	suggestedFixes := suggestedFixesForReplacement(field, tagInfo, matcher, "")
	reportConventionWithSuggestedFixes(pass, field, convention, qualifiedFieldName, suggestedFixes...)
}

func reportReplace(pass *analysis.Pass, field *ast.Field, tagInfo extractjsontags.FieldTagInfo, convention Convention, matcher *regexp.Regexp, qualifiedFieldName string) {
	suggestedFixes := suggestedFixesForReplacement(field, tagInfo, matcher, convention.Replacement)
	reportConventionWithSuggestedFixes(pass, field, convention, qualifiedFieldName, suggestedFixes...)
}

func suggestedFixesForReplacement(field *ast.Field, tagInfo extractjsontags.FieldTagInfo, matcher *regexp.Regexp, replacementStr string) []analysis.SuggestedFix {
	suggestedFixes := []analysis.SuggestedFix{}

	suggestedFixes = append(suggestedFixes, suggestFixesForGoFieldName(field, matcher, replacementStr)...)
	suggestedFixes = append(suggestedFixes, suggestFixesForSerializedFieldName(tagInfo, matcher, replacementStr)...)

	return suggestedFixes
}

func suggestFixesForSerializedFieldName(tagInfo extractjsontags.FieldTagInfo, matcher *regexp.Regexp, replacementStr string) []analysis.SuggestedFix {
	replacement := matcher.ReplaceAllString(tagInfo.Name, replacementStr)

	// If dropping the offending text from the field name would result in an empty
	// field name, just issue the failure with no suggested fix.
	if len(replacement) == 0 {
		return nil
	}

	// This should prevent panics from slice access when the replacement
	// string ends up being a length of 1 and still result in a technically
	// correct JSON tag name value.
	tagNameReplacement := strings.ToLower(replacement)
	if len(replacement) > 1 {
		tagNameReplacement = fmt.Sprintf("%s%s", strings.ToLower(replacement[:1]), replacement[1:])
	}

	tagReplacement := strings.ReplaceAll(tagInfo.RawValue, tagInfo.Name, tagNameReplacement)

	tagReplacementMessage := fmt.Sprintf("replace offending text in serialized field name with %q", replacementStr)

	if len(replacementStr) == 0 {
		tagReplacementMessage = "remove offending text from serialized field name"
	}

	return []analysis.SuggestedFix{
		{
			Message: tagReplacementMessage,
			TextEdits: []analysis.TextEdit{
				{
					Pos:     tagInfo.Pos,
					NewText: []byte(tagReplacement),
					End:     tagInfo.End,
				},
			},
		},
	}
}

func suggestFixesForGoFieldName(field *ast.Field, matcher *regexp.Regexp, replacementStr string) []analysis.SuggestedFix {
	fieldName := utils.FieldName(field)
	replacement := matcher.ReplaceAllString(fieldName, replacementStr)

	// If dropping the offending text from the field name would result in an empty
	// field name, just issue the failure with no suggested fix.
	if len(replacement) == 0 {
		return nil
	}

	replacementMessage := fmt.Sprintf("replace offending text in Go type with %q", replacementStr)

	if len(replacementStr) == 0 {
		replacementMessage = "remove offending text from Go type field"
	}

	return []analysis.SuggestedFix{
		{
			Message: replacementMessage,
			TextEdits: []analysis.TextEdit{
				{
					Pos:     field.Pos(),
					NewText: []byte(replacement),
					End:     field.Pos() + token.Pos(len(fieldName)),
				},
			},
		},
	}
}
