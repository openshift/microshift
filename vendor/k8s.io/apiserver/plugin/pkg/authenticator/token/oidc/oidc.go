/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
oidc implements the authenticator.Token interface using the OpenID Connect protocol.

	config := oidc.Options{
		IssuerURL:     "https://accounts.google.com",
		ClientID:      os.Getenv("GOOGLE_CLIENT_ID"),
		UsernameClaim: "email",
	}
	tokenAuthenticator, err := oidc.New(config)
*/
package oidc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/go-oidc"

	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/apis/apiserver"
	apiservervalidation "k8s.io/apiserver/pkg/apis/apiserver/validation"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog/v2"
)

var (
	// synchronizeTokenIDVerifierForTest should be set to true to force a
	// wait until the token ID verifiers are ready.
	synchronizeTokenIDVerifierForTest = false
)

type Options struct {
	// JWTAuthenticator is the authenticator that will be used to verify the JWT.
	JWTAuthenticator apiserver.JWTAuthenticator
	// Optional KeySet to allow for synchronous initialization instead of fetching from the remote issuer.
	KeySet oidc.KeySet

	// PEM encoded root certificate contents of the provider.  Mutually exclusive with Client.
	CAContentProvider CAContentProvider

	// Optional http.Client used to make all requests to the remote issuer.  Mutually exclusive with CAContentProvider.
	Client *http.Client

	// SupportedSigningAlgs sets the accepted set of JOSE signing algorithms that
	// can be used by the provider to sign tokens.
	//
	// https://tools.ietf.org/html/rfc7518#section-3.1
	//
	// This value defaults to RS256, the value recommended by the OpenID Connect
	// spec:
	//
	// https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation
	SupportedSigningAlgs []string

	// now is used for testing. It defaults to time.Now.
	now func() time.Time
}

// Subset of dynamiccertificates.CAContentProvider that can be used to dynamically load root CAs.
type CAContentProvider interface {
	CurrentCABundleContent() []byte
}

// initVerifier creates a new ID token verifier for the given configuration and issuer URL.  On success, calls setVerifier with the
// resulting verifier.
func initVerifier(ctx context.Context, config *oidc.Config, iss string, audiences []string) (*idTokenVerifier, error) {
	provider, err := oidc.NewProvider(ctx, iss)
	if err != nil {
		return nil, fmt.Errorf("init verifier failed: %v", err)
	}
	return &idTokenVerifier{provider.Verifier(config), audiences}, nil
}

// asyncIDTokenVerifier is an ID token verifier that allows async initialization
// of the issuer check.  Must be passed by reference as it wraps sync.Mutex.
type asyncIDTokenVerifier struct {
	m sync.Mutex

	// v is the ID token verifier initialized asynchronously.  It remains nil
	// up until it is eventually initialized.
	// Guarded by m
	v *idTokenVerifier
}

// newAsyncIDTokenVerifier creates a new asynchronous token verifier.  The
// verifier is available immediately, but may remain uninitialized for some time
// after creation.
func newAsyncIDTokenVerifier(ctx context.Context, c *oidc.Config, iss string, audiences []string) *asyncIDTokenVerifier {
	t := &asyncIDTokenVerifier{}

	sync := make(chan struct{})
	// Polls indefinitely in an attempt to initialize the distributed claims
	// verifier, or until context canceled.
	initFn := func() (done bool, err error) {
		klog.V(4).Infof("oidc authenticator: attempting init: iss=%v", iss)
		v, err := initVerifier(ctx, c, iss, audiences)
		if err != nil {
			klog.Errorf("oidc authenticator: async token verifier for issuer: %q: %v", iss, err)
			return false, nil
		}
		t.m.Lock()
		defer t.m.Unlock()
		t.v = v
		close(sync)
		return true, nil
	}

	go func() {
		if done, _ := initFn(); !done {
			go wait.PollUntil(time.Second*10, initFn, ctx.Done())
		}
	}()

	if synchronizeTokenIDVerifierForTest {
		<-sync
	}

	return t
}

// verifier returns the underlying ID token verifier, or nil if one is not yet initialized.
func (a *asyncIDTokenVerifier) verifier() *idTokenVerifier {
	a.m.Lock()
	defer a.m.Unlock()
	return a.v
}

type Authenticator struct {
	jwtAuthenticator apiserver.JWTAuthenticator

	// Contains an *oidc.IDTokenVerifier. Do not access directly use the
	// idTokenVerifier method.
	verifier atomic.Value

	cancel context.CancelFunc

	// resolver is used to resolve distributed claims.
	resolver *claimResolver
}

