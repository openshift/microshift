package emptypartialschemas

import (
	"fmt"
	"strconv"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/gengo/v2"
	"k8s.io/gengo/v2/types"
)

// known tags that are handled by empty partial schema generation.
const (
	openshiftPackageGenerationEnablementMarkerName = "openshift:featuregated-schema-gen"
	tagGroupName                                   = "groupName"
	kubeBuilderObjectRoot                          = "kubebuilder:object:root"
	kubeBuilderResource                            = "kubebuilder:resource"
	kubeBuilderStatus                              = "kubebuilder:subresource:status"
	kubeBuilderPrinterColumn                       = "kubebuilder:printcolumn"
	kubeBuilderMetadata                            = "kubebuilder:metadata"

	openshiftCRDFilenameMarkerName                 = "openshift:file-pattern"
	openshiftApprovedPRMarkerName                  = "openshift:api-approved.openshift.io"
	openshiftCapabilityMarkerName                  = "openshift:capability"
	openshiftFeatureGateMarkerName                 = "openshift:enable:FeatureGate"
	openshiftFeatureGateAwareEnumMarkerName        = "openshift:validation:FeatureGateAwareEnum"
	openshiftFeatureGateAwareMaxItemsMarkerName    = "openshift:validation:FeatureGateAwareMaxItems"
	openshiftFeatureGateAwareXValidationMarkerName = "openshift:validation:FeatureGateAwareXValidation"
)

func extractFeatureGatesForType(t *types.Type) []string {
	return extractFeatureGatesFromComments(allCommentsForType(t))
}

func extractFeatureGatesFromComments(comments []string) []string {
	anyPossibilityFound := false
	for _, line := range comments {
		if strings.Contains(line, openshiftFeatureGateMarkerName) {
			anyPossibilityFound = true
			break
		}
		if strings.Contains(line, openshiftFeatureGateAwareXValidationMarkerName) {
			anyPossibilityFound = true
			break
		}
		if strings.Contains(line, openshiftFeatureGateAwareEnumMarkerName) {
			anyPossibilityFound = true
			break
		}
		if strings.Contains(line, openshiftFeatureGateAwareMaxItemsMarkerName) {
			anyPossibilityFound = true
			break
		}
	}
	if !anyPossibilityFound {
		return nil
	}

	ret := []string{}
	ret = append(ret, extractStringSliceTagFromComments(comments, openshiftFeatureGateMarkerName)...)
	ret = append(ret, extractFeatureGatesFromValidationMarker(comments, openshiftFeatureGateAwareXValidationMarkerName)...)
	ret = append(ret, extractFeatureGatesFromValidationMarker(comments, openshiftFeatureGateAwareEnumMarkerName)...)
	ret = append(ret, extractFeatureGatesFromValidationMarker(comments, openshiftFeatureGateAwareMaxItemsMarkerName)...)

	return ret
}

func extractFeatureGatesFromValidationMarker(comments []string, tagName string) []string {
	ret := []string{}
	tagVals := kubeBuilderCompatibleExtractValueStringsForMarkerFromComments(comments, tagName)

	for _, tagVal := range tagVals {
		featureGatesToSplit := extractNamedValuesFromSingleLine(tagVal)["featureGate"]
		featureGates := strings.Split(featureGatesToSplit, ";")

		for _, featureGate := range featureGates {
			if featureGate == `""` {
				// this happens when the "no feature gate" is needed for things like enums
				continue
			}
			if len(featureGate) > 0 {
				ret = append(ret, featureGate)
			}
		}
	}

	for _, tagVal := range tagVals {
		requiredFeatureGates := extractNamedValuesFromSingleLine(tagVal)["requiredFeatureGate"]

		// Use + as the separator between required feature gates in file names.
		requiredFeatureGates = strings.Replace(requiredFeatureGates, ";", "+", -1)

		if len(requiredFeatureGates) > 0 {
			ret = append(ret, requiredFeatureGates)
		}
	}

	return ret
}

