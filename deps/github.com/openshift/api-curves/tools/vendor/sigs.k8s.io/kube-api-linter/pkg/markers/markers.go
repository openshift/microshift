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
package markers

const (
	// OptionalMarker is the marker that indicates that a field is optional.
	OptionalMarker = "optional"

	// RequiredMarker is the marker that indicates that a field is required.
	RequiredMarker = "required"

	// NullableMarker is the marker that indicates that a field can be null.
	NullableMarker = "nullable"

	// DefaultMarker is the marker that specifies the default value of a field or type.
	DefaultMarker = "default"
)

const (
	// KubebuilderRootMarker is the marker that indicates that a struct is the object root for code and CRD generation.
	KubebuilderRootMarker = "kubebuilder:object:root"

	// KubebuilderStatusSubresourceMarker is the marker that indicates that the CRD generated for a struct should include the /status subresource.
	KubebuilderStatusSubresourceMarker = "kubebuilder:subresource:status"

	// KubebuilderAtLeastOneOfMarker is the marker that indicates that a type has a CEL validation in kubebuilder enforcing that at least one field is set.
	KubebuilderAtLeastOneOfMarker = "kubebuilder:validation:AtLeastOneOf"

	// KubebuilderEnumMarker is the marker that indicates that a field has an enum in kubebuilder.
	KubebuilderEnumMarker = "kubebuilder:validation:Enum"

	// KubebuilderFormatMarker is the marker that indicates that a field has a format in kubebuilder.
	KubebuilderFormatMarker = "kubebuilder:validation:Format"

	// KubebuilderMaximumMarker is the marker that indicates that a field has a maximum value in kubebuilder.
	KubebuilderMaximumMarker = "kubebuilder:validation:Maximum"

	// KubebuilderMaxItemsMarker is the marker that indicates that a field has a maximum number of items in kubebuilder.
	KubebuilderMaxItemsMarker = "kubebuilder:validation:MaxItems"

	// KubebuilderMaxLengthMarker is the marker that indicates that a field has a maximum length in kubebuilder.
	KubebuilderMaxLengthMarker = "kubebuilder:validation:MaxLength"

	// KubebuilderMaxPropertiesMarker is the marker that indicates that a field has a maximum number of properties in kubebuilder.
	KubebuilderMaxPropertiesMarker = "kubebuilder:validation:MaxProperties"

	// KubebuilderMinimumMarker is the marker that indicates that a field has a minimum value in kubebuilder.
	KubebuilderMinimumMarker = "kubebuilder:validation:Minimum"

	// KubebuilderMinItemsMarker is the marker that indicates that a field has a minimum number of items in kubebuilder.
	KubebuilderMinItemsMarker = "kubebuilder:validation:MinItems"

	// KubebuilderMinLengthMarker is the marker that indicates that a field has a minimum length in kubebuilder.
	KubebuilderMinLengthMarker = "kubebuilder:validation:MinLength"

	// KubebuilderMinPropertiesMarker is the marker that indicates that a field has a minimum number of properties in kubebuilder.
	KubebuilderMinPropertiesMarker = "kubebuilder:validation:MinProperties"

	// KubebuilderItemsMaxItemsMarker is the marker that indicates that a nested array field has a maximum number of items in kubebuilder.
	KubebuilderItemsMaxItemsMarker = "kubebuilder:validation:items:MaxItems"

	// KubebuilderItemsMaxPropertiesMarker is the marker that indicates that array items have a maximum number of properties in kubebuilder.
	KubebuilderItemsMaxPropertiesMarker = "kubebuilder:validation:items:MaxProperties"

	// KubebuilderItemsMaximumMarker is the marker that indicates that array items have a maximum in kubebuilder.
	KubebuilderItemsMaximumMarker = "kubebuilder:validation:items:Maximum"

	// KubebuilderItemsMinItemsMarker is the marker that indicates that a nested array field has a minimum number of items in kubebuilder.
	KubebuilderItemsMinItemsMarker = "kubebuilder:validation:items:MinItems"

	// KubebuilderItemsMinLengthMarker is the marker that indicates that array items have a minimum length in kubebuilder.
	KubebuilderItemsMinLengthMarker = "kubebuilder:validation:items:MinLength"

	// KubebuilderItemsMinPropertiesMarker is the marker that indicates that array items have a minimum number of properties in kubebuilder.
	KubebuilderItemsMinPropertiesMarker = "kubebuilder:validation:items:MinProperties"

	// KubebuilderItemsMinimumMarker is the marker that indicates that array items have a minimum in kubebuilder.
	KubebuilderItemsMinimumMarker = "kubebuilder:validation:items:Minimum"

	// KubebuilderOptionalMarker is the marker that indicates that a field is optional in kubebuilder.
	KubebuilderOptionalMarker = "kubebuilder:validation:Optional"

	// KubebuilderRequiredMarker is the marker that indicates that a field is required in kubebuilder.
	KubebuilderRequiredMarker = "kubebuilder:validation:Required"

	// KubebuilderExactlyOneOf is the marker that indicates that a type has a CEL validation in kubebuilder enforcing that exactly one field is set.
	KubebuilderExactlyOneOf = "kubebuilder:validation:ExactlyOneOf"

	// KubebuilderItemsMaxLengthMarker is the marker that indicates that a field has a maximum length in kubebuilder.
	KubebuilderItemsMaxLengthMarker = "kubebuilder:validation:items:MaxLength"

	// KubebuilderItemsEnumMarker is the marker that indicates that a field has an enum in kubebuilder.
	KubebuilderItemsEnumMarker = "kubebuilder:validation:items:Enum"

	// KubebuilderItemsFormatMarker is the marker that indicates that a field has a format in kubebuilder.
	KubebuilderItemsFormatMarker = "kubebuilder:validation:items:Format"

	// KubebuilderDefaultMarker is the marker used to specify the default value for a type or field in kubebuilder.
	KubebuilderDefaultMarker = "kubebuilder:default"

	// KubebuilderExampleMarker is the marker used to specify an example value for the type or field in kubebuilder.
	KubebuilderExampleMarker = "kubebuilder:example"

	// KubebuilderExclusiveMaximumMarker is the marker used to specify that the maximum value is excluded from the allowed values (i.e "up to, but not including") for a type or field in kubebuilder.
	KubebuilderExclusiveMaximumMarker = "kubebuilder:validation:ExclusiveMaximum"

	// KubebuilderExclusiveMinimumMarker is the marker used to specify that the minimum value is excluded from the allowed values (i.e "up to, but not including") for a type or field in kubebuilder.
	KubebuilderExclusiveMinimumMarker = "kubebuilder:validation:ExclusiveMinimum"

	// KubebuilderMultipleOfMarker is the marker used to specify that the value for a type or field must be a multiple of X in kubebuilder.
	KubebuilderMultipleOfMarker = "kubebuilder:validation:MultipleOf"

	// KubebuilderPatternMarker is the marker used to specify that the value for a type or field must follow a particular regex pattern in kubebuilder.
	KubebuilderPatternMarker = "kubebuilder:validation:Pattern"

	// KubebuilderTypeMarker is the marker used to specify the type a value should be for a type or field in kubebuilder.
	KubebuilderTypeMarker = "kubebuilder:validation:Type"

	// KubebuilderUniqueItemsMarker is the marker used to specify that a type or field must contain unique items in kubebuilder.
	KubebuilderUniqueItemsMarker = "kubebuilder:validation:UniqueItems"

	// KubebuilderItemsExclusiveMaximumMarker is the marker used to specify that the maximum value for an array item is excluded from the allowed values (i.e "up to, but not including") for a type or field in kubebuilder.
	KubebuilderItemsExclusiveMaximumMarker = "kubebuilder:validation:items:ExclusiveMaximum"

	// KubebuilderItemsExclusiveMinimumMarker is the marker used to specify that the minimum value for an array item is excluded from the allowed values (i.e "up to, but not including") for a type or field in kubebuilder.
	KubebuilderItemsExclusiveMinimumMarker = "kubebuilder:validation:items:ExclusiveMinimum"

	// KubebuilderItemsMultipleOfMarker is the marker used to specify that the value of an array item for the type or field must be a multiple of X in kubebuilder.
	KubebuilderItemsMultipleOfMarker = "kubebuilder:validation:items:MultipleOf"

	// KubebuilderItemsPatternMarker is the marker used to specify that the value of an array item for the type or field must follow a particular regex pattern in kubebuilder.
	KubebuilderItemsPatternMarker = "kubebuilder:validation:items:Pattern"

	// KubebuilderItemsTypeMarker is the marker used to specify the type an array item should be for the type or field in kubebuilder.
	KubebuilderItemsTypeMarker = "kubebuilder:validation:items:Type"

	// KubebuilderItemsUniqueItemsMarker is the marker used to specify that entries to a nested array type or field must contain unique items in kubebuilder.
	KubebuilderItemsUniqueItemsMarker = "kubebuilder:validation:items:UniqueItems"

	// KubebuilderXValidationMarker is the marker used to specify CEL validation rules for a type or field in kubebuilder.
	KubebuilderXValidationMarker = "kubebuilder:validation:XValidation"

	// KubebuilderItemsXValidationMarker is the marker used to specify CEL validation rules for entries to a nested array type or field in kubebuilder.
	KubebuilderItemsXValidationMarker = "kubebuilder:validation:items:XValidation"

	// KubebuilderListTypeMarker is the marker used to specify the type of list for server-side apply operations.
	KubebuilderListTypeMarker = "listType"

	// KubebuilderListMapKeyMarker is the marker used to specify the key field for map-type lists.
	KubebuilderListMapKeyMarker = "listMapKey"

	// KubebuilderSchemaLessMarker is the marker that indicates that a struct is schemaless.
	KubebuilderSchemaLessMarker = "kubebuilder:validation:Schemaless"
)

