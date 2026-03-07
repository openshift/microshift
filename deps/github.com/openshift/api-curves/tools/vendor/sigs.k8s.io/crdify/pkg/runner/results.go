// Copyright 2025 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
	"sigs.k8s.io/crdify/pkg/validations"
)

// Results is a utility type to hold the validation results of
// running different validators.
type Results struct {
	// CRDValidation is the set of validation comparison results
	// at the whole CustomResourceDefinition scope
	CRDValidation []validations.ComparisonResult `json:"crdValidation,omitempty"`

	// SameVersionValidation is the set of validation comparison
	// results at the CustomResourceDefinition version level. Specifically
	// for same version comparisons across an old and new CustomResourceDefinition
	// instance (i.e comparing v1alpha1 with v1alpha1)
	SameVersionValidation map[string]map[string][]validations.ComparisonResult `json:"sameVersionValidation,omitempty"`

	// ServedVersionValidation is the set of validation comparison
	// results at the CustomResourceDefinition version level. Specifically
	// for served version comparisons across an old and new CustomResourceDefinition
	// instance (i.e comparing v1alpha1 with v1 if both are served)
	ServedVersionValidation map[string]map[string][]validations.ComparisonResult `json:"servedVersionValidation,omitempty"`
}

// Format is a representation of an output format.
type Format string

const (
	// FormatJSON represents a JSON output format.
	FormatJSON Format = "json"

	// FormatYAML represents a YAML output format.
	FormatYAML Format = "yaml"

	// FormatPlainText represents a PlainText output format.
	FormatPlainText Format = "plaintext"

	// FormatMarkdown represents a Markdown output format.
	FormatMarkdown Format = "markdown"
)

// Render returns the string representation of the provided
// format or an error if one is encountered.
// Currently supported render formats are json, yaml, plaintext, and markdown.
// Unknown formats will result in an error.
func (rr *Results) Render(format Format) (string, error) {
	switch format {
	case FormatJSON:
		return rr.RenderJSON()
	case FormatYAML:
		return rr.RenderYAML()
	case FormatMarkdown:
		return rr.RenderMarkdown(), nil
	case FormatPlainText:
		return rr.RenderPlainText(), nil
	default:
		return "", fmt.Errorf("%w : %q", errUnknownRenderFormat, format)
	}
}

var errUnknownRenderFormat = errors.New("unknown render format")

// RenderJSON returns a string of the results rendered in JSON or an error.
func (rr *Results) RenderJSON() (string, error) {
	outBytes, err := json.MarshalIndent(rr, "", " ")
	return string(outBytes), err
}

// RenderYAML returns a string of the results rendered in YAML or an error.
func (rr *Results) RenderYAML() (string, error) {
	outBytes, err := yaml.Marshal(rr)
	return string(outBytes), err
}

// RenderMarkdown returns a string of the results rendered as Markdown
//
//nolint:dupl
func (rr *Results) RenderMarkdown() string { //nolint:gocognit,cyclop
	var out strings.Builder

	out.WriteString("# CRD Validations\n")

	for _, result := range rr.CRDValidation {
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				out.WriteString(fmt.Sprintf("- **%s** - `ERROR` - %s\n", result.Name, err))
			}
		}

		if len(result.Warnings) > 0 {
			for _, err := range result.Warnings {
				out.WriteString(fmt.Sprintf("- **%s** - `WARNING` - %s\n", result.Name, err))
			}
		}

		if len(result.Errors) == 0 && len(result.Warnings) == 0 {
			out.WriteString(fmt.Sprintf("- **%s** - ✓\n", result.Name))
		}
	}

	out.WriteString("\n\n")
	out.WriteString("# Same Version Validations\n")

	for version, result := range rr.SameVersionValidation {
		for property, results := range result {
			for _, propertyResult := range results {
				if len(propertyResult.Errors) > 0 {
					for _, err := range propertyResult.Errors {
						out.WriteString(fmt.Sprintf("- **%s** - *%s* - %s - `ERROR` - %s\n", version, property, propertyResult.Name, err))
					}
				}

				if len(propertyResult.Warnings) > 0 {
					for _, err := range propertyResult.Warnings {
						out.WriteString(fmt.Sprintf("- **%s** - *%s* - %s - `WARNING` - %s\n", version, property, propertyResult.Name, err))
					}
				}

				if len(propertyResult.Errors) == 0 && len(propertyResult.Warnings) == 0 {
					out.WriteString(fmt.Sprintf("- **%s** - *%s* - %s - ✓\n", version, property, propertyResult.Name))
				}
			}
		}
	}

	out.WriteString("\n\n")
	out.WriteString("# Served Version Validations\n")

	for version, result := range rr.ServedVersionValidation {
		for property, results := range result {
			for _, propertyResult := range results {
				if len(propertyResult.Errors) > 0 {
					for _, err := range propertyResult.Errors {
						out.WriteString(fmt.Sprintf("- **%s** - *%s* - %s - `ERROR` - %s\n", version, property, propertyResult.Name, err))
					}
				}

				if len(propertyResult.Warnings) > 0 {
					for _, err := range propertyResult.Warnings {
						out.WriteString(fmt.Sprintf("- **%s** - *%s* - %s - `WARNING` - %s\n", version, property, propertyResult.Name, err))
					}
				}

				if len(propertyResult.Errors) == 0 && len(propertyResult.Warnings) == 0 {
					out.WriteString(fmt.Sprintf("- **%s** - *%s* - %s - ✓\n", version, property, propertyResult.Name))
				}
			}
		}
	}

	return out.String()
}

