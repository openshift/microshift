package brokertemplateinstance

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	templateapi "github.com/openshift/openshift-apiserver/pkg/template/apis/template"
	"github.com/openshift/openshift-apiserver/pkg/template/apis/template/validation"
)

// brokerTemplateInstanceStrategy implements behavior for BrokerTemplateInstances
type brokerTemplateInstanceStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating BrokerTemplateInstance
// objects via the REST API.
var Strategy = brokerTemplateInstanceStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped is false for brokertemplateinstances.
func (brokerTemplateInstanceStrategy) NamespaceScoped() bool {
	return false
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (brokerTemplateInstanceStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

// Canonicalize normalizes the object after validation.
func (brokerTemplateInstanceStrategy) Canonicalize(obj runtime.Object) {
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (brokerTemplateInstanceStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

// Validate validates a new brokertemplateinstance.
func (brokerTemplateInstanceStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateBrokerTemplateInstance(obj.(*templateapi.BrokerTemplateInstance))
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (brokerTemplateInstanceStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// AllowCreateOnUpdate is false for brokertemplateinstances.
func (brokerTemplateInstanceStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (brokerTemplateInstanceStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user.
func (brokerTemplateInstanceStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateBrokerTemplateInstanceUpdate(obj.(*templateapi.BrokerTemplateInstance), old.(*templateapi.BrokerTemplateInstance))
}

// WarningsOnUpdate returns warnings for the given update.
func (brokerTemplateInstanceStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