func extractNamedValuesForType(t *types.Type, tagName string) (map[string]string, error) {
	return extractNamedValues(allCommentsForType(t), tagName)
}

func extractNamedValues(comments []string, tagName string) (map[string]string, error) {
	tagVals := gengo.ExtractCommentTags("+", comments)[tagName]
	if len(tagVals) > 1 {
		return nil, fmt.Errorf("too many tag values: %d", len(tagVals))
	}
	if len(tagVals) == 0 {
		return nil, nil
	}

	return extractNamedValuesFromSingleLine(tagVals[0]), nil
}

func extractPrinterColumnsForType(t *types.Type) []apiextensionsv1.CustomResourceColumnDefinition {
	return extractPrinterColumnsFromComments(allCommentsForType(t))
}

func extractPrinterColumnsFromComments(comments []string) []apiextensionsv1.CustomResourceColumnDefinition {
	ret := []apiextensionsv1.CustomResourceColumnDefinition{}

	lines := kubeBuilderCompatibleExtractValueForMarkerFromComments(comments, kubeBuilderPrinterColumn)
	for _, lineValues := range lines {
		ret = append(ret, apiextensionsv1.CustomResourceColumnDefinition{
			Name:        maybeStripQuotes(lineValues["name"]),
			Type:        maybeStripQuotes(lineValues["type"]),
			JSONPath:    maybeStripQuotes(lineValues["JSONPath"]),
			Description: maybeStripQuotes(lineValues["description"]),
			Priority:    extractInt(lineValues["priority"]),
		})
	}

	return ret
}

func extractMetadataForType(t *types.Type) (map[string]string, map[string]string) {
	return extractMetadataFromComments(allCommentsForType(t))
}

func extractMetadataFromComments(comments []string) (map[string]string, map[string]string) {
	annotations := map[string]string{}
	labels := map[string]string{}

	lines := kubeBuilderCompatibleExtractValueForMarkerFromComments(comments, kubeBuilderMetadata)
	for _, lineValues := range lines {
		currJoinedAnnotation := maybeStripQuotes(lineValues["annotations"])
		if len(currJoinedAnnotation) > 0 {
			tokens := strings.SplitN(currJoinedAnnotation, "=", 2)
			annotations[tokens[0]] = tokens[1]
		}
		currJoinedLabel := maybeStripQuotes(lineValues["labels"])
		if len(currJoinedLabel) > 0 {
			tokens := strings.SplitN(currJoinedLabel, "=", 2)
			labels[tokens[0]] = tokens[1]
		}
	}

	return annotations, labels
}

func extractInt(in string) int32 {
	if len(in) == 0 {
		return 0
	}

	ret, err := strconv.Atoi(in)
	if err != nil {
		return 0
	}
	return int32(ret)
}

func maybeStripQuotes(in string) string {
	if !strings.HasPrefix(in, `"`) || !strings.HasSuffix(in, `"`) {
		return in
	}
	ret, err := strconv.Unquote(in)
	if err != nil {
		return in
	}
	return ret
}

func extractNamedValuesFromSingleLine(singleLine string) map[string]string {
	currMap := map[string]string{}

	keyValuePairings := strings.Split(singleLine, ",")
	for _, currPairing := range keyValuePairings {
		tokens := strings.SplitN(currPairing, "=", 2)
		if len(tokens) == 0 {
			continue
		}
		if len(tokens) == 1 {
			currMap[tokens[0]] = ""
			continue
		}
		currMap[tokens[0]] = tokens[1]
	}

	return currMap
}

func allCommentsForType(t *types.Type) []string {
	return append(append([]string{}, t.SecondClosestCommentLines...), t.CommentLines...)
}

func extractStringTagForType(t *types.Type, tagName string) string {
	ret, _, _ := extractStringTagFromComments(allCommentsForType(t), tagName)
	return ret
}

