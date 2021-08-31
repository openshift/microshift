package controller

import (
	"crypto/x509"
	"time"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/group"
	"k8s.io/apiserver/pkg/authentication/request/anonymous"
	"k8s.io/apiserver/pkg/authentication/request/bearertoken"
	"k8s.io/apiserver/pkg/authentication/request/union"
	x509request "k8s.io/apiserver/pkg/authentication/request/x509"
	"k8s.io/apiserver/pkg/authentication/token/cache"
	webhooktoken "k8s.io/apiserver/plugin/pkg/authenticator/token/webhook"
	authenticationclient "k8s.io/client-go/kubernetes/typed/authentication/v1"
)

// TODO this should really be removed in favor of the generic apiserver
// newRemoteAuthenticator creates an authenticator that checks the provided remote endpoint for tokens, allows any linked clientCAs to be checked, and caches
// responses as indicated.  If no authentication is possible, the user will be system:anonymous.
func newRemoteAuthenticator(tokenReview authenticationclient.TokenReviewInterface, clientCAs *x509.CertPool, cacheTTL time.Duration) (authenticator.Request, error) {
	authenticators := []authenticator.Request{}

	// TODO audiences
	// Upstream kubernetes/apiserver introduced timeouts in https://github.com/kubernetes/kubernetes/pull/100959.
	// This change is included in v0.20.9 (used by openshift-controller-manager) but not v0.20.4 (used by openshift-apiserver).
	// Breaking the tie by reverting 4ff13dee21a76d7bd107426f18b0e560aa2c94ed.
	tokenAuthenticator, err := webhooktoken.NewFromInterface(tokenReview, nil, *webhooktoken.DefaultRetryBackoff())
	if err != nil {
		return nil, err
	}
	cachingTokenAuth := cache.New(tokenAuthenticator, false, cacheTTL, cacheTTL)
	authenticators = append(authenticators, bearertoken.New(cachingTokenAuth))

	// Client-cert auth
	if clientCAs != nil {
		opts := x509request.DefaultVerifyOptions()
		opts.Roots = clientCAs
		certauth := x509request.New(opts, x509request.CommonNameUserConversion)
		authenticators = append(authenticators, certauth)
	}

	// Anonymous requests will pass the token and cert checks without errors
	// Bad tokens or bad certs will produce errors, in which case we should not continue to authenticate them as "system:anonymous"
	return union.NewFailOnError(
		// Add the "system:authenticated" group to users that pass token/cert authentication
		group.NewAuthenticatedGroupAdder(union.New(authenticators...)),
		// Fall back to the "system:anonymous" user
		anonymous.NewAuthenticator(),
	), nil
}
