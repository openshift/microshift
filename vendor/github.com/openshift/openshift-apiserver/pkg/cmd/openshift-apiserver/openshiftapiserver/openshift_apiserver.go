package openshiftapiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/emicklei/go-restful"

	corev1 "k8s.io/api/core/v1"
	kapierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericmux "k8s.io/apiserver/pkg/server/mux"
	kubeinformers "k8s.io/client-go/informers"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
	openapicontroller "k8s.io/kube-aggregator/pkg/controllers/openapi"
	"k8s.io/kube-aggregator/pkg/controllers/openapi/aggregator"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	rbacrest "k8s.io/kubernetes/pkg/registry/rbac/rest"
	rbacregistryvalidation "k8s.io/kubernetes/pkg/registry/rbac/validation"
	rbacauthorizer "k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	operatorinformers "github.com/openshift/client-go/operator/informers/externalversions"
	quotainformer "github.com/openshift/client-go/quota/informers/externalversions"
	securityv1informer "github.com/openshift/client-go/security/informers/externalversions"
	"github.com/openshift/library-go/pkg/quota/clusterquotamapping"
	oappsapiserver "github.com/openshift/openshift-apiserver/pkg/apps/apiserver"
	authorizationapiserver "github.com/openshift/openshift-apiserver/pkg/authorization/apiserver"
	"github.com/openshift/openshift-apiserver/pkg/bootstrappolicy"
	buildapiserver "github.com/openshift/openshift-apiserver/pkg/build/apiserver"
	"github.com/openshift/openshift-apiserver/pkg/cmd/openshift-apiserver/openshiftapiserver/configprocessing"
	imageapiserver "github.com/openshift/openshift-apiserver/pkg/image/apiserver"
	"github.com/openshift/openshift-apiserver/pkg/image/apiserver/registryhostname"
	projectapiserver "github.com/openshift/openshift-apiserver/pkg/project/apiserver"
	projectauth "github.com/openshift/openshift-apiserver/pkg/project/auth"
	projectcache "github.com/openshift/openshift-apiserver/pkg/project/cache"
	quotaapiserver "github.com/openshift/openshift-apiserver/pkg/quota/apiserver"
	routeapiserver "github.com/openshift/openshift-apiserver/pkg/route/apiserver"
	"github.com/openshift/openshift-apiserver/pkg/route/apiserver/routeallocationcontroller"
	securityapiserver "github.com/openshift/openshift-apiserver/pkg/security/apiserver"
	templateapiserver "github.com/openshift/openshift-apiserver/pkg/template/apiserver"
	"github.com/openshift/openshift-apiserver/pkg/version"

	// register api groups
	_ "github.com/openshift/openshift-apiserver/pkg/api/install"
)

type OpenshiftAPIExtraConfig struct {
	// we phrase it like this so we can build the post-start-hook, but no one can take more indirect dependencies on informers
	InformerStart func(stopCh <-chan struct{})

	KubeAPIServerClientConfig *rest.Config
	KubeInformers             kubeinformers.SharedInformerFactory

	QuotaInformers    quotainformer.SharedInformerFactory
	SecurityInformers securityv1informer.SharedInformerFactory
	OperatorInformers operatorinformers.SharedInformerFactory

	// these are all required to build our storage
	RuleResolver   rbacregistryvalidation.AuthorizationRuleResolver
	SubjectLocator rbacauthorizer.SubjectLocator

	// for Images
	// RegistryHostnameRetriever retrieves the internal and external hostname of
	// the integrated registry, or false if no such registry is available.
	RegistryHostnameRetriever          registryhostname.RegistryHostnameRetriever
	AllowedRegistriesForImport         openshiftcontrolplanev1.AllowedRegistries
	MaxImagesBulkImportedPerRepository int
	AdditionalTrustedCA                []byte

	RouteAllocator *routeallocationcontroller.RouteAllocationController

	ProjectAuthorizationCache *projectauth.AuthorizationCache
	ProjectCache              *projectcache.ProjectCache
	ProjectRequestTemplate    string
	ProjectRequestMessage     string
	RESTMapper                *restmapper.DeferredDiscoveryRESTMapper

	ClusterQuotaMappingController *clusterquotamapping.ClusterQuotaMappingController
}