// RenderPlainText returns a string of the results rendered as PlainText
//
//nolint:dupl
func (rr *Results) RenderPlainText() string { //nolint:gocognit,cyclop
	var out strings.Builder

	out.WriteString("CRD Validations\n")

	for _, result := range rr.CRDValidation {
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				out.WriteString(fmt.Sprintf("- %s - ERROR - %s\n", result.Name, err))
			}
		}

		if len(result.Warnings) > 0 {
			for _, err := range result.Warnings {
				out.WriteString(fmt.Sprintf("- %s - WARNING - %s\n", result.Name, err))
			}
		}

		if len(result.Errors) == 0 && len(result.Warnings) == 0 {
			out.WriteString(fmt.Sprintf("- %s - ✓\n", result.Name))
		}
	}

	out.WriteString("\n\n")
	out.WriteString("Same Version Validations\n")

	for version, result := range rr.SameVersionValidation {
		for property, results := range result {
			for _, propertyResult := range results {
				if len(propertyResult.Errors) > 0 {
					for _, err := range propertyResult.Errors {
						out.WriteString(fmt.Sprintf("- %s - %s - %s - ERROR - %s\n", version, property, propertyResult.Name, err))
					}
				}

				if len(propertyResult.Warnings) > 0 {
					for _, err := range propertyResult.Warnings {
						out.WriteString(fmt.Sprintf("- %s - %s - %s - WARNING - %s\n", version, property, propertyResult.Name, err))
					}
				}

				if len(propertyResult.Errors) == 0 && len(propertyResult.Warnings) == 0 {
					out.WriteString(fmt.Sprintf("- %s - %s - %s - ✓\n", version, property, propertyResult.Name))
				}
			}
		}
	}

	out.WriteString("\n\n")
	out.WriteString("Served Version Validations\n")

	for version, result := range rr.ServedVersionValidation {
		for property, results := range result {
			for _, propertyResult := range results {
				if len(propertyResult.Errors) > 0 {
					for _, err := range propertyResult.Errors {
						out.WriteString(fmt.Sprintf("- %s - %s - %s - ERROR - %s\n", version, property, propertyResult.Name, err))
					}
				}

				if len(propertyResult.Warnings) > 0 {
					for _, err := range propertyResult.Warnings {
						out.WriteString(fmt.Sprintf("- %s - %s - %s - WARNING - %s\n", version, property, propertyResult.Name, err))
					}
				}

				if len(propertyResult.Errors) == 0 && len(propertyResult.Warnings) == 0 {
					out.WriteString(fmt.Sprintf("- %s - %s - %s - ✓\n", version, property, propertyResult.Name))
				}
			}
		}
	}

	return out.String()
}

// HasFailures returns a boolean signaling if any of the validation results contain any errors.
func (rr *Results) HasFailures() bool {
	return rr.HasCRDValidationFailures() || rr.HasSameVersionValidationFailures() || rr.HasServedVersionValidationFailures()
}

// HasCRDValidationFailures returns a boolean signaling if the CRD scoped validations contain any errors.
func (rr *Results) HasCRDValidationFailures() bool {
	for _, result := range rr.CRDValidation {
		if len(result.Errors) > 0 {
			return true
		}
	}

	return false
}

// HasSameVersionValidationFailures returns a boolean signaling if the same version validations contain any errors.
func (rr *Results) HasSameVersionValidationFailures() bool {
	for _, versionResults := range rr.SameVersionValidation {
		for _, propertyResults := range versionResults {
			for _, result := range propertyResults {
				if len(result.Errors) > 0 {
					return true
				}
			}
		}
	}

	return false
}

// HasServedVersionValidationFailures returns a boolean signaling if the served version validations contain any errors.
func (rr *Results) HasServedVersionValidationFailures() bool {
	for _, versionResults := range rr.ServedVersionValidation {
		for _, propertyResults := range versionResults {
			for _, result := range propertyResults {
				if len(result.Errors) > 0 {
					return true
				}
			}
		}
	}

	return false
}
