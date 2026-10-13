package selfsubjectaccessreview

import (
	"context"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	authorizationv1 "github.com/openshift/api/authorization/v1"
	authorizationscope "github.com/openshift/apiserver-library-go/pkg/authorization/scope"

	"k8s.io/kubernetes/pkg/api/legacyscheme"
	authorizationapi "k8s.io/kubernetes/pkg/apis/authorization"
	_ "k8s.io/kubernetes/pkg/apis/authorization/install"
)

type fakeAuthorizer struct {
	attrs authorizer.Attributes
}

func (f *fakeAuthorizer) Authorize(ctx context.Context, attrs authorizer.Attributes) (authorizer.Decision, string, error) {
	f.attrs = attrs
	return authorizer.DecisionNoOpinion, "", nil
}

func TestCreate(t *testing.T) {
	userNilExtra := &user.DefaultInfo{}

	userNoExtra := &user.DefaultInfo{
		Extra: make(map[string][]string),
	}

	userNoScopes := &user.DefaultInfo{
		Extra: map[string][]string{
			"extra": {"ex1", "ex2"},
		},
	}

	userWithScopesNoCheckAccess := &user.DefaultInfo{
		Extra: map[string][]string{
			"extra": {"ex1", "ex2"},
			authorizationv1.ScopesKey: {
				authorizationscope.UserInfo,
				authorizationscope.UserListAllProjects,
			},
		},
	}

	userWithScopesWithCheckAccess := &user.DefaultInfo{
		Extra: map[string][]string{
			"extra": {"ex1", "ex2"},
			authorizationv1.ScopesKey: {
				authorizationscope.UserAccessCheck,
				authorizationscope.UserInfo,
			},
		},
	}

	userWithScopeUserFull := &user.DefaultInfo{
		Extra: map[string][]string{
			"extra": {"ex1", "ex2"},
			authorizationv1.ScopesKey: {
				authorizationscope.UserAccessCheck,
				authorizationscope.UserInfo,
				authorizationscope.UserFull,
			},
		},
	}

	userWithRoleScope := &user.DefaultInfo{
		Extra: map[string][]string{
			"extra": {"ex1", "ex2"},
			authorizationv1.ScopesKey: {
				authorizationscope.UserAccessCheck,
				"role:testrole:testns",
			},
		},
	}

	testcases := map[string]struct {
		user         user.Info
		expectedUser user.Info
	}{
		"nil extra": {
			user:         userNilExtra,
			expectedUser: userNilExtra,
		},

		"no extra": {
			user:         userNoExtra,
			expectedUser: userNoExtra,
		},

		"no scopes": {
			user:         userNoScopes,
			expectedUser: userNoScopes,
		},

		"scopes exclude user:check-access": {
			user:         userWithScopesNoCheckAccess,
			expectedUser: userWithScopesNoCheckAccess,
		},

		"scopes include user:check-access": {
			user:         userWithScopesWithCheckAccess,
			expectedUser: userWithScopeUserFull,
		},

		"scopes include role scope": {
			user:         userWithRoleScope,
			expectedUser: userWithRoleScope,
		},
	}

	for k, tc := range testcases {
		auth := &fakeAuthorizer{}
		storage := NewREST(auth, legacyscheme.Scheme)
		spec := authorizationapi.SelfSubjectAccessReviewSpec{
			NonResourceAttributes: &authorizationapi.NonResourceAttributes{Verb: "get", Path: "/mypath"},
		}

		ctx := genericapirequest.WithRequestInfo(
			genericapirequest.WithUser(genericapirequest.NewContext(), tc.user),
			&genericapirequest.RequestInfo{
				APIGroup:          "authorization.k8s.io",
				APIVersion:        "v1",
				Resource:          "selfsubjectaccessreviews",
				IsResourceRequest: true,
				Verb:              "create",
			},
		)
		_, err := storage.Create(ctx, &authorizationapi.SelfSubjectAccessReview{Spec: spec}, rest.ValidateAllObjectFunc, &metav1.CreateOptions{})
		if err != nil {
			t.Errorf("%s: %v", k, err)
			continue
		}

		if !reflect.DeepEqual(auth.attrs.GetUser(), tc.expectedUser) {
			t.Errorf("%s: expected\n%#v\ngot\n%#v", k, tc.expectedUser, auth.attrs.GetUser())
		}
	}
}