// idTokenVerifier is a wrapper around oidc.IDTokenVerifier. It uses the oidc.IDTokenVerifier
// to verify the raw ID token and then performs audience validation locally.
type idTokenVerifier struct {
	verifier  *oidc.IDTokenVerifier
	audiences []string
}

func (a *Authenticator) setVerifier(v *idTokenVerifier) {
	a.verifier.Store(v)
}

func (a *Authenticator) idTokenVerifier() (*idTokenVerifier, bool) {
	if v := a.verifier.Load(); v != nil {
		return v.(*idTokenVerifier), true
	}
	return nil, false
}

func (a *Authenticator) Close() {
	a.cancel()
}

// whitelist of signing algorithms to ensure users don't mistakenly pass something
// goofy.
var allowedSigningAlgs = map[string]bool{
	oidc.RS256: true,
	oidc.RS384: true,
	oidc.RS512: true,
	oidc.ES256: true,
	oidc.ES384: true,
	oidc.ES512: true,
	oidc.PS256: true,
	oidc.PS384: true,
	oidc.PS512: true,
}

func New(opts Options) (*Authenticator, error) {
	if err := apiservervalidation.ValidateJWTAuthenticator(opts.JWTAuthenticator).ToAggregate(); err != nil {
		return nil, err
	}

	supportedSigningAlgs := opts.SupportedSigningAlgs
	if len(supportedSigningAlgs) == 0 {
		// RS256 is the default recommended by OpenID Connect and an 'alg' value
		// providers are required to implement.
		supportedSigningAlgs = []string{oidc.RS256}
	}
	for _, alg := range supportedSigningAlgs {
		if !allowedSigningAlgs[alg] {
			return nil, fmt.Errorf("oidc: unsupported signing alg: %q", alg)
		}
	}

	if opts.Client != nil && opts.CAContentProvider != nil {
		return nil, fmt.Errorf("oidc: Client and CAContentProvider are mutually exclusive")
	}

	client := opts.Client

	if client == nil {
		var roots *x509.CertPool
		var err error
		if opts.CAContentProvider != nil {
			// TODO(enj): make this reload CA data dynamically
			roots, err = certutil.NewPoolFromBytes(opts.CAContentProvider.CurrentCABundleContent())
			if err != nil {
				return nil, fmt.Errorf("Failed to read the CA contents: %v", err)
			}
		} else {
			klog.Info("OIDC: No x509 certificates provided, will use host's root CA set")
		}

		// Copied from http.DefaultTransport.
		tr := net.SetTransportDefaults(&http.Transport{
			// According to golang's doc, if RootCAs is nil,
			// TLS uses the host's root CA set.
			TLSClientConfig: &tls.Config{RootCAs: roots},
		})

		client = &http.Client{Transport: tr, Timeout: 30 * time.Second}
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctx = oidc.ClientContext(ctx, client)

	now := opts.now
	if now == nil {
		now = time.Now
	}

	audiences := opts.JWTAuthenticator.Issuer.Audiences
	verifierConfig := &oidc.Config{
		ClientID:             audiences[0],
		SupportedSigningAlgs: supportedSigningAlgs,
		Now:                  now,
	}
	if len(audiences) > 1 {
		verifierConfig.ClientID = ""
		// SkipClientIDCheck is set to true because we want to support multiple audiences
		// in the authentication configuration.
		// The go oidc library does not support validating
		// multiple audiences, so we have to skip the client ID check and do it ourselves.
		// xref: https://github.com/coreos/go-oidc/issues/397
		verifierConfig.SkipClientIDCheck = true
	}

	var resolver *claimResolver
	groupsClaim := opts.JWTAuthenticator.ClaimMappings.Groups.Claim
	if groupsClaim != "" {
		resolver = newClaimResolver(groupsClaim, client, verifierConfig, audiences)
	}

	authenticator := &Authenticator{
		jwtAuthenticator: opts.JWTAuthenticator,
		cancel:           cancel,
		resolver:         resolver,
	}

	if opts.KeySet != nil {
		// We already have a key set, synchronously initialize the verifier.
		authenticator.setVerifier(&idTokenVerifier{
			oidc.NewVerifier(opts.JWTAuthenticator.Issuer.URL, opts.KeySet, verifierConfig),
			audiences,
		})
	} else {
		// Asynchronously attempt to initialize the authenticator. This enables
		// self-hosted providers, providers that run on top of Kubernetes itself.
		go wait.PollImmediateUntil(10*time.Second, func() (done bool, err error) {
			provider, err := oidc.NewProvider(ctx, opts.JWTAuthenticator.Issuer.URL)
			if err != nil {
				klog.Errorf("oidc authenticator: initializing plugin: %v", err)
				return false, nil
			}

			verifier := provider.Verifier(verifierConfig)
			authenticator.setVerifier(&idTokenVerifier{verifier, audiences})
			return true, nil
		}, ctx.Done())
	}

	return authenticator, nil
}

// untrustedIssuer extracts an untrusted "iss" claim from the given JWT token,
// or returns an error if the token can not be parsed.  Since the JWT is not
// verified, the returned issuer should not be trusted.
func untrustedIssuer(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("malformed token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("error decoding token: %v", err)
	}
	claims := struct {
		// WARNING: this JWT is not verified. Do not trust these claims.
		Issuer string `json:"iss"`
	}{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("while unmarshaling token: %v", err)
	}
	// Coalesce the legacy GoogleIss with the new one.
	//
	// http://openid.net/specs/openid-connect-core-1_0.html#GoogleIss
	if claims.Issuer == "accounts.google.com" {
		return "https://accounts.google.com", nil
	}
	return claims.Issuer, nil
}

func hasCorrectIssuer(iss, tokenData string) bool {
	uiss, err := untrustedIssuer(tokenData)
	if err != nil {
		return false
	}
	if uiss != iss {
		return false
	}
	return true
}

// endpoint represents an OIDC distributed claims endpoint.
type endpoint struct {
	// URL to use to request the distributed claim.  This URL is expected to be
	// prefixed by one of the known issuer URLs.
	URL string `json:"endpoint,omitempty"`
	// AccessToken is the bearer token to use for access.  If empty, it is
	// not used.  Access token is optional per the OIDC distributed claims
	// specification.
	// See: http://openid.net/specs/openid-connect-core-1_0.html#DistributedExample
	AccessToken string `json:"access_token,omitempty"`
	// JWT is the container for aggregated claims.  Not supported at the moment.
	// See: http://openid.net/specs/openid-connect-core-1_0.html#AggregatedExample
	JWT string `json:"JWT,omitempty"`
}

// claimResolver expands distributed claims by calling respective claim source
// endpoints.
type claimResolver struct {
	// claim is the distributed claim that may be resolved.
	claim string

	// audiences is the set of acceptable audiences the JWT must be issued to.
	// At least one of the entries must match the "aud" claim in presented JWTs.
	audiences []string

	// client is the to use for resolving distributed claims
	client *http.Client

	// config is the OIDC configuration used for resolving distributed claims.
	config *oidc.Config

	// verifierPerIssuer contains, for each issuer, the appropriate verifier to use
	// for this claim.  It is assumed that there will be very few entries in
	// this map.
	// Guarded by m.
	verifierPerIssuer map[string]*asyncIDTokenVerifier

	m sync.Mutex
}

// newClaimResolver creates a new resolver for distributed claims.
func newClaimResolver(claim string, client *http.Client, config *oidc.Config, audiences []string) *claimResolver {
	return &claimResolver{
		claim:             claim,
		audiences:         audiences,
		client:            client,
		config:            config,
		verifierPerIssuer: map[string]*asyncIDTokenVerifier{},
	}
}

// Verifier returns either the verifier for the specified issuer, or error.
func (r *claimResolver) Verifier(iss string) (*idTokenVerifier, error) {
	r.m.Lock()
	av := r.verifierPerIssuer[iss]
	if av == nil {
		// This lazy init should normally be very quick.
		// TODO: Make this context cancelable.
		ctx := oidc.ClientContext(context.Background(), r.client)
		av = newAsyncIDTokenVerifier(ctx, r.config, iss, r.audiences)
		r.verifierPerIssuer[iss] = av
	}
	r.m.Unlock()

	v := av.verifier()
	if v == nil {
		return nil, fmt.Errorf("verifier not initialized for issuer: %q", iss)
	}
	return v, nil
}

// expand extracts the distributed claims from claim names and claim sources.
// The extracted claim value is pulled up into the supplied claims.
//
// Distributed claims are of the form as seen below, and are defined in the
// OIDC Connect Core 1.0, section 5.6.2.
// See: https://openid.net/specs/openid-connect-core-1_0.html#AggregatedDistributedClaims
//
//	{
//	  ... (other normal claims)...
//	  "_claim_names": {
//	    "groups": "src1"
//	  },
//	  "_claim_sources": {
//	    "src1": {
//	      "endpoint": "https://www.example.com",
//	      "access_token": "f005ba11"
//	    },
//	  },
//	}
func (r *claimResolver) expand(ctx context.Context, c claims) error {
	const (
		// The claim containing a map of endpoint references per claim.
		// OIDC Connect Core 1.0, section 5.6.2.
		claimNamesKey = "_claim_names"
		// The claim containing endpoint specifications.
		// OIDC Connect Core 1.0, section 5.6.2.
		claimSourcesKey = "_claim_sources"
	)

	_, ok := c[r.claim]
	if ok {
		// There already is a normal claim, skip resolving.
		return nil
	}
	names, ok := c[claimNamesKey]
	if !ok {
		// No _claim_names, no keys to look up.
		return nil
	}

	claimToSource := map[string]string{}
	if err := json.Unmarshal([]byte(names), &claimToSource); err != nil {
		return fmt.Errorf("oidc: error parsing distributed claim names: %v", err)
	}

	rawSources, ok := c[claimSourcesKey]
	if !ok {
		// Having _claim_names claim,  but no _claim_sources is not an expected
		// state.
		return fmt.Errorf("oidc: no claim sources")
	}

	var sources map[string]endpoint
	if err := json.Unmarshal([]byte(rawSources), &sources); err != nil {
		// The claims sources claim is malformed, this is not an expected state.
		return fmt.Errorf("oidc: could not parse claim sources: %v", err)
	}

	src, ok := claimToSource[r.claim]
	if !ok {
		// No distributed claim present.
		return nil
	}
	ep, ok := sources[src]
	if !ok {
		return fmt.Errorf("id token _claim_names contained a source %s missing in _claims_sources", src)
	}
	if ep.URL == "" {
		// This is maybe an aggregated claim (ep.JWT != "").
		return nil
	}
	return r.resolve(ctx, ep, c)
}

// resolve requests distributed claims from all endpoints passed in,
// and inserts the lookup results into allClaims.
func (r *claimResolver) resolve(ctx context.Context, endpoint endpoint, allClaims claims) error {
	// TODO: cache resolved claims.
	jwt, err := getClaimJWT(ctx, r.client, endpoint.URL, endpoint.AccessToken)
	if err != nil {
		return fmt.Errorf("while getting distributed claim %q: %v", r.claim, err)
	}
	untrustedIss, err := untrustedIssuer(jwt)
	if err != nil {
		return fmt.Errorf("getting untrusted issuer from endpoint %v failed for claim %q: %v", endpoint.URL, r.claim, err)
	}
	v, err := r.Verifier(untrustedIss)
	if err != nil {
		return fmt.Errorf("verifying untrusted issuer %v failed: %v", untrustedIss, err)
	}
	t, err := v.Verify(ctx, jwt)
	if err != nil {
		return fmt.Errorf("verify distributed claim token: %v", err)
	}
	var distClaims claims
	if err := t.Claims(&distClaims); err != nil {
		return fmt.Errorf("could not parse distributed claims for claim %v: %v", r.claim, err)
	}
	value, ok := distClaims[r.claim]
	if !ok {
		return fmt.Errorf("jwt returned by distributed claim endpoint %q did not contain claim: %v", endpoint.URL, r.claim)
	}
	allClaims[r.claim] = value
	return nil
}

func (v *idTokenVerifier) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	t, err := v.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}
	if err := v.verifyAudience(t); err != nil {
		return nil, err
	}
	return t, nil
}

