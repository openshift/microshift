package manifestcomparators

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsvalidation "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel/model"
	"k8s.io/apimachinery/pkg/util/validation/field"
	celconfig "k8s.io/apiserver/pkg/apis/cel"
	apiservercel "k8s.io/apiserver/pkg/cel"
	"k8s.io/apiserver/pkg/cel/environment"
)

type mustNotExceedCostBudget struct{}

func MustNotExceedCostBudget() CRDComparator {
	return mustNotExceedCostBudget{}
}

func (mustNotExceedCostBudget) Name() string {
	return "MustNotExceedCostBudget"
}

func (mustNotExceedCostBudget) WhyItMatters() string {
	return ""
}

func (b mustNotExceedCostBudget) Validate(crd *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	errsToReport := []string{}
	warnings := []string{}
	infos := []string{}

	for _, newVersion := range crd.Spec.Versions {
		schema := &apiextensions.JSONSchemaProps{}
		if err := apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(newVersion.Schema.OpenAPIV3Schema, schema, nil); err != nil {
			errsToReport = append(errsToReport, err.Error())
			continue
		}

		rootCELContext := apiextensionsvalidation.RootCELContext(schema)

		SchemaHas(newVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, ancestry []*apiextensionsv1.JSONSchemaProps) bool {
				schema := &apiextensions.JSONSchemaProps{}
				if err := apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(s, schema, nil); err != nil {
					errsToReport = append(errsToReport, err.Error())
					return false
				}

				if schema.XValidations == nil {
					// There are no XValidations at this level, we do not need to continue with checks.
					return false
				}

				schemaInfos, schemaWarnings, err := inspectSchema(schema, simpleLocation, len(ancestry) == 0)
				if err != nil {
					errsToReport = append(errsToReport, err.Error())
				}
				infos = append(infos, schemaInfos...)
				warnings = append(warnings, schemaWarnings...)

				celContext, err := extractCELContext(append(ancestry, s), fldPath)
				if err != nil {
					errsToReport = append(errsToReport, err.Error())
					return false
				}

				typeInfo, err := celContext.TypeInfo()
				if err != nil {
					errsToReport = append(errsToReport, err.Error())
					return false
				}

				if typeInfo == nil {
					// No validations to check.
					return false
				}

				compResults, err := cel.Compile(
					typeInfo.Schema,
					typeInfo.DeclType,
					celconfig.PerCallLimit,
					environment.MustBaseEnvSet(environment.DefaultCompatibilityVersion(), true),
					cel.NewExpressionsEnvLoader(),
				)
				if err != nil {
					fieldErr := field.InternalError(fldPath, fmt.Errorf("failed to compile x-kubernetes-validations rules: %w", err))
					errsToReport = append(errsToReport, fieldErr.Error())
					return false
				}

				for i, cr := range compResults {
					if celContext.MaxCardinality == nil {
						unboundedParents, err := getUnboundedParentFields(ancestry, fldPath)
						if err != nil {
							errsToReport = append(errsToReport, err.Error())
						}
						warnings = append(warnings, fmt.Sprintf("%s: Field has unbounded cardinality. At least one, variable parent field does not have a maxItems or maxProperties constraint: %s. Falling back to CEL calculated worst case of %d executions.", simpleLocation.String(), strings.Join(unboundedParents, ","), cr.MaxCardinality))
					} else {
						msg := fmt.Sprintf("%s: Field has a maximum cardinality of %d.", simpleLocation.String(), *celContext.MaxCardinality)
						if *celContext.MaxCardinality > 1 {
							msg += " This is the calculated, worst case number of times the rule will be evaluated."
						}

						infos = append(infos, msg)
					}

					expressionCost := getExpressionCost(cr, celContext)

					if expressionCost > apiextensionsvalidation.StaticEstimatedCostLimit {
						costErrorMsg := getCostErrorMessage("estimated rule cost", expressionCost, apiextensionsvalidation.StaticEstimatedCostLimit)
						errsToReport = append(errsToReport, field.Forbidden(fldPath, costErrorMsg).Error())
					}
					if rootCELContext.TotalCost != nil {
						rootCELContext.TotalCost.ObserveExpressionCost(fldPath, expressionCost)
					}

					if cr.Error != nil {
						if cr.Error.Type == apiservercel.ErrorTypeRequired {
							errsToReport = append(errsToReport, field.Required(fldPath, cr.Error.Detail).Error())
						} else {
							errsToReport = append(errsToReport, field.Invalid(fldPath, schema.XValidations[i], cr.Error.Detail).Error())
						}
					} else {
						infos = append(infos, fmt.Sprintf("%s: Rule %d raw cost is %d. Estimated total cost of %d. The maximum allowable value is %d. Rule is %.2f%% of allowed budget.", simpleLocation.String(), i, cr.MaxCost, expressionCost, apiextensionsvalidation.StaticEstimatedCostLimit, float64(expressionCost*100)/apiextensionsvalidation.StaticEstimatedCostLimit))
					}

					if cr.MessageExpressionError != nil {
						errsToReport = append(errsToReport, field.Invalid(fldPath, schema.XValidations[i], cr.MessageExpressionError.Detail).Error())
					} else if cr.MessageExpression != nil {
						if cr.MessageExpressionMaxCost > apiextensionsvalidation.StaticEstimatedCostLimit {
							costErrorMsg := getCostErrorMessage("estimated messageExpression cost", cr.MessageExpressionMaxCost, apiextensionsvalidation.StaticEstimatedCostLimit)
							errsToReport = append(errsToReport, field.Forbidden(fldPath, costErrorMsg).Error())
						}
						if celContext.TotalCost != nil {
							celContext.TotalCost.ObserveExpressionCost(fldPath, cr.MessageExpressionMaxCost)
						}
					}
				}

				return false
			})

		if rootCELContext != nil && rootCELContext.TotalCost != nil && rootCELContext.TotalCost.Total > apiextensionsvalidation.StaticEstimatedCRDCostLimit {
			costErrorMsg := getCostErrorMessage("total CRD cost", rootCELContext.TotalCost.Total, apiextensionsvalidation.StaticEstimatedCRDCostLimit)
			errsToReport = append(errsToReport, field.Forbidden(field.NewPath("^"), costErrorMsg).Error())
		}
	}

	return ComparisonResults{
		Name:         b.Name(),
		WhyItMatters: b.WhyItMatters(),

		Errors:   errsToReport,
		Warnings: warnings,
		Infos:    infos,
	}, nil
}

