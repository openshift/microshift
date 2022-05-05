package buildconfig

import (
	"context"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	buildapi "github.com/openshift/openshift-apiserver/pkg/build/apis/build"
	"github.com/openshift/openshift-apiserver/pkg/build/apis/build/validation"
)

var (
	// GroupStrategy is the logic that applies when creating and updating BuildConfig objects
	// in the Group api.
	// This differs from the LegacyStrategy in that on create it will set a default build
	// pruning limit value for both successful and failed builds.  This is new behavior that
	// can only be introduced to users consuming the new group based api.
	GroupStrategy = groupStrategy{strategy{legacyscheme.Scheme, names.SimpleNameGenerator}}

	// LegacyStrategy is the default logic that applies when creating BuildConfig objects.
	// Specifically it will not set the default build pruning limit values because that was not
	// part of the legacy api.
	LegacyStrategy = legacyStrategy{strategy{legacyscheme.Scheme, names.SimpleNameGenerator}}
)

// strategy implements most of the behavior for BuildConfig objects
// It does not provide a PrepareForCreate implementation, that is expected
// to be implemented by the child implementation.
type strategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (strategy) NamespaceScoped() bool {
	return true
}

// AllowCreateOnUpdate is false for BuildConfig objects.
func (strategy) AllowCreateOnUpdate() bool {
	return false
}

func (strategy) AllowUnconditionalUpdate() bool {
	return false
}

// Canonicalize normalizes the object after validation.
func (strategy) Canonicalize(obj runtime.Object) {
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
// This is invoked by the Group and Legacy strategies.
func (s strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	bc := obj.(*buildapi.BuildConfig)
	bc.Generation = 1
	dropUnknownTriggers(bc)
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (strategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newBC := obj.(*buildapi.BuildConfig)
	oldBC := old.(*buildapi.BuildConfig)
	dropUnknownTriggers(newBC)
	// Do not allow the build version to go backwards or we'll
	// get conflicts with existing builds.
	if newBC.Status.LastVersion < oldBC.Status.LastVersion {
		newBC.Status.LastVersion = oldBC.Status.LastVersion
	}

	if !reflect.DeepEqual(oldBC.Spec, newBC.Spec) || newBC.Status.LastVersion != oldBC.Status.LastVersion {
		newBC.Generation = oldBC.Generation + 1
	}
}

// Validate validates a new policy.
func (strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateBuildConfig(obj.(*buildapi.BuildConfig))
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (strategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// ValidateUpdate is the default update validation for an end user.
func (strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateBuildConfigUpdate(obj.(*buildapi.BuildConfig), old.(*buildapi.BuildConfig))
}

// WarningsOnUpdate returns warnings for the given update.
func (strategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// groupStrategy implements behavior for BuildConfig objects in the Group api
type groupStrategy struct {
	strategy
}

// PrepareForCreate delegates to the common strategy and sets default pruning limits
func (s groupStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	s.strategy.PrepareForCreate(ctx, obj)

	bc := obj.(*buildapi.BuildConfig)
	if bc.Spec.SuccessfulBuildsHistoryLimit == nil {
		v := buildapi.DefaultSuccessfulBuildsHistoryLimit
		bc.Spec.SuccessfulBuildsHistoryLimit = &v
	}
	if bc.Spec.FailedBuildsHistoryLimit == nil {
		v := buildapi.DefaultFailedBuildsHistoryLimit
		bc.Spec.FailedBuildsHistoryLimit = &v
	}
}

// legacyStrategy implements behavior for BuildConfig objects in the legacy api
type legacyStrategy struct {
	strategy
}

// PrepareForCreate delegates to the common strategy.
func (s legacyStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	s.strategy.PrepareForCreate(ctx, obj)

	// legacy buildconfig api does not apply default pruning values, to maintain
	// backwards compatibility.

}

var _ rest.GarbageCollectionDeleteStrategy = legacyStrategy{}

// DefaultGarbageCollectionPolicy for legacy buildconfigs will orphan dependents.
func (s legacyStrategy) DefaultGarbageCollectionPolicy(ctx context.Context) rest.GarbageCollectionPolicy {
	return rest.OrphanDependents
}

// CheckGracefulDelete allows a build config to be gracefully deleted.
func (strategy) CheckGracefulDelete(obj runtime.Object, options *metav1.DeleteOptions) bool {
	return false
}

// dropUnknownTriggers drops any triggers that are of an unknown type
func dropUnknownTriggers(bc *buildapi.BuildConfig) {
	triggers := []buildapi.BuildTriggerPolicy{}
	for _, t := range bc.Spec.Triggers {
		if buildapi.KnownTriggerTypes.Has(string(t.Type)) {
			triggers = append(triggers, t)
		}
	}
	bc.Spec.Triggers = triggers
}
