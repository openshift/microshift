package validation

import (
	"fmt"
	"strings"

	kvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/api/validation/path"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/openshift/apiserver-library-go/pkg/apivalidation"
	userapi "github.com/openshift/oauth-apiserver/pkg/user/apis/user"
)

func ValidateIdentityName(name string, _ bool) []string {
	if reasons := path.ValidatePathSegmentName(name, false); len(reasons) != 0 {
		return reasons
	}

	parts := strings.Split(name, ":")
	usernameIndex := 1
	// allow identities in the form "<providerName>:b64:<base64-encoded username>"
	if len(parts) == 3 && parts[1] == "b64" {
		usernameIndex++
	} else if len(parts) != 2 {
		return []string{`must be in the format <providerName>:<providerUserName>`}
	}
	if len(parts[0]) == 0 {
		return []string{`must be in the format <providerName>:<providerUserName> with a non-empty providerName`}
	}
	if len(parts[usernameIndex]) == 0 {
		return []string{`must be in the format <providerName>:<providerUserName> with a non-empty providerUserName`}
	}
	return nil
}

// if you change this, update the peer in oauth admission validation.  also, don't change this.
func ValidateIdentityProviderName(name string) []string {
	if reasons := path.ValidatePathSegmentName(name, false); len(reasons) != 0 {
		return reasons
	}

	if strings.Contains(name, ":") {
		return []string{`may not contain ":"`}
	}
	return nil
}

func ValidateIdentityProviderUserName(name string) []string {
	// Any provider user name must be a valid user name
	return apivalidation.ValidateUserName(name, false)
}

func ValidateGroup(group *userapi.Group) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMeta(&group.ObjectMeta, false, apivalidation.ValidateGroupName, field.NewPath("metadata"))

	userPath := field.NewPath("user")
	for index, user := range group.Users {
		idxPath := userPath.Index(index)
		if len(user) == 0 {
			allErrs = append(allErrs, field.Invalid(idxPath, user, "may not be empty"))
			continue
		}
		if reasons := apivalidation.ValidateUserName(user, false); len(reasons) != 0 {
			allErrs = append(allErrs, field.Invalid(idxPath, user, strings.Join(reasons, ", ")))
		}
	}

	return allErrs
}

func ValidateGroupUpdate(group *userapi.Group, old *userapi.Group) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMetaUpdate(&group.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateGroup(group)...)
	return allErrs
}

func ValidateUser(user *userapi.User) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMeta(&user.ObjectMeta, false, apivalidation.ValidateUserName, field.NewPath("metadata"))
	identitiesPath := field.NewPath("identities")
	for index, identity := range user.Identities {
		idxPath := identitiesPath.Index(index)
		if reasons := ValidateIdentityName(identity, false); len(reasons) != 0 {
			allErrs = append(allErrs, field.Invalid(idxPath, identity, strings.Join(reasons, ", ")))
		}
	}

	// our strategy should prevent us from ever hitting this case
	if len(user.Groups) != 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("groups"), user.Groups, "is deprecated and cannot be set"))
	}

	return allErrs
}

func ValidateUserUpdate(user *userapi.User, old *userapi.User) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMetaUpdate(&user.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateUser(user)...)
	return allErrs
}

func ValidateIdentity(identity *userapi.Identity) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMeta(&identity.ObjectMeta, false, ValidateIdentityName, field.NewPath("metadata"))

	if len(identity.ProviderName) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("providerName"), ""))
	} else if reasons := ValidateIdentityProviderName(identity.ProviderName); len(reasons) != 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("providerName"), identity.ProviderName, strings.Join(reasons, ", ")))
	}

	if len(identity.ProviderUserName) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("providerUserName"), ""))
	} else if reasons := ValidateIdentityProviderUserName(identity.ProviderUserName); len(reasons) != 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("providerUserName"), identity.ProviderUserName, strings.Join(reasons, ", ")))
	}

	if len(identity.ProviderName) > 0 && len(identity.ProviderUserName) > 0 {
		expectedIdentityName := identity.ProviderName + ":" + identity.ProviderUserName
		if identity.Name != expectedIdentityName {
			allErrs = append(allErrs, field.Invalid(field.NewPath("metadata", "name"), identity.Name, fmt.Sprintf("must be %s", expectedIdentityName)))
		}
	}

	userPath := field.NewPath("user")
	if reasons := apivalidation.ValidateUserName(identity.User.Name, false); len(reasons) != 0 {
		allErrs = append(allErrs, field.Invalid(userPath.Child("name"), identity.User.Name, strings.Join(reasons, ", ")))
	}
	if len(identity.User.Name) == 0 && len(identity.User.UID) != 0 {
		allErrs = append(allErrs, field.Required(userPath.Child("username"), "username is required when uid is provided"))
	}
	if len(identity.User.Name) != 0 && len(identity.User.UID) == 0 {
		allErrs = append(allErrs, field.Required(userPath.Child("uid"), "uid is required when username is provided"))
	}

	return allErrs
}

func ValidateIdentityUpdate(identity *userapi.Identity, old *userapi.Identity) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMetaUpdate(&identity.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateIdentity(identity)...)

	if identity.ProviderName != old.ProviderName {
		allErrs = append(allErrs, field.Invalid(field.NewPath("providerName"), identity.ProviderName, "may not change providerName"))
	}
	if identity.ProviderUserName != old.ProviderUserName {
		allErrs = append(allErrs, field.Invalid(field.NewPath("providerUserName"), identity.ProviderUserName, "may not change providerUserName"))
	}

	return allErrs
}

func ValidateUserIdentityMapping(mapping *userapi.UserIdentityMapping) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMeta(&mapping.ObjectMeta, false, ValidateIdentityName, field.NewPath("metadata"))

	identityPath := field.NewPath("identity")
	if len(mapping.Identity.Name) == 0 {
		allErrs = append(allErrs, field.Required(identityPath.Child("name"), ""))
	} else if mapping.Identity.Name != mapping.Name {
		allErrs = append(allErrs, field.Invalid(identityPath.Child("name"), mapping.Identity.Name, "must match metadata.name"))
	}

	if len(mapping.User.Name) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("user", "name"), ""))
	} else if reasons := apivalidation.ValidateUserName(mapping.User.Name, false); len(reasons) != 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("user", "name"), mapping.User.Name, strings.Join(reasons, ", ")))
	}

	return allErrs
}

func ValidateUserIdentityMappingUpdate(mapping *userapi.UserIdentityMapping, old *userapi.UserIdentityMapping) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMetaUpdate(&mapping.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateUserIdentityMapping(mapping)...)
	return allErrs
}