func extractStringTagFromComments(comments []string, tagName string) (string, bool, error) {
	tagVals := gengo.ExtractCommentTags("+", comments)[tagName]
	if tagVals == nil {
		// No match for the tag.
		return "", false, nil
	}
	// If there are multiple values, abort.
	if len(tagVals) > 1 {
		return "", false, fmt.Errorf("Found %d %s tags: %q", len(tagVals), tagName, tagVals)
	}

	return tagVals[0], true, nil
}

func extractStringSliceTagForType(t *types.Type, tagName string) []string {
	return extractStringSliceTagFromComments(allCommentsForType(t), tagName)
}

func extractStringSliceTagFromComments(comments []string, tagName string) []string {
	ret := []string{}
	tagVals := gengo.ExtractCommentTags("+", comments)[tagName]
	for _, tagVal := range tagVals {
		values := strings.Split(tagVal, ",")
		for _, curr := range values {
			if len(curr) > 0 {
				ret = append(ret, curr)
			}
		}
	}
	return ret
}

func tagExistsForType(t *types.Type, tagName string) bool {
	return tagExistsFromComments(allCommentsForType(t), tagName)
}

func tagExistsFromComments(comments []string, tagName string) bool {
	_, ok := gengo.ExtractCommentTags("+", comments)[tagName]
	return ok
}

func isCRDType(t *types.Type) string {
	return extractStringTagForType(t, kubeBuilderObjectRoot)
}

// This is from kubebuilder.  They created a hard to parse syntax that is close to, but not quite gengo.
// I suspect it resulted from a misunderstanding, but I'm not certain.
func kubeBuilderCompatibleExtractValuesForMarkerForType(t *types.Type, markerName string) []map[string]string {
	return kubeBuilderCompatibleExtractValueForMarkerFromComments(allCommentsForType(t), markerName)
}

func kubeBuilderCompatibleExtractValueForMarkerFromComments(comments []string, tagName string) []map[string]string {
	ret := []map[string]string{}

	tagVals := kubeBuilderCompatibleExtractValueStringsForMarkerFromComments(comments, tagName)

	for _, tagVal := range tagVals {
		currMap := extractNamedValuesFromSingleLine(tagVal)
		ret = append(ret, currMap)
	}

	return ret
}

func kubeBuilderCompatibleExtractValueStringsForMarkerForType(t *types.Type, markerName string) []string {
	return kubeBuilderCompatibleExtractValueStringsForMarkerFromComments(allCommentsForType(t), markerName)
}

func kubeBuilderCompatibleExtractValueStringsForMarkerFromComments(lines []string, markerName string) []string {
	ret := []string{}
	for _, line := range lines {
		line = strings.Trim(line, " ")
		if len(line) == 0 {
			continue
		}
		if !strings.HasPrefix(line, "+"+markerName) {
			continue
		}

		nameIfNormal, nameIfKubeBuilderStruct, theRestOfTheArgs := splitMarker(line)
		firstPartOfArg := nameIfKubeBuilderStruct[len(nameIfNormal)+1:]
		ret = append(ret, fmt.Sprintf("%s=%s", firstPartOfArg, theRestOfTheArgs))
	}
	return ret
}

// splitMarker takes a marker in the form of `+a:b:c=arg,d=arg` and splits it
// into the name (`a:b`), the name if it's not a struct (`a:b:c`), and the parts
// that are definitely fields (`arg,d=arg`).
func splitMarker(raw string) (name string, anonymousName string, restFields string) {
	raw = raw[1:] // get rid of the leading '+'
	nameFieldParts := strings.SplitN(raw, "=", 2)
	if len(nameFieldParts) == 1 {
		return nameFieldParts[0], nameFieldParts[0], ""
	}
	anonymousName = nameFieldParts[0]
	name = anonymousName
	restFields = nameFieldParts[1]

	nameParts := strings.Split(name, ":")
	if len(nameParts) > 1 {
		name = strings.Join(nameParts[:len(nameParts)-1], ":")
	}
	return name, anonymousName, restFields
}