// Validate helps ensure that we build this config correctly, because there are lots of bits to remember for now
func (c *OpenshiftAPIExtraConfig) Validate() error {
	ret := []error{}

	if c.KubeInformers == nil {
		ret = append(ret, fmt.Errorf("KubeInformers is required"))
	}
	if c.QuotaInformers == nil {
		ret = append(ret, fmt.Errorf("QuotaInformers is required"))
	}
	if c.SecurityInformers == nil {
		ret = append(ret, fmt.Errorf("SecurityInformers is required"))
	}
	if c.OperatorInformers == nil {
		ret = append(ret, fmt.Errorf("OperatorInformers is required"))
	}
	if c.RuleResolver == nil {
		ret = append(ret, fmt.Errorf("RuleResolver is required"))
	}
	if c.SubjectLocator == nil {
		ret = append(ret, fmt.Errorf("SubjectLocator is required"))
	}
	if c.RegistryHostnameRetriever == nil {
		ret = append(ret, fmt.Errorf("RegistryHostnameRetriever is required"))
	}
	if c.RouteAllocator == nil {
		ret = append(ret, fmt.Errorf("RouteAllocator is required"))
	}
	if c.ProjectAuthorizationCache == nil {
		ret = append(ret, fmt.Errorf("ProjectAuthorizationCache is required"))
	}
	if c.ProjectCache == nil {
		ret = append(ret, fmt.Errorf("ProjectCache is required"))
	}
	if c.ClusterQuotaMappingController == nil {
		ret = append(ret, fmt.Errorf("ClusterQuotaMappingController is required"))
	}
	if c.RESTMapper == nil {
		ret = append(ret, fmt.Errorf("RESTMapper is required"))
	}

	return utilerrors.NewAggregate(ret)
}

type OpenshiftAPIConfig struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   OpenshiftAPIExtraConfig
}

// OpenshiftAPIServer is only responsible for serving the APIs for Openshift
// It does NOT expose oauth, related oauth endpoints, or any kube APIs.
type OpenshiftAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *OpenshiftAPIExtraConfig
}

type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *OpenshiftAPIConfig) Complete() completedConfig {
	cfg := completedConfig{
		c.GenericConfig.Complete(),
		&c.ExtraConfig,
	}

	return cfg
}

func (c *completedConfig) withAppsAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &oappsapiserver.AppsServerConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{Config: *c.GenericConfig.Config, SharedInformerFactory: c.GenericConfig.SharedInformerFactory},
		ExtraConfig: oappsapiserver.ExtraConfig{
			KubeAPIServerClientConfig: c.ExtraConfig.KubeAPIServerClientConfig,
			Codecs:                    legacyscheme.Codecs,
			Scheme:                    legacyscheme.Scheme,
		},
	}
	config := cfg.Complete()
	server, err := config.New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return server.GenericAPIServer, nil
}

func (c *completedConfig) withAuthorizationAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &authorizationapiserver.AuthorizationAPIServerConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{Config: *c.GenericConfig.Config, SharedInformerFactory: c.GenericConfig.SharedInformerFactory},
		ExtraConfig: authorizationapiserver.ExtraConfig{
			KubeAPIServerClientConfig: c.ExtraConfig.KubeAPIServerClientConfig,
			KubeInformers:             c.ExtraConfig.KubeInformers,
			RuleResolver:              c.ExtraConfig.RuleResolver,
			SubjectLocator:            c.ExtraConfig.SubjectLocator,
			Codecs:                    legacyscheme.Codecs,
			Scheme:                    legacyscheme.Scheme,
		},
	}
	config := cfg.Complete()
	server, err := config.New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return server.GenericAPIServer, nil
}

func (c *completedConfig) withBuildAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &buildapiserver.BuildServerConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{Config: *c.GenericConfig.Config, SharedInformerFactory: c.GenericConfig.SharedInformerFactory},
		ExtraConfig: buildapiserver.ExtraConfig{
			KubeAPIServerClientConfig: c.ExtraConfig.KubeAPIServerClientConfig,
			Codecs:                    legacyscheme.Codecs,
			Scheme:                    legacyscheme.Scheme,
		},
	}
	config := cfg.Complete()
	server, err := config.New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return server.GenericAPIServer, nil
}

