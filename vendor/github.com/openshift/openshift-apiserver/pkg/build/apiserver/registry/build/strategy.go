package build

import (
	"context"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	kapi "k8s.io/kubernetes/pkg/apis/core"

	buildapi "github.com/openshift/openshift-apiserver/pkg/build/apis/build"
	buildinternalhelpers "github.com/openshift/openshift-apiserver/pkg/build/apis/build/internal_helpers"
	"github.com/openshift/openshift-apiserver/pkg/build/apis/build/validation"
)

// strategy implements behavior for Build objects
type strategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Build objects.
var Strategy = strategy{legacyscheme.Scheme, names.SimpleNameGenerator}

func (strategy) NamespaceScoped() bool {
	return true
}

// AllowCreateOnUpdate is false for Build objects.
func (strategy) AllowCreateOnUpdate() bool {
	return false
}

func (strategy) AllowUnconditionalUpdate() bool {
	return false
}

// manageConditions updates the build argument to make the conditions array match the current
// build phase/reason/message information in the build object.  It will set existing conditions
// to false other than the condition representing the current phase, and it will add a new
// condition representing the current phase if no such condition exists.
func manageConditions(build *buildapi.Build) {
	// if we have no phase information, we can't reasonably update the conditions
	if len(build.Status.Phase) == 0 {
		return
	}
	now := metav1.Now()
	found := false
	for i, c := range build.Status.Conditions {
		if buildapi.BuildPhase(c.Type) == build.Status.Phase {
			found = true
			if c.Status != kapi.ConditionTrue || c.Reason != string(build.Status.Reason) || c.Message != build.Status.Message {
				if c.Status != kapi.ConditionTrue {
					build.Status.Conditions[i].Status = kapi.ConditionTrue
					build.Status.Conditions[i].LastTransitionTime = now
				}
				build.Status.Conditions[i].LastUpdateTime = now
				build.Status.Conditions[i].Reason = string(build.Status.Reason)
				build.Status.Conditions[i].Message = build.Status.Message
			}
		} else {
			if c.Status != kapi.ConditionFalse {
				build.Status.Conditions[i].Status = kapi.ConditionFalse
				build.Status.Conditions[i].LastTransitionTime = now
				build.Status.Conditions[i].LastUpdateTime = now
				build.Status.Conditions[i].Reason = ""
				build.Status.Conditions[i].Message = ""
			}
		}
	}
	if !found {
		condition := buildapi.BuildCondition{
			Type:               buildapi.BuildConditionType(build.Status.Phase),
			Status:             kapi.ConditionTrue,
			LastUpdateTime:     now,
			LastTransitionTime: now,
			Reason:             string(build.Status.Reason),
			Message:            build.Status.Message,
		}
		build.Status.Conditions = append(build.Status.Conditions, condition)
	}
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	build := obj.(*buildapi.Build)
	build.Generation = 1
	if len(build.Status.Phase) == 0 {
		build.Status.Phase = buildapi.BuildPhaseNew
	}
	manageConditions(build)
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (strategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newBuild := obj.(*buildapi.Build)
	oldBuild := old.(*buildapi.Build)
	// If the build is already in a failed state, do not allow an update
	// of the reason and message. This is to prevent the build controller from
	// overwriting the reason and message that was set by the builder pod
	// when it updated the build's details.
	// Only allow OOMKilled override because various processes in a container
	// can get OOMKilled and this confuses builder to prematurely populate
	// failure reason
	if oldBuild.Status.Phase == buildapi.BuildPhaseFailed &&
		newBuild.Status.Reason != buildapi.StatusReasonOutOfMemoryKilled {
		newBuild.Status.Reason = oldBuild.Status.Reason
		newBuild.Status.Message = oldBuild.Status.Message
	}
	manageConditions(newBuild)

	if !reflect.DeepEqual(oldBuild.Spec, newBuild.Spec) {
		newBuild.Generation = oldBuild.Generation + 1
	}
}

// Canonicalize normalizes the object after validation.
func (strategy) Canonicalize(obj runtime.Object) {
}

// Validate validates a new policy.
func (strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateBuild(obj.(*buildapi.Build))
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (strategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// ValidateUpdate is the default update validation for an end user.
func (strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateBuildUpdate(obj.(*buildapi.Build), old.(*buildapi.Build))
}

// WarningsOnUpdate returns warnings for the given update.
func (strategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// CheckGracefulDelete allows a build to be gracefully deleted.
func (strategy) CheckGracefulDelete(obj runtime.Object, options *metav1.DeleteOptions) bool {
	return false
}

type detailsStrategy struct {
	strategy
}

// Prepares a build for update by only allowing an update to build details.
// Build details currently consists of: Spec.Revision, Status.Reason, and
// Status.Message, all of which are updated from within the build pod
func (detailsStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newBuild := obj.(*buildapi.Build)
	oldBuild := old.(*buildapi.Build)

	// ignore phase updates unless the caller is updating the build to
	// a completed phase.
	phase := oldBuild.Status.Phase
	stages := newBuild.Status.Stages
	if buildinternalhelpers.IsBuildComplete(newBuild) {
		phase = newBuild.Status.Phase
	}
	revision := newBuild.Spec.Revision
	message := newBuild.Status.Message
	reason := newBuild.Status.Reason
	outputTo := newBuild.Status.Output.To
	*newBuild = *oldBuild
	newBuild.Status.Phase = phase
	newBuild.Status.Stages = stages
	newBuild.Spec.Revision = revision
	newBuild.Status.Reason = reason
	newBuild.Status.Message = message
	newBuild.Status.Output.To = outputTo
	manageConditions(newBuild)
}

// Validates that an update is valid by ensuring that no Revision exists and that it's not getting updated to blank
func (detailsStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newBuild := obj.(*buildapi.Build)
	oldBuild := old.(*buildapi.Build)
	oldRevision := oldBuild.Spec.Revision
	newRevision := newBuild.Spec.Revision
	errors := field.ErrorList{}

	if newRevision == nil && oldRevision != nil {
		errors = append(errors, field.Invalid(field.NewPath("spec", "revision"), nil, "cannot set an empty revision in build spec"))
	}
	if !reflect.DeepEqual(oldRevision, newRevision) && oldRevision != nil {
		// If there was already a revision, then return an error
		errors = append(errors, field.Duplicate(field.NewPath("spec", "revision"), oldBuild.Spec.Revision))
	}
	return errors
}

// AllowUnconditionalUpdate returns true to allow a Build with an empty resourceVersion to update the Revision
func (detailsStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// DetailsStrategy is the strategy used to manage updates to a Build revision
var DetailsStrategy = detailsStrategy{Strategy}