func (b mustNotExceedCostBudget) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	return RatchetCompare(b, existingCRD, newCRD)
}

// multiplyWithOverflowGuard returns the product of baseCost and cardinality unless that product
// would exceed math.MaxUint, in which case math.MaxUint is returned.
func multiplyWithOverflowGuard(baseCost, cardinality uint64) uint64 {
	if baseCost == 0 {
		// an empty rule can return 0, so guard for that here
		return 0
	} else if math.MaxUint/baseCost < cardinality {
		return math.MaxUint
	}
	return baseCost * cardinality
}

// unbounded uses nil to represent an unbounded cardinality value.
var unbounded *uint64 = nil //nolint:revive // Using as a named variable to provide the meaning of nil in this context.

func getExpressionCost(cr cel.CompilationResult, cardinalityCost *apiextensionsvalidation.CELSchemaContext) uint64 {
	if cardinalityCost.MaxCardinality != unbounded {
		return multiplyWithOverflowGuard(cr.MaxCost, *cardinalityCost.MaxCardinality)
	}
	return multiplyWithOverflowGuard(cr.MaxCost, cr.MaxCardinality)
}

func getCostErrorMessage(costName string, expressionCost, costLimit uint64) string {
	exceedFactor := float64(expressionCost) / float64(costLimit)
	var factor string
	if exceedFactor > 100.0 {
		// if exceedFactor is greater than 2 orders of magnitude, the rule is likely O(n^2) or worse
		// and will probably never validate without some set limits
		// also in such cases the cost estimation is generally large enough to not add any value
		factor = "more than 100x"
	} else if exceedFactor < 1.5 {
		factor = fmt.Sprintf("%fx", exceedFactor) // avoid reporting "exceeds budge by a factor of 1.0x"
	} else {
		factor = fmt.Sprintf("%.1fx", exceedFactor)
	}
	return fmt.Sprintf("%s exceeds budget by factor of %s (try simplifying the rule(s), or adding maxItems, maxProperties, and maxLength where arrays, maps, and strings are declared)", costName, factor)
}