func (c *completedConfig) withImageAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &imageapiserver.ImageAPIServerConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{Config: *c.GenericConfig.Config, SharedInformerFactory: c.GenericConfig.SharedInformerFactory},
		ExtraConfig: imageapiserver.ExtraConfig{
			KubeAPIServerClientConfig:          c.ExtraConfig.KubeAPIServerClientConfig,
			RegistryHostnameRetriever:          c.ExtraConfig.RegistryHostnameRetriever,
			AllowedRegistriesForImport:         c.ExtraConfig.AllowedRegistriesForImport,
			MaxImagesBulkImportedPerRepository: c.ExtraConfig.MaxImagesBulkImportedPerRepository,
			Codecs:                             legacyscheme.Codecs,
			Scheme:                             legacyscheme.Scheme,
			AdditionalTrustedCA:                c.ExtraConfig.AdditionalTrustedCA,
			OperatorInformers:                  c.ExtraConfig.OperatorInformers,
		},
	}
	config := cfg.Complete()
	server, err := config.New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return server.GenericAPIServer, nil
}

func (c *completedConfig) withProjectAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &projectapiserver.ProjectAPIServerConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{Config: *c.GenericConfig.Config, SharedInformerFactory: c.GenericConfig.SharedInformerFactory},
		ExtraConfig: projectapiserver.ExtraConfig{
			KubeAPIServerClientConfig: c.ExtraConfig.KubeAPIServerClientConfig,
			ProjectAuthorizationCache: c.ExtraConfig.ProjectAuthorizationCache,
			ProjectCache:              c.ExtraConfig.ProjectCache,
			ProjectRequestTemplate:    c.ExtraConfig.ProjectRequestTemplate,
			ProjectRequestMessage:     c.ExtraConfig.ProjectRequestMessage,
			RESTMapper:                c.ExtraConfig.RESTMapper,
			Codecs:                    legacyscheme.Codecs,
			Scheme:                    legacyscheme.Scheme,
		},
	}
	config := cfg.Complete()
	server, err := config.New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return server.GenericAPIServer, nil
}

func (c *completedConfig) withQuotaAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &quotaapiserver.QuotaAPIServerConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{Config: *c.GenericConfig.Config, SharedInformerFactory: c.GenericConfig.SharedInformerFactory},
		ExtraConfig: quotaapiserver.ExtraConfig{
			ClusterQuotaMappingController: c.ExtraConfig.ClusterQuotaMappingController,
			QuotaInformers:                c.ExtraConfig.QuotaInformers,
			Codecs:                        legacyscheme.Codecs,
			Scheme:                        legacyscheme.Scheme,
		},
	}
	config := cfg.Complete()
	server, err := config.New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return server.GenericAPIServer, nil
}

func (c *completedConfig) withRouteAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &routeapiserver.RouteAPIServerConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{Config: *c.GenericConfig.Config, SharedInformerFactory: c.GenericConfig.SharedInformerFactory},
		ExtraConfig: routeapiserver.ExtraConfig{
			KubeAPIServerClientConfig: c.ExtraConfig.KubeAPIServerClientConfig,
			RouteAllocator:            c.ExtraConfig.RouteAllocator,
			Codecs:                    legacyscheme.Codecs,
			Scheme:                    legacyscheme.Scheme,
		},
	}
	config := cfg.Complete()
	server, err := config.New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return server.GenericAPIServer, nil
}

func (c *completedConfig) withSecurityAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &securityapiserver.SecurityAPIServerConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{Config: *c.GenericConfig.Config, SharedInformerFactory: c.GenericConfig.SharedInformerFactory},
		ExtraConfig: securityapiserver.ExtraConfig{
			KubeAPIServerClientConfig: c.ExtraConfig.KubeAPIServerClientConfig,
			SecurityInformers:         c.ExtraConfig.SecurityInformers,
			KubeInformers:             c.ExtraConfig.KubeInformers,
			Authorizer:                c.GenericConfig.Authorization.Authorizer,
			Codecs:                    legacyscheme.Codecs,
			Scheme:                    legacyscheme.Scheme,
		},
	}
	config := cfg.Complete()
	server, err := config.New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return server.GenericAPIServer, nil
}

