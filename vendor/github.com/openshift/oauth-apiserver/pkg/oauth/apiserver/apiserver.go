package apiserver

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/group"
	"k8s.io/apiserver/pkg/authentication/request/bearertoken"
	tokenunion "k8s.io/apiserver/pkg/authentication/token/union"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	oauthapiv1 "github.com/openshift/api/oauth/v1"
	oauthclients "github.com/openshift/client-go/oauth/clientset/versioned"
	oauthinformer "github.com/openshift/client-go/oauth/informers/externalversions"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	userclient "github.com/openshift/client-go/user/clientset/versioned"
	userinformer "github.com/openshift/client-go/user/informers/externalversions"
	bootstrap "github.com/openshift/library-go/pkg/authentication/bootstrapauthenticator"
	"github.com/openshift/library-go/pkg/oauth/oauthserviceaccountclient"

	accesstokenetcd "github.com/openshift/oauth-apiserver/pkg/oauth/apiserver/registry/oauthaccesstoken/etcd"
	authorizetokenetcd "github.com/openshift/oauth-apiserver/pkg/oauth/apiserver/registry/oauthauthorizetoken/etcd"
	clientetcd "github.com/openshift/oauth-apiserver/pkg/oauth/apiserver/registry/oauthclient/etcd"
	clientauthetcd "github.com/openshift/oauth-apiserver/pkg/oauth/apiserver/registry/oauthclientauthorization/etcd"
	tokenreviews "github.com/openshift/oauth-apiserver/pkg/oauth/apiserver/registry/tokenreviews"
	useroauthaccesstokensdelegate "github.com/openshift/oauth-apiserver/pkg/oauth/apiserver/registry/useroauthaccesstokens/delegate"
	"github.com/openshift/oauth-apiserver/pkg/serverscheme"
	"github.com/openshift/oauth-apiserver/pkg/tokenvalidation"
	"github.com/openshift/oauth-apiserver/pkg/tokenvalidation/usercache"
)

const (
	defaultInformerResyncPeriod     = 10 * time.Minute
	minimumInactivityTimeoutSeconds = 300
	authenticatedOAuthGroup         = "system:authenticated:oauth"
)

type ExtraConfig struct {
	ServiceAccountMethod         string
	AccessTokenInactivityTimeout time.Duration
	ImplicitAudiences            authenticator.Audiences

	UserInformers  userinformer.SharedInformerFactory
	OAuthInformers oauthinformer.SharedInformerFactory
}

type OAuthAPIServerConfig struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

type OAuthAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig             genericapiserver.CompletedConfig
	ExtraConfig               *ExtraConfig
	kubeAPIServerClientConfig *restclient.Config
}

type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *OAuthAPIServerConfig) Complete() completedConfig {
	cfg := completedConfig{
		GenericConfig:             c.GenericConfig.Complete(),
		ExtraConfig:               &c.ExtraConfig,
		kubeAPIServerClientConfig: c.GenericConfig.ClientConfig,
	}

	return cfg
}

