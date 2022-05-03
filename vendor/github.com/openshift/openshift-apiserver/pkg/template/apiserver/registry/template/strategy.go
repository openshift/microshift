package template

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	templateapi "github.com/openshift/openshift-apiserver/pkg/template/apis/template"
	"github.com/openshift/openshift-apiserver/pkg/template/apis/template/validation"
)

// templateStrategy implements behavior for Templates
type templateStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Template
// objects via the REST API.
var Strategy = templateStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped is true for templates.
func (templateStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (templateStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {}

// Canonicalize normalizes the object after validation.
func (templateStrategy) Canonicalize(obj runtime.Object) {
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (templateStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

// Validate validates a new template.
func (templateStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateTemplate(obj.(*templateapi.Template))
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (templateStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// AllowCreateOnUpdate is false for templates.
func (templateStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (templateStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user.
func (templateStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateTemplateUpdate(obj.(*templateapi.Template), old.(*templateapi.Template))
}

// WarningsOnUpdate returns warnings for the given update.
func (templateStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
