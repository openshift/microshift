package hardcodedauthorizer

import (
	"context"

	"k8s.io/apiserver/pkg/authorization/authorizer"
)

type tokenReviewAuthorizer struct{}

// GetUser() user.Info - checked
// GetVerb() string - checked
// IsReadOnly() bool - na
// GetNamespace() string - checked
// GetResource() string - checked
// GetSubresource() string - checked
// GetName() string - na
// GetAPIGroup() string - checked
// GetAPIVersion() string - na
// IsResourceRequest() bool - checked
// GetPath() string - na
func (tokenReviewAuthorizer) Authorize(ctx context.Context, a authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
	if a.GetUser().GetName() != "system:serviceaccount:openshift-oauth-apiserver:openshift-authenticator" {
		return authorizer.DecisionNoOpinion, "", nil
	}
	if a.IsResourceRequest() &&
		a.GetVerb() == "create" &&
		a.GetAPIGroup() == "oauth.openshift.io" &&
		a.GetResource() == "tokenreviews" &&
		len(a.GetSubresource()) == 0 &&
		len(a.GetNamespace()) == 0 {
		return authorizer.DecisionAllow, "requesting tokenreviews is allowed", nil
	}

	return authorizer.DecisionNoOpinion, "", nil
}

// NewHardCodedTokenReviewAuthorizer returns an authorizer that allows the expected kube-apiserver user to run tokenreviews.
func NewHardCodedTokenReviewAuthorizer() *tokenReviewAuthorizer {
	return new(tokenReviewAuthorizer)
}