// extractCELContext takes a series of CEL contextxs and returns the child context of the last schema in the series.
// This is used so that the calculated maximum cardinality of the field is correct.
func extractCELContext(schemas []*apiextensionsv1.JSONSchemaProps, fldPath *field.Path) (*apiextensionsvalidation.CELSchemaContext, error) {
	var celContext *apiextensionsvalidation.CELSchemaContext

	for _, s := range schemas {
		schema := &apiextensions.JSONSchemaProps{}
		if err := apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(s, schema, nil); err != nil {
			return nil, fmt.Errorf("failed to convert schema: %w", err)
		}

		if celContext == nil {
			celContext = apiextensionsvalidation.RootCELContext(schema)
			continue
		}

		celContext = celContext.ChildPropertyContext(schema, s.ID)
	}

	return celContext, nil
}

// getUnboundedParentFields returns a list of field paths that have unbounded cardinality in the ancestry path.
// This is aiming to help users identify where the unbounded cardinality is coming from.
func getUnboundedParentFields(ancestry []*apiextensionsv1.JSONSchemaProps, fldPath *field.Path) ([]string, error) {
	cleanPathParts := getCleanPathParts(fldPath)
	var path *field.Path

	if len(ancestry)+1 != len(cleanPathParts) {
		// Ancestry does not include the field itself, the last part of cleanPathParts is the field itself.
		return nil, errors.New("ancestry and field path do not match")
	}

	unboundedParents := []string{}
	for i, s := range ancestry {
		schema := &apiextensions.JSONSchemaProps{}
		if err := apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(s, schema, nil); err != nil {
			continue
		}

		if path == nil {
			path = field.NewPath(cleanPathParts[i])
		} else if cleanPathParts[i] == "items" {
			path = path.Index(-1)
		} else {
			path = path.Child(cleanPathParts[i])
		}

		if isUnboundedCardinality(schema) {
			// Replace the -1 index with * that we use as a placeholder.
			unboundedParents = append(unboundedParents, strings.Replace(path.String(), "-1", "*", -1))
		}
	}
	return unboundedParents, nil
}

func getCleanPathParts(fldPath *field.Path) []string {
	cleanPathParts := []string{}
	for _, part := range strings.Split(fldPath.String(), ".") {
		if strings.HasPrefix(part, "properties[") {
			part = strings.TrimPrefix(strings.TrimSuffix(part, "]"), "properties[")
		}
		cleanPathParts = append(cleanPathParts, part)
	}
	return cleanPathParts
}

func isUnboundedCardinality(schema *apiextensions.JSONSchemaProps) bool {
	switch schema.Type {
	case "object":
		return schema.AdditionalProperties != nil && schema.MaxProperties == nil
	case "array":
		return schema.MaxItems == nil
	default:
		return false
	}
}

func inspectSchema(schema *apiextensions.JSONSchemaProps, simpleLocation *field.Path, isRoot bool) ([]string, []string, error) {
	typeInfo, err := getDeclType(schema, isRoot)
	if err != nil {
		return nil, nil, err
	}

	var infos, warnings []string

	switch schema.Type {
	case "string":
		switch {
		case len(schema.Enum) > 0:
			// Enums estimated lengths are based on the longest item in the list.
			// Not a concern for cost estimation generally.
		case schema.MaxLength == nil:
			warnings = append(warnings, fmt.Sprintf("%s: String has unbounded maxLength. It will be considered to have length %d. Consider adding a maxLength constraint to reduce the raw rule cost.", simpleLocation.String(), typeInfo.MaxElements))
		default:
			infos = append(infos, fmt.Sprintf("%s: String has maxLength of %d.", simpleLocation.String(), *schema.MaxLength))
		}
	case "array":
		switch {
		case schema.MaxItems == nil:
			warnings = append(warnings, fmt.Sprintf("%s: Array has unbounded maxItems. It will be considered to have %d items. Consider adding a maxItems constraint to reduce the raw rule cost.", simpleLocation.String(), typeInfo.MaxElements))
		default:
			infos = append(infos, fmt.Sprintf("%s: Array has maxItems of %d.", simpleLocation.String(), *schema.MaxItems))
		}
	}

	return infos, warnings, nil
}

func getDeclType(schema *apiextensions.JSONSchemaProps, isRoot bool) (*apiservercel.DeclType, error) {
	structural, err := structuralschema.NewStructural(schema)
	if err != nil {
		return nil, err
	}
	declType := model.SchemaDeclType(structural, isRoot)
	if declType == nil {
		return nil, fmt.Errorf("unable to convert structural schema to CEL declarations")
	}
	return declType, nil
}