func (c *completedConfig) withTemplateAPIServer(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error) {
	cfg := &templateapiserver.TemplateConfig{
		GenericConfig: &genericapiserver.RecommendedConfig{Config: *c.GenericConfig.Config, SharedInformerFactory: c.GenericConfig.SharedInformerFactory},
		ExtraConfig: templateapiserver.ExtraConfig{
			KubeAPIServerClientConfig: c.ExtraConfig.KubeAPIServerClientConfig,
			Codecs:                    legacyscheme.Codecs,
			Scheme:                    legacyscheme.Scheme,
		},
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
		configprocessing.DefaultOpenAPIConfig(),
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

type apiServerAppenderFunc func(delegateAPIServer genericapiserver.DelegationTarget) (genericapiserver.DelegationTarget, error)

func addAPIServerOrDie(delegateAPIServer genericapiserver.DelegationTarget, apiServerAppenderFn apiServerAppenderFunc) genericapiserver.DelegationTarget {
	delegateAPIServer, err := apiServerAppenderFn(delegateAPIServer)
	if err != nil {
		klog.Fatal(err)
	}

	return delegateAPIServer
}

func (c completedConfig) New(delegationTarget genericapiserver.DelegationTarget) (*OpenshiftAPIServer, error) {
	delegateAPIServer := delegationTarget

	delegateAPIServer = addAPIServerOrDie(delegateAPIServer, c.withAppsAPIServer)
	delegateAPIServer = addAPIServerOrDie(delegateAPIServer, c.withAuthorizationAPIServer)
	delegateAPIServer = addAPIServerOrDie(delegateAPIServer, c.withBuildAPIServer)
	delegateAPIServer = addAPIServerOrDie(delegateAPIServer, c.withImageAPIServer)
	delegateAPIServer = addAPIServerOrDie(delegateAPIServer, c.withProjectAPIServer)
	delegateAPIServer = addAPIServerOrDie(delegateAPIServer, c.withQuotaAPIServer)
	delegateAPIServer = addAPIServerOrDie(delegateAPIServer, c.withRouteAPIServer)
	delegateAPIServer = addAPIServerOrDie(delegateAPIServer, c.withSecurityAPIServer)
	delegateAPIServer = addAPIServerOrDie(delegateAPIServer, c.withTemplateAPIServer)

	genericServer, err := c.GenericConfig.New("openshift-apiserver", delegateAPIServer)
	if err != nil {
		return nil, err
	}

	s := &OpenshiftAPIServer{
		GenericAPIServer: genericServer,
	}

	// this remains a non-healthz endpoint so that you can be healthy without being ready.
	addReadinessCheckRoute(s.GenericAPIServer.Handler.NonGoRestfulMux, "/healthz/ready", c.ExtraConfig.ProjectAuthorizationCache.ReadyForAccess)

	// this remains here and separate so that you can check both kube and openshift levels
	AddOpenshiftVersionRoute(s.GenericAPIServer.Handler.GoRestfulContainer, "/version/openshift")

	// register our poststarthooks
	s.GenericAPIServer.AddPostStartHookOrDie("authorization.openshift.io-bootstrapclusterroles",
		func(context genericapiserver.PostStartHookContext) error {
			newContext := genericapiserver.PostStartHookContext{
				LoopbackClientConfig: c.ExtraConfig.KubeAPIServerClientConfig,
				StopCh:               context.StopCh,
			}
			return bootstrapData(bootstrappolicy.Policy()).EnsureRBACPolicy()(newContext)

		})
	s.GenericAPIServer.AddPostStartHookOrDie("authorization.openshift.io-ensurenodebootstrap-sa", c.EnsureNodeBootstrapServiceAccount)
	s.GenericAPIServer.AddPostStartHookOrDie("project.openshift.io-projectcache", c.startProjectCache)
	s.GenericAPIServer.AddPostStartHookOrDie("project.openshift.io-projectauthorizationcache", c.startProjectAuthorizationCache)
	s.GenericAPIServer.AddPostStartHookOrDie("openshift.io-startinformers", func(context genericapiserver.PostStartHookContext) error {
		c.ExtraConfig.InformerStart(context.StopCh)
		return nil
	})
	s.GenericAPIServer.AddPostStartHookOrDie("openshift.io-restmapperupdater", func(context genericapiserver.PostStartHookContext) error {
		go func() {
			wait.Until(func() {
				c.ExtraConfig.RESTMapper.Reset()
			}, 10*time.Second, context.StopCh)
		}()
		return nil

	})
	s.GenericAPIServer.AddPostStartHookOrDie("quota.openshift.io-clusterquotamapping", func(context genericapiserver.PostStartHookContext) error {
		go c.ExtraConfig.ClusterQuotaMappingController.Run(5, context.StopCh)
		return nil
	})

	return s, nil
}

// initReadinessCheckRoute initializes an HTTP endpoint for readiness checking
func addReadinessCheckRoute(mux *genericmux.PathRecorderMux, path string, readyFunc func() bool) {
	mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		if readyFunc() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))

		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})
}