// verifyAudience verifies the audience field in the ID token matches the expected audience.
// This is added based on https://github.com/coreos/go-oidc/blob/b203e58c24394ddf5e816706a7645f01280245c7/oidc/verify.go#L275-L281
// with the difference that we allow multiple audiences.
//
// AuthenticationConfiguration has a audienceMatchPolicy field, but the only supported value now is "MatchAny".
// So, The default match behavior is to match at least one of the audiences in the ID token.
func (v *idTokenVerifier) verifyAudience(t *oidc.IDToken) error {
	// We validate audience field is not empty in the authentication configuration.
	// This check ensures callers of "Verify" using idTokenVerifier are not passing
	// an empty audience.
	if len(v.audiences) == 0 {
		return fmt.Errorf("oidc: invalid configuration, audiences cannot be empty")
	}
	tokenAudiences := sets.NewString(t.Audience...)
	for _, aud := range v.audiences {
		if tokenAudiences.Has(aud) {
			return nil
		}
	}

	return fmt.Errorf("oidc: expected audience %q got %q", v.audiences, t.Audience)
}

func (a *Authenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	if !hasCorrectIssuer(a.jwtAuthenticator.Issuer.URL, token) {
		return nil, false, nil
	}

	verifier, ok := a.idTokenVerifier()
	if !ok {
		return nil, false, fmt.Errorf("oidc: authenticator not initialized")
	}

	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return nil, false, fmt.Errorf("oidc: verify token: %v", err)
	}

	var c claims
	if err := idToken.Claims(&c); err != nil {
		return nil, false, fmt.Errorf("oidc: parse claims: %v", err)
	}
	if a.resolver != nil {
		if err := a.resolver.expand(ctx, c); err != nil {
			return nil, false, fmt.Errorf("oidc: could not expand distributed claims: %v", err)
		}
	}

	var username string
	usernameClaim := a.jwtAuthenticator.ClaimMappings.Username.Claim
	if err := c.unmarshalClaim(usernameClaim, &username); err != nil {
		return nil, false, fmt.Errorf("oidc: parse username claims %q: %v", usernameClaim, err)
	}

	if usernameClaim == "email" {
		// If the email_verified claim is present, ensure the email is valid.
		// https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
		if hasEmailVerified := c.hasClaim("email_verified"); hasEmailVerified {
			var emailVerified bool
			if err := c.unmarshalClaim("email_verified", &emailVerified); err != nil {
				return nil, false, fmt.Errorf("oidc: parse 'email_verified' claim: %v", err)
			}

			// If the email_verified claim is present we have to verify it is set to `true`.
			if !emailVerified {
				return nil, false, fmt.Errorf("oidc: email not verified")
			}
		}
	}

	userNamePrefix := a.jwtAuthenticator.ClaimMappings.Username.Prefix
	if userNamePrefix != nil && *userNamePrefix != "" {
		username = *userNamePrefix + username
	}

	info := &user.DefaultInfo{Name: username}
	groupsClaim := a.jwtAuthenticator.ClaimMappings.Groups.Claim
	if groupsClaim != "" {
		if _, ok := c[groupsClaim]; ok {
			// Some admins want to use string claims like "role" as the group value.
			// Allow the group claim to be a single string instead of an array.
			//
			// See: https://github.com/kubernetes/kubernetes/issues/33290
			var groups stringOrArray
			if err := c.unmarshalClaim(groupsClaim, &groups); err != nil {
				return nil, false, fmt.Errorf("oidc: parse groups claim %q: %v", groupsClaim, err)
			}
			info.Groups = []string(groups)
		}
	}

	groupsPrefix := a.jwtAuthenticator.ClaimMappings.Groups.Prefix
	if groupsPrefix != nil && *groupsPrefix != "" {
		for i, group := range info.Groups {
			info.Groups[i] = *groupsPrefix + group
		}
	}

	// check to ensure all required claims are present in the ID token and have matching values.
	for _, claimValidationRule := range a.jwtAuthenticator.ClaimValidationRules {
		claim := claimValidationRule.Claim
		value := claimValidationRule.RequiredValue

		if !c.hasClaim(claim) {
			return nil, false, fmt.Errorf("oidc: required claim %s not present in ID token", claim)
		}

		// NOTE: Only string values are supported as valid required claim values.
		var claimValue string
		if err := c.unmarshalClaim(claim, &claimValue); err != nil {
			return nil, false, fmt.Errorf("oidc: parse claim %s: %v", claim, err)
		}
		if claimValue != value {
			return nil, false, fmt.Errorf("oidc: required claim %s value does not match. Got = %s, want = %s", claim, claimValue, value)
		}
	}

	return &authenticator.Response{User: info}, true, nil
}