const (
	// K8sOptionalMarker is the marker that indicates that a field is optional in k8s declarative validation.
	K8sOptionalMarker = "k8s:optional"

	// K8sRequiredMarker is the marker that indicates that a field is required in k8s declarative validation.
	K8sRequiredMarker = "k8s:required"

	// K8sFormatMarker is the marker that indicates that a field has a format in k8s declarative validation.
	K8sFormatMarker = "k8s:format"

	// K8sMinLengthMarker is the marker that indicates that a field has a minimum length in k8s declarative validation.
	K8sMinLengthMarker = "k8s:minLength"

	// K8sMaxLengthMarker is the marker that indicates that a field has a maximum length in k8s declarative validation.
	K8sMaxLengthMarker = "k8s:maxLength"

	// K8sMinItemsMarker is the marker that indicates that a field has a minimum number of items in k8s declarative validation.
	K8sMinItemsMarker = "k8s:minItems"

	// K8sMaxItemsMarker is the marker that indicates that a field has a maximum number of items in k8s declarative validation.
	K8sMaxItemsMarker = "k8s:maxItems"

	// K8sEnumMarker is the marker that indicates that a field has an enum in k8s declarative validation.
	K8sEnumMarker = "k8s:enum"

	// K8sMinimumMarker is the marker that indicates that a field has a minimum value in k8s declarative validation.
	K8sMinimumMarker = "k8s:minimum"

	// K8sMaximumMarker is the marker that indicates that a field has a maximum value in k8s declarative validation.
	K8sMaximumMarker = "k8s:maximum"

	// K8sExclusiveMaximumMarker is the marker that indicates that a field has an exclusive maximum value in k8s declarative validation.
	K8sExclusiveMaximumMarker = "k8s:exclusiveMaximum"

	// K8sExclusiveMinimumMarker is the marker that indicates that a field has an exclusive minimum value in k8s declarative validation.
	K8sExclusiveMinimumMarker = "k8s:exclusiveMinimum"

	// K8sListTypeMarker is the marker that indicates that a field is a list in k8s declarative validation.
	K8sListTypeMarker = "k8s:listType"

	// K8sListMapKeyMarker is the marker that indicates that a field is a map in k8s declarative validation.
	K8sListMapKeyMarker = "k8s:listMapKey"

	// K8sDefaultMarker is the marker that indicates the default value for a field in k8s declarative validation.
	K8sDefaultMarker = "k8s:default"
)
