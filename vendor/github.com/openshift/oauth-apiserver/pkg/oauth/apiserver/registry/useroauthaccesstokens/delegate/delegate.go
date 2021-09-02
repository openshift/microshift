package delegate

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metainternal "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/openshift/api/oauth"
	oauthv1 "github.com/openshift/api/oauth/v1"

	oauthapi "github.com/openshift/oauth-apiserver/pkg/oauth/apis/oauth"
	accesstokenregistry "github.com/openshift/oauth-apiserver/pkg/oauth/apiserver/registry/oauthaccesstoken/etcd"
	oauthprinters "github.com/openshift/oauth-apiserver/pkg/oauth/printers/internalversion"
	"github.com/openshift/oauth-apiserver/pkg/printers"
	"github.com/openshift/oauth-apiserver/pkg/printerstorage"
)

// REST implements a RESTStorage for access tokens a user owns (based on the userName field)
type REST struct {
	accessTokenStorage *accesstokenregistry.REST

	tableConvertor rest.TableConvertor
}

// we only allow retrieving the tokens => no Create() or Update()
var _ rest.Lister = &REST{}
var _ rest.Getter = &REST{}
var _ rest.Watcher = &REST{}
var _ rest.GracefulDeleter = &REST{}
var _ rest.Scoper = &REST{}

// NewREST returns a RESTStorage object that will work against access tokens
func NewREST(accessTokenStorage *accesstokenregistry.REST) (*REST, error) {
	return &REST{
		accessTokenStorage: accessTokenStorage,
		tableConvertor:     printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(oauthprinters.AddOAuthOpenShiftHandler)},
	}, nil
}

func (r *REST) New() runtime.Object {
	return &oauthv1.UserOAuthAccessToken{}
}

func (r *REST) NewList() runtime.Object {
	return &oauthv1.UserOAuthAccessTokenList{}
}

func (r *REST) List(ctx context.Context, options *metainternal.ListOptions) (runtime.Object, error) {
	ctxUserName, ok := getUserFromContext(ctx)
	if !ok {
		return &oauthapi.UserOAuthAccessTokenList{}, nil
	}

	sanitizedListOpts := listOptionsWithUserNameFilter(options, ctxUserName)

	tokenListGenericObj, err := r.accessTokenStorage.List(ctx, sanitizedListOpts)
	if err != nil {
		return nil, err
	}

	tokenList, ok := tokenListGenericObj.(*oauthapi.OAuthAccessTokenList)
	if !ok {
		return nil, errors.NewInternalError(fmt.Errorf("failed to convert generic accesstoken LIST result to its typed version"))
	}

	return oauthAccessTokenListToUserOAuthAccessTokenList(tokenList, ctxUserName), nil
}

func (r *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	ctxUserName, ok := getUserFromContext(ctx)
	if !ok {
		return nil, errors.NewNotFound(oauth.Resource("useroauthaccesstokens"), name)
	}

	tokenGenericObj, err := r.accessTokenStorage.Get(ctx, name, options)
	if err != nil {
		return nil, err
	}

	token, ok := tokenGenericObj.(*oauthapi.OAuthAccessToken)
	if !ok {
		return nil, errors.NewInternalError(fmt.Errorf("failed to convert generic accesstoken GET result to its typed version"))
	}

	if !isValidUserToken(token, ctxUserName) {
		return nil, errors.NewNotFound(oauth.Resource("useroauthaccesstokens"), name)
	}

	return (*oauthapi.UserOAuthAccessToken)(token), nil
}

func (r *REST) Watch(ctx context.Context, options *metainternal.ListOptions) (watch.Interface, error) {
	ctxUserName, ok := getUserFromContext(ctx)
	if !ok {
		return watch.NewEmptyWatch(), nil
	}

	sanitizedListOpts := listOptionsWithUserNameFilter(options, ctxUserName)

	tokenListWatcher, err := r.accessTokenStorage.Watch(ctx, sanitizedListOpts)
	if err != nil {
		return nil, err
	}

	watcher := newOAuthAccessTokenWatcher(tokenListWatcher, ctxUserName)
	go watcher.Watch(ctx)
	return watcher, nil
}

func (r *REST) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return r.tableConvertor.ConvertToTable(ctx, object, tableOptions)
}

func (r *REST) NamespaceScoped() bool {
	return false
}

func (r *REST) Delete(ctx context.Context, name string, validateFunc rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	ctxUserName, ok := getUserFromContext(ctx)
	if !ok {
		return nil, false, errors.NewNotFound(oauth.Resource("useroauthaccesstokens"), name)
	}

	var deletedRV string
	if options.Preconditions != nil && options.Preconditions.ResourceVersion != nil {
		deletedRV = *options.Preconditions.ResourceVersion
	}

	deletedObj, err := r.Get(ctx, name, &metav1.GetOptions{TypeMeta: options.TypeMeta, ResourceVersion: deletedRV})
	if err != nil {
		return nil, false, err
	}

	deletedTyped := deletedObj.(*oauthapi.UserOAuthAccessToken)
	if !isValidUserToken((*oauthapi.OAuthAccessToken)(deletedTyped), ctxUserName) {
		return nil, false, errors.NewNotFound(oauth.Resource("useroauthaccesstokens"), name)
	}

	// don't mutate the original options
	newOpts := options.DeepCopy()
	if newOpts.Preconditions == nil {
		newOpts.Preconditions = &metav1.Preconditions{UID: &deletedTyped.UID}
	} else if newOpts.Preconditions.UID == nil {
		newOpts.Preconditions.UID = &deletedTyped.UID
	}

	return r.accessTokenStorage.Delete(ctx, name, validateFunc, newOpts)
}

func getUserFromContext(ctx context.Context) (string, bool) {
	userInfo, exists := apirequest.UserFrom(ctx)
	if !exists {
		return "", false
	}

	userName := userInfo.GetName()
	if len(userName) == 0 {
		return "", false
	}

	return userName, true
}

func oauthAccessTokenListToUserOAuthAccessTokenList(l *oauthapi.OAuthAccessTokenList, username string) *oauthapi.UserOAuthAccessTokenList {
	ret := oauthapi.UserOAuthAccessTokenList{}
	for _, t := range l.Items {
		if !isValidUserToken(&t, username) {
			continue
		}
		ret.Items = append(ret.Items, oauthapi.UserOAuthAccessToken(t))
	}

	ret.ListMeta = l.ListMeta
	return &ret
}

func listOptionsWithUserNameFilter(opts *metainternal.ListOptions, userName string) *metainternal.ListOptions {
	var newOpts *metainternal.ListOptions
	if opts == nil {
		newOpts = &metainternal.ListOptions{}
	} else {
		newOpts = opts.DeepCopy()
	}

	userSelector := fields.Set{"userName": userName}.AsSelector()
	if newOpts.FieldSelector != nil {
		fields.AndSelectors(userSelector, newOpts.FieldSelector)
	} else {
		newOpts.FieldSelector = userSelector
	}
	return newOpts
}

// isValidUserToken returns true if the token has the sha256~ prefix and token.User matches
// the username provided
func isValidUserToken(token *oauthapi.OAuthAccessToken, username string) bool {
	// don't reveal other people tokens' hashes or non-hashed tokens
	return token.UserName == username && strings.HasPrefix(token.Name, "sha256~")
}
