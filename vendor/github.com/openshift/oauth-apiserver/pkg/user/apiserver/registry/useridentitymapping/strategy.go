package useridentitymapping

import (
	"context"

	"github.com/openshift/oauth-apiserver/pkg/serverscheme"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"

	userapi "github.com/openshift/oauth-apiserver/pkg/user/apis/user"
	"github.com/openshift/oauth-apiserver/pkg/user/apis/user/validation"
)

// userIdentityMappingStrategy implements behavior for image repository mappings.
type userIdentityMappingStrategy struct {
	runtime.ObjectTyper
}

// Strategy is the default logic that applies when creating UserIdentityMapping
// objects via the REST API.
var Strategy = userIdentityMappingStrategy{serverscheme.Scheme}

// NamespaceScoped is true for image repository mappings.
func (s userIdentityMappingStrategy) NamespaceScoped() bool {
	return false
}

func (userIdentityMappingStrategy) GenerateName(base string) string {
	return base
}

func (userIdentityMappingStrategy) AllowCreateOnUpdate() bool {
	return true
}

func (userIdentityMappingStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (s userIdentityMappingStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	mapping := obj.(*userapi.UserIdentityMapping)

	if len(mapping.Name) == 0 {
		mapping.Name = mapping.Identity.Name
	}
	mapping.Namespace = ""
	mapping.ResourceVersion = ""

	mapping.Identity.Namespace = ""
	mapping.Identity.Kind = ""
	mapping.Identity.UID = ""

	mapping.User.Namespace = ""
	mapping.User.Kind = ""
	mapping.User.UID = ""
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update
func (s userIdentityMappingStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	mapping := obj.(*userapi.UserIdentityMapping)

	if len(mapping.Name) == 0 {
		mapping.Name = mapping.Identity.Name
	}
	mapping.Namespace = ""

	mapping.Identity.Namespace = ""
	mapping.Identity.Kind = ""
	mapping.Identity.UID = ""

	mapping.User.Namespace = ""
	mapping.User.Kind = ""
	mapping.User.UID = ""
}

// Canonicalize normalizes the object after validation.
func (s userIdentityMappingStrategy) Canonicalize(obj runtime.Object) {
}

// Validate validates a new UserIdentityMapping.
func (s userIdentityMappingStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateUserIdentityMapping(obj.(*userapi.UserIdentityMapping))
}

// Validate validates an updated UserIdentityMapping.
func (s userIdentityMappingStrategy) ValidateUpdate(ctx context.Context, obj runtime.Object, old runtime.Object) field.ErrorList {
	return validation.ValidateUserIdentityMappingUpdate(obj.(*userapi.UserIdentityMapping), old.(*userapi.UserIdentityMapping))
}
