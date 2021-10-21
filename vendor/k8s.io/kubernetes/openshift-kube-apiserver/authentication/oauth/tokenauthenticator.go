package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kauthenticator "k8s.io/apiserver/pkg/authentication/authenticator"
	kuser "k8s.io/apiserver/pkg/authentication/user"

	authorizationv1 "github.com/openshift/api/authorization/v1"
	oauthclient "github.com/openshift/client-go/oauth/clientset/versioned/typed/oauth/v1"
	userclient "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
)

var (
	errLookup    = errors.New("token lookup failed")
	errOldFormat = errors.New("old, insecure token format")
)

type tokenAuthenticator struct {
	tokens       oauthclient.OAuthAccessTokenInterface
	users        userclient.UserInterface
	groupMapper  UserToGroupMapper
	validators   OAuthTokenValidator
	implicitAuds kauthenticator.Audiences
}

func NewTokenAuthenticator(tokens oauthclient.OAuthAccessTokenInterface, users userclient.UserInterface, groupMapper UserToGroupMapper, implicitAuds kauthenticator.Audiences, validators ...OAuthTokenValidator) kauthenticator.Token {
	return &tokenAuthenticator{
		tokens:       tokens,
		users:        users,
		groupMapper:  groupMapper,
		validators:   OAuthTokenValidators(validators),
		implicitAuds: implicitAuds,
	}
}

const sha256Prefix = "sha256~"

func (a *tokenAuthenticator) AuthenticateToken(ctx context.Context, name string) (*kauthenticator.Response, bool, error) {
	// hash token for new-style sha256~ prefixed token
	// TODO: reject non-sha256 prefix tokens in 4.7+
	if !strings.HasPrefix(name, sha256Prefix) {
		return nil, false, errOldFormat
	}

	withoutPrefix := strings.TrimPrefix(name, sha256Prefix)
	h := sha256.Sum256([]byte(withoutPrefix))
	name = sha256Prefix + base64.RawURLEncoding.EncodeToString(h[0:])

	token, err := a.tokens.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, false, errLookup // mask the error so we do not leak token data in logs
	}

	user, err := a.users.Get(context.TODO(), token.UserName, metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}

	if err := a.validators.Validate(token, user); err != nil {
		return nil, false, err
	}

	groups, err := a.groupMapper.GroupsFor(user.Name)
	if err != nil {
		return nil, false, err
	}
	groupNames := make([]string, 0, len(groups))
	for _, group := range groups {
		groupNames = append(groupNames, group.Name)
	}

	tokenAudiences := a.implicitAuds
	requestedAudiences, ok := kauthenticator.AudiencesFrom(ctx)
	if !ok {
		// default to apiserver audiences
		requestedAudiences = a.implicitAuds
	}

	auds := kauthenticator.Audiences(tokenAudiences).Intersect(requestedAudiences)
	if len(auds) == 0 && len(a.implicitAuds) != 0 {
		return nil, false, fmt.Errorf("token audiences %q is invalid for the target audiences %q", tokenAudiences, requestedAudiences)
	}

	return &kauthenticator.Response{
		User: &kuser.DefaultInfo{
			Name:   user.Name,
			UID:    string(user.UID),
			Groups: groupNames,
			Extra: map[string][]string{
				authorizationv1.ScopesKey: token.Scopes,
			},
		},
		Audiences: auds,
	}, true, nil
}
