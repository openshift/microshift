package project

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	kapi "k8s.io/kubernetes/pkg/apis/core"

	projectapi "github.com/openshift/openshift-apiserver/pkg/project/apis/project"
	"github.com/openshift/openshift-apiserver/pkg/project/apis/project/validation"
)

// projectStrategy implements behavior for projects
type projectStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Project
// objects via the REST API.
var Strategy = projectStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped is false for projects.
func (projectStrategy) NamespaceScoped() bool {
	return false
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (projectStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	project := obj.(*projectapi.Project)
	hasProjectFinalizer := false
	for i := range project.Spec.Finalizers {
		if project.Spec.Finalizers[i] == projectapi.FinalizerOrigin {
			hasProjectFinalizer = true
			break
		}
	}
	if !hasProjectFinalizer {
		if len(project.Spec.Finalizers) == 0 {
			project.Spec.Finalizers = []kapi.FinalizerName{projectapi.FinalizerOrigin}
		} else {
			project.Spec.Finalizers = append(project.Spec.Finalizers, projectapi.FinalizerOrigin)
		}
	}
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (projectStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newProject := obj.(*projectapi.Project)
	oldProject := old.(*projectapi.Project)
	newProject.Spec.Finalizers = oldProject.Spec.Finalizers
	newProject.Status = oldProject.Status
}

// Validate validates a new project.
func (projectStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateProject(obj.(*projectapi.Project))
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (projectStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// AllowCreateOnUpdate is false for project.
func (projectStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (projectStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// Canonicalize normalizes the object after validation.
func (projectStrategy) Canonicalize(obj runtime.Object) {
}

// ValidateUpdate is the default update validation for an end user.
func (projectStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateProjectUpdate(obj.(*projectapi.Project), old.(*projectapi.Project))
}

// WarningsOnUpdate returns warnings for the given update.
func (projectStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