// New returns a new instance of OAuthAPIServer from the given config.
func (c completedConfig) New(delegationTarget genericapiserver.DelegationTarget) (*OAuthAPIServer, error) {
	genericServer, err := c.GenericConfig.New("oauth.openshift.io-apiserver", delegationTarget)
	if err != nil {
		return nil, err
	}

	s := &OAuthAPIServer{
		GenericAPIServer: genericServer,
	}

	coreV1Client, err := corev1.NewForConfig(c.kubeAPIServerClientConfig)
	if err != nil {
		return nil, err
	}
	oauthClient, err := oauthclients.NewForConfig(c.GenericConfig.LoopbackClientConfig)
	if err != nil {
		return nil, err
	}
	userClient, err := userclient.NewForConfig(c.GenericConfig.LoopbackClientConfig)
	if err != nil {
		return nil, err
	}

	c.ExtraConfig.UserInformers = userinformer.NewSharedInformerFactory(userClient, defaultInformerResyncPeriod)
	// add indexes to the userinformer for the users to be used in user <-> groups mapping
	if err := c.ExtraConfig.UserInformers.User().V1().Groups().Informer().AddIndexers(cache.Indexers{
		usercache.ByUserIndexName: usercache.ByUserIndexKeys,
	}); err != nil {
		return nil, err
	}
	postStartHooks := map[string]genericapiserver.PostStartHookFunc{}
	postStartHooks["openshift.io-StartUserInformer"] = func(ctx genericapiserver.PostStartHookContext) error {
		go c.ExtraConfig.UserInformers.Start(ctx.StopCh)
		return nil
	}

	c.ExtraConfig.OAuthInformers = oauthinformer.NewSharedInformerFactory(oauthClient, defaultInformerResyncPeriod)
	postStartHooks["openshift.io-StartOAuthInformer"] = func(ctx genericapiserver.PostStartHookContext) error {
		go c.ExtraConfig.OAuthInformers.Start(ctx.StopCh)
		return nil
	}

	v1Storage, storagePostStartHooks, err := c.newV1RESTStorage(coreV1Client, oauthClient, userClient)
	if err != nil {
		return nil, err
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(oauthapiv1.GroupName, serverscheme.Scheme, metav1.ParameterCodec, serverscheme.Codecs)
	apiGroupInfo.VersionedResourcesStorageMap[oauthapiv1.SchemeGroupVersion.Version] = v1Storage
	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	for hookname := range postStartHooks {
		s.GenericAPIServer.AddPostStartHookOrDie(hookname, postStartHooks[hookname])
	}

	for hookname := range storagePostStartHooks {
		s.GenericAPIServer.AddPostStartHookOrDie(hookname, storagePostStartHooks[hookname])
	}

	return s, nil
}

func (c *completedConfig) newV1RESTStorage(
	corev1Client corev1.CoreV1Interface,
	oauthClient *oauthclients.Clientset,
	userClient *userclient.Clientset,
) (map[string]rest.Storage, map[string]genericapiserver.PostStartHookFunc, error) {
	clientStorage, err := clientetcd.NewREST(c.GenericConfig.RESTOptionsGetter)
	if err != nil {
		return nil, nil, fmt.Errorf("error building REST storage: %v", err)
	}

	// If OAuth is disabled, set the strategy to Deny
	saAccountGrantMethod := oauthapiv1.GrantHandlerDeny
	if len(c.ExtraConfig.ServiceAccountMethod) > 0 {
		// Otherwise, take the value provided in master-config.yaml
		saAccountGrantMethod = oauthapiv1.GrantHandlerType(c.ExtraConfig.ServiceAccountMethod)
	}

	routeClient, err := routeclient.NewForConfig(c.kubeAPIServerClientConfig)
	if err != nil {
		return nil, nil, err
	}

	combinedOAuthClientGetter := oauthserviceaccountclient.NewServiceAccountOAuthClientGetter(
		corev1Client,
		corev1Client,
		corev1Client.Events(""),
		routeClient,
		oauthClient.OauthV1().OAuthClients(),
		saAccountGrantMethod,
	)
	authorizeTokenStorage, err := authorizetokenetcd.NewREST(c.GenericConfig.RESTOptionsGetter, combinedOAuthClientGetter)
	if err != nil {
		return nil, nil, fmt.Errorf("error building REST storage: %v", err)
	}
	accessTokenStorage, err := accesstokenetcd.NewREST(c.GenericConfig.RESTOptionsGetter, combinedOAuthClientGetter)
	if err != nil {
		return nil, nil, fmt.Errorf("error building REST storage: %v", err)
	}
	clientAuthorizationStorage, err := clientauthetcd.NewREST(c.GenericConfig.RESTOptionsGetter, combinedOAuthClientGetter)
	if err != nil {
		return nil, nil, fmt.Errorf("error building REST storage: %v", err)
	}
	userOAuthAccessTokensDelegate, err := useroauthaccesstokensdelegate.NewREST(accessTokenStorage)
	if err != nil {
		return nil, nil, fmt.Errorf("error building REST storage: %v", err)
	}
	tokenReviewStorage, tokenReviewPostStartHooks, err := c.tokenReviewStorage(corev1Client, oauthClient, userClient)
	if err != nil {
		return nil, nil, fmt.Errorf("error building REST storage: %v", err)
	}

	v1Storage := map[string]rest.Storage{
		"oAuthAuthorizeTokens":      authorizeTokenStorage,
		"oAuthAccessTokens":         accessTokenStorage,
		"oAuthClients":              clientStorage,
		"oAuthClientAuthorizations": clientAuthorizationStorage,
		"userOAuthAccessTokens":     userOAuthAccessTokensDelegate,
		"tokenReviews":              tokenReviewStorage,
	}
	return v1Storage, tokenReviewPostStartHooks, nil
}

func (c *completedConfig) tokenReviewStorage(
	corev1Client corev1.CoreV1Interface,
	oauthClient *oauthclients.Clientset,
	userClient *userclient.Clientset,
) (rest.Storage, map[string]genericapiserver.PostStartHookFunc, error) {
	openshiftAuthenticators, postStartHooks := c.getOpenShiftAuthenticators(corev1Client, oauthClient, userClient)

	tokenAuth := bearertoken.New(tokenunion.New(openshiftAuthenticators...))
	tokenReviewWrapper, err := tokenreviews.NewREST(tokenAuth)

	return tokenReviewWrapper, postStartHooks, err
}

func (c *completedConfig) getOpenShiftAuthenticators(
	corev1Client corev1.CoreV1Interface,
	oauthClient *oauthclients.Clientset,
	userClient *userclient.Clientset,
) ([]authenticator.Token, map[string]genericapiserver.PostStartHookFunc) {
	tokenAuthenticators := []authenticator.Token{}

	bootstrapUserDataGetter := bootstrap.NewBootstrapUserDataGetter(corev1Client, corev1Client)

	oauthInformer := c.ExtraConfig.OAuthInformers
	userInformer := c.ExtraConfig.UserInformers

	timeoutValidator := tokenvalidation.NewTimeoutValidator(oauthClient.OauthV1().OAuthAccessTokens(), oauthInformer.Oauth().V1().OAuthClients().Lister(), c.ExtraConfig.AccessTokenInactivityTimeout, minimumInactivityTimeoutSeconds)
	// add our oauth token validator
	validators := []tokenvalidation.OAuthTokenValidator{tokenvalidation.NewExpirationValidator(), tokenvalidation.NewUIDValidator(), timeoutValidator}

	postStartHooks := map[string]genericapiserver.PostStartHookFunc{}
	postStartHooks["openshift.io-StartTokenTimeoutUpdater"] = func(ctx genericapiserver.PostStartHookContext) error {
		go timeoutValidator.Run(ctx.StopCh)
		return nil
	}

	groupMapper := usercache.NewGroupCache(userInformer.User().V1().Groups())
	oauthTokenAuthenticator := tokenvalidation.NewTokenAuthenticator(oauthClient.OauthV1().OAuthAccessTokens(), userClient.UserV1().Users(), groupMapper, c.ExtraConfig.ImplicitAudiences, validators...)
	tokenAuthenticators = append(tokenAuthenticators,
		// if you have an OAuth bearer token, you're a human (usually)
		group.NewTokenGroupAdder(oauthTokenAuthenticator, []string{authenticatedOAuthGroup}))

	// add the bootstrap user token authenticator
	tokenAuthenticators = append(tokenAuthenticators,
		// bootstrap oauth user that can do anything, backed by a secret
		tokenvalidation.NewBootstrapAuthenticator(oauthClient.OauthV1().OAuthAccessTokens(), bootstrapUserDataGetter, c.ExtraConfig.ImplicitAudiences, validators...))

	return tokenAuthenticators, postStartHooks
}
