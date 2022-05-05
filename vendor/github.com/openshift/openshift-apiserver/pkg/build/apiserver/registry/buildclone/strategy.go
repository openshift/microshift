package buildclone

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	buildapi "github.com/openshift/openshift-apiserver/pkg/build/apis/build"
	buildvalidation "github.com/openshift/openshift-apiserver/pkg/build/apis/build/validation"
)

type strategy struct {
	runtime.ObjectTyper
}

var Strategy = strategy{legacyscheme.Scheme}

func (s strategy) NamespaceScoped() bool {
	return true
}

func (s strategy) AllowCreateOnUpdate() bool {
	return false
}

func (s strategy) GenerateName(base string) string {
	return base
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (s strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

// Validate validates a new role.
func (s strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return buildvalidation.ValidateBuildRequest(obj.(*buildapi.BuildRequest))
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (strategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// Canonicalize normalizes the object after validation.
func (strategy) Canonicalize(obj runtime.Object) {
}
