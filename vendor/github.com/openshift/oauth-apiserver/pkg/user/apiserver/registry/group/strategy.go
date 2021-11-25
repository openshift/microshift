package group

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/openshift/oauth-apiserver/pkg/serverscheme"
	userapi "github.com/openshift/oauth-apiserver/pkg/user/apis/user"
	"github.com/openshift/oauth-apiserver/pkg/user/apis/user/validation"
)

// groupStrategy implements behavior for Groups
type groupStrategy struct {
	runtime.ObjectTyper
}

// Strategy is the default logic that applies when creating and updating Group
// objects via the REST API.
var Strategy = groupStrategy{serverscheme.Scheme}

var _ rest.GarbageCollectionDeleteStrategy = groupStrategy{}

func (groupStrategy) DefaultGarbageCollectionPolicy(ctx context.Context) rest.GarbageCollectionPolicy {
	return rest.Unsupported
}

func (groupStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {}

// NamespaceScoped is false for groups
func (groupStrategy) NamespaceScoped() bool {
	return false
}

func (groupStrategy) GenerateName(base string) string {
	return base
}

func (groupStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

// Validate validates a new group
func (groupStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateGroup(obj.(*userapi.Group))
}

// AllowCreateOnUpdate is false for groups
func (groupStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (groupStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// Canonicalize normalizes the object after validation.
func (groupStrategy) Canonicalize(obj runtime.Object) {
}

// ValidateUpdate is the default update validation for an end group.
func (groupStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateGroupUpdate(obj.(*userapi.Group), old.(*userapi.Group))
}
