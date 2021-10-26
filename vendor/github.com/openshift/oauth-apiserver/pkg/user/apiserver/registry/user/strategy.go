package user

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/openshift/oauth-apiserver/pkg/serverscheme"
	userapi "github.com/openshift/oauth-apiserver/pkg/user/apis/user"
	"github.com/openshift/oauth-apiserver/pkg/user/apis/user/validation"
)

// userStrategy implements behavior for Users
type userStrategy struct {
	runtime.ObjectTyper
}

// Strategy is the default logic that applies when creating and updating User
// objects via the REST API.
var Strategy = userStrategy{serverscheme.Scheme}

var _ rest.GarbageCollectionDeleteStrategy = userStrategy{}

func (userStrategy) DefaultGarbageCollectionPolicy(ctx context.Context) rest.GarbageCollectionPolicy {
	return rest.Unsupported
}

func (userStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	// persisting this field is no longer allowed
	// it solely exists to echo back the user's current groups via the ~ virtual user
	obj.(*userapi.User).Groups = nil
}

// NamespaceScoped is false for users
func (userStrategy) NamespaceScoped() bool {
	return false
}

func (userStrategy) GenerateName(base string) string {
	return base
}

func (userStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	// persisting this field is no longer allowed
	// it solely exists to echo back the user's current groups via the ~ virtual user
	obj.(*userapi.User).Groups = nil
}

// Validate validates a new user
func (userStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateUser(obj.(*userapi.User))
}

// AllowCreateOnUpdate is false for users
func (userStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (userStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// Canonicalize normalizes the object after validation.
func (userStrategy) Canonicalize(obj runtime.Object) {
}

// ValidateUpdate is the default update validation for an end user.
func (userStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateUserUpdate(obj.(*userapi.User), old.(*userapi.User))
}