// getClaimJWT gets a distributed claim JWT from url, using the supplied access
// token as bearer token.  If the access token is "", the authorization header
// will not be set.
// TODO: Allow passing in JSON hints to the IDP.
func getClaimJWT(ctx context.Context, client *http.Client, url, accessToken string) (string, error) {
	// TODO: Allow passing request body with configurable information.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("while calling %v: %v", url, err)
	}
	if accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	}
	req = req.WithContext(ctx)
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	// Report non-OK status code as an error.
	if response.StatusCode < http.StatusOK || response.StatusCode > http.StatusIMUsed {
		return "", fmt.Errorf("error while getting distributed claim JWT: %v", response.Status)
	}
	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("could not decode distributed claim response")
	}
	return string(responseBytes), nil
}

type stringOrArray []string

func (s *stringOrArray) UnmarshalJSON(b []byte) error {
	var a []string
	if err := json.Unmarshal(b, &a); err == nil {
		*s = a
		return nil
	}
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	*s = []string{str}
	return nil
}

type claims map[string]json.RawMessage

func (c claims) unmarshalClaim(name string, v interface{}) error {
	val, ok := c[name]
	if !ok {
		return fmt.Errorf("claim not present")
	}
	return json.Unmarshal([]byte(val), v)
}

func (c claims) hasClaim(name string) bool {
	if _, ok := c[name]; !ok {
		return false
	}
	return true
}