// initVersionRoute initializes an HTTP endpoint for the server's version information.
func AddOpenshiftVersionRoute(container *restful.Container, path string) {
	// Build version info once
	versionInfo, err := json.MarshalIndent(version.Get(), "", "  ")
	if err != nil {
		klog.Errorf("Unable to initialize version route: %v", err)
		return
	}

	// Set up a service to return the git code version.
	ws := new(restful.WebService)
	ws.Path(path)
	ws.Doc("git code version from which this is built")
	ws.Route(
		ws.GET("/").To(func(_ *restful.Request, resp *restful.Response) {
			writeJSON(resp, versionInfo)
		}).
			Doc("get the code version").
			Operation("getCodeVersion").
			Produces(restful.MIME_JSON))

	container.Add(ws)
}

func writeJSON(resp *restful.Response, json []byte) {
	resp.ResponseWriter.Header().Set("Content-Type", "application/json")
	resp.ResponseWriter.WriteHeader(http.StatusOK)
	resp.ResponseWriter.Write(json)
}

func (c *completedConfig) startProjectCache(context genericapiserver.PostStartHookContext) error {
	// RunProjectCache populates project cache, used by scheduler and project admission controller.
	klog.Infof("Using default project node label selector: %s", c.ExtraConfig.ProjectCache.DefaultNodeSelector)
	go c.ExtraConfig.ProjectCache.Run(context.StopCh)
	return nil
}

func (c *completedConfig) startProjectAuthorizationCache(context genericapiserver.PostStartHookContext) error {
	period := 1 * time.Second
	c.ExtraConfig.ProjectAuthorizationCache.Run(period)
	return nil
}

// EnsureNodeBootstrapServiceAccount is called as part of global policy initialization to ensure node bootstrap SA exists
func (c *completedConfig) EnsureNodeBootstrapServiceAccount(_ genericapiserver.PostStartHookContext) error {
	namespaceName := bootstrappolicy.DefaultOpenShiftInfraNamespace

	var coreClient *corev1client.CoreV1Client
	err := wait.Poll(1*time.Second, 30*time.Second, func() (bool, error) {
		var err error
		coreClient, err = corev1client.NewForConfig(c.ExtraConfig.KubeAPIServerClientConfig)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("unable to initialize client: %v", err))
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error getting client: %v", err))
		return err
	}

	// Ensure we have the bootstrap SA for Nodes
	_, err = coreClient.ServiceAccounts(namespaceName).Create(context.TODO(), &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: bootstrappolicy.InfraNodeBootstrapServiceAccountName}}, metav1.CreateOptions{})
	if err != nil && !kapierror.IsAlreadyExists(err) {
		klog.Errorf("Error creating service account %s/%s: %v", namespaceName, bootstrappolicy.InfraNodeBootstrapServiceAccountName, err)
	}

	return nil
}

// bootstrapData casts our policy data to the rbacrest helper that can
// materialize the policy.
func bootstrapData(data *bootstrappolicy.PolicyData) *rbacrest.PolicyData {
	return &rbacrest.PolicyData{
		ClusterRoles:            data.ClusterRoles,
		ClusterRoleBindings:     data.ClusterRoleBindings,
		Roles:                   data.Roles,
		RoleBindings:            data.RoleBindings,
		ClusterRolesToAggregate: data.ClusterRolesToAggregate,
	}
}
