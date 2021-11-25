package apiserver

import (
	"time"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	genericapiserver "k8s.io/apiserver/pkg/server"
	restclient "k8s.io/client-go/rest"
	openapicontroller "k8s.io/kube-aggregator/pkg/controllers/openapi"
	"k8s.io/kube-aggregator/pkg/controllers/openapi/aggregator"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"

	"github.com/openshift/oauth-apiserver/pkg/cmd/oauth-apiserver/openapiconfig"
	oauthapiserver "github.com/openshift/oauth-apiserver/pkg/oauth/apiserver"
	"github.com/openshift/oauth-apiserver/pkg/serverscheme"
	userapiserver "github.com/openshift/oauth-apiserver/pkg/user/apiserver"
	"github.com/openshift/oauth-apiserver/pkg/version"
)

type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   OAuthAPIExtraConfig
}

// OAuthAPIExtraConfig is a set of options specific to the OAuth API server
type OAuthAPIExtraConfig struct {
	// AccessTokenInactivityTimeout is a time period after which an oauthaccesstoken
	// is considered invalid unless it gets used again
	AccessTokenInactivityTimeout time.Duration
	APIAudiences                 authenticator.Audiences
}

type OAuthAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ClientConfig  *restclient.Config
	ExtraConfig   *OAuthAPIExtraConfig
}

// CompletedConfig embeds a private pointer that cannot be instantiated outside of this package.
type CompletedConfig struct {
	*completedConfig
}

func NewConfig() *Config {
	return &Config{
		GenericConfig: genericapiserver.NewRecommendedConfig(serverscheme.Codecs),
	}
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		GenericConfig: cfg.GenericConfig.Complete(),
		ClientConfig:  cfg.GenericConfig.ClientConfig,

		ExtraConfig: &cfg.ExtraConfig,
	}

	v := version.Get()
	c.GenericConfig.Version = &v

	return CompletedConfig{&c}
}

// New returns a new instance of WardleServer from the given config.
func (c completedConfig) New(delegationTarget genericapiserver.DelegationTarget) (*OAuthAPIServer, error) {
	delegateAPIServer := delegationTarget
	var err error

	delegateAPIServer, err = c.withOAuthAPIServer(delegateAPIServer)
	if err != nil {
		return nil, err
	}
	delegateAPIServer, err = c.withUserAPIServer(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	genericServer, err := c.GenericConfig.New("oauth-apiserver", delegateAPIServer)
	if err != nil {
		return nil, err
	}

	s := &OAuthAPIServer{
		GenericAPIServer: genericServer,
	}

	return s, nil
}

func (c *completedConfig) withOAuthAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &oauthapiserver.OAuthAPIServerConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{
			Config:                *c.GenericConfig.Config,
			SharedInformerFactory: c.GenericConfig.SharedInformerFactory,
			ClientConfig:          c.ClientConfig,
		},
		ExtraConfig: oauthapiserver.ExtraConfig{
			// no one is allowed to set this today
			ServiceAccountMethod: string(openshiftcontrolplanev1.GrantHandlerPrompt),

			AccessTokenInactivityTimeout: c.ExtraConfig.AccessTokenInactivityTimeout,
			ImplicitAudiences:            c.ExtraConfig.APIAudiences,
		},
	}
	config := cfg.Complete()
	server, err := config.New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return server.GenericAPIServer, nil
}

func (c *completedConfig) withUserAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &userapiserver.UserConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{Config: *c.GenericConfig.Config, SharedInformerFactory: c.GenericConfig.SharedInformerFactory, ClientConfig: c.ClientConfig},
		ExtraConfig:   userapiserver.ExtraConfig{},
	}
	config := cfg.Complete()
	server, err := config.New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return server.GenericAPIServer, nil
}

func (c *completedConfig) WithOpenAPIAggregationController(delegatedAPIServer *genericapiserver.GenericAPIServer) error {
	// We must remove openapi config-related fields from the head of the delegation chain that we pass to the OpenAPI aggregation controller.
	// This is necessary in order to prevent conflicts with the aggregation controller, as it expects the apiserver passed to it to have
	// no openapi config previously set. An alternative to stripping this data away would be to create and append a new apiserver to the head
	// of the delegation chain altogether, then pass that to the controller. But in the spirit of simplicity, we'll just strip default
	// openapi fields that may have been previously set.
	delegatedAPIServer.RemoveOpenAPIData()

	specDownloader := aggregator.NewDownloader()
	openAPIAggregator, err := aggregator.BuildAndRegisterAggregator(
		&specDownloader,
		delegatedAPIServer,
		delegatedAPIServer.Handler.GoRestfulContainer.RegisteredWebServices(),
		openapiconfig.DefaultOpenAPIConfig(),
		delegatedAPIServer.Handler.NonGoRestfulMux)
	if err != nil {
		return err
	}
	openAPIAggregationController := openapicontroller.NewAggregationController(&specDownloader, openAPIAggregator)

	delegatedAPIServer.AddPostStartHook("apiservice-openapi-controller", func(context genericapiserver.PostStartHookContext) error {
		go openAPIAggregationController.Run(context.StopCh)
		return nil
	})
	return nil
}
