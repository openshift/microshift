package openshiftkubeapiserver

import (
	"os"
	"time"

	"github.com/openshift/apiserver-library-go/pkg/admission/imagepolicy"
	"github.com/openshift/apiserver-library-go/pkg/admission/imagepolicy/imagereferencemutators"
	"github.com/openshift/apiserver-library-go/pkg/securitycontextconstraints/sccadmission"
	securityv1client "github.com/openshift/client-go/security/clientset/versioned"
	securityv1informer "github.com/openshift/client-go/security/informers/externalversions"
	"github.com/openshift/library-go/pkg/apiserver/admission/admissionrestconfig"
	"github.com/openshift/library-go/pkg/apiserver/apiserverconfig"
	"k8s.io/apiserver/pkg/admission"
	genericapiserver "k8s.io/apiserver/pkg/server"
	clientgoinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/openshift-kube-apiserver/admission/scheduler/nodeenv"
	"k8s.io/kubernetes/openshift-kube-apiserver/enablement"

	// magnet to get authorizer package in hack/update-vendor.sh
	_ "github.com/openshift/library-go/pkg/authorization/hardcodedauthorizer"
)

func OpenShiftKubeAPIServerConfigPatch(genericConfig *genericapiserver.Config, kubeInformers clientgoinformers.SharedInformerFactory, pluginInitializers *[]admission.PluginInitializer) error {
	if !enablement.IsOpenShift() {
		return nil
	}

	openshiftInformers, err := newInformers(genericConfig.LoopbackClientConfig)
	if err != nil {
		return err
	}

	// AUTHORIZER
	genericConfig.RequestInfoResolver = apiserverconfig.OpenshiftRequestInfoResolver()
	// END AUTHORIZER

	// Inject OpenShift API long running endpoints (like for binary builds).
	// TODO: We should disable the timeout code for aggregated endpoints as this can cause problems when upstream add additional endpoints.
	genericConfig.LongRunningFunc = apiserverconfig.IsLongRunningRequest

	// ADMISSION

	*pluginInitializers = append(*pluginInitializers,
		imagepolicy.NewInitializer(imagereferencemutators.KubeImageMutators{}, enablement.OpenshiftConfig().ImagePolicyConfig.InternalRegistryHostname),
		sccadmission.NewInitializer(openshiftInformers.getOpenshiftSecurityInformers().Security().V1().SecurityContextConstraints()),
		nodeenv.NewInitializer(enablement.OpenshiftConfig().ProjectConfig.DefaultNodeSelector),
		admissionrestconfig.NewInitializer(*rest.CopyConfig(genericConfig.LoopbackClientConfig)),
	)

	// This is needed in order to have the correct initializers for the SCC admission plugin which is used to mutate
	// PodSpecs for PodSpec-y workload objects in the pod security admission plugin.
	enablement.SCCAdmissionPlugin.SetAuthorizer(genericConfig.Authorization.Authorizer)
	enablement.SCCAdmissionPlugin.SetSecurityInformers(openshiftInformers.getOpenshiftSecurityInformers().Security().V1().SecurityContextConstraints())
	enablement.SCCAdmissionPlugin.SetExternalKubeInformerFactory(kubeInformers)
	// END ADMISSION

	openshiftAPIServiceReachabilityCheck := newOpenshiftAPIServiceReachabilityCheck(genericConfig.PublicAddress)
	oauthAPIServiceReachabilityCheck := newOAuthPIServiceReachabilityCheck(genericConfig.PublicAddress)
	genericConfig.ReadyzChecks = append(genericConfig.ReadyzChecks, openshiftAPIServiceReachabilityCheck, oauthAPIServiceReachabilityCheck)

	genericConfig.AddPostStartHookOrDie("openshift.io-startkubeinformers", func(context genericapiserver.PostStartHookContext) error {
		go openshiftInformers.Start(context.Done())
		return nil
	})
	genericConfig.AddPostStartHookOrDie("openshift.io-openshift-apiserver-reachable", func(context genericapiserver.PostStartHookContext) error {
		go openshiftAPIServiceReachabilityCheck.checkForConnection(context)
		return nil
	})
	genericConfig.AddPostStartHookOrDie("openshift.io-oauth-apiserver-reachable", func(context genericapiserver.PostStartHookContext) error {
		go oauthAPIServiceReachabilityCheck.checkForConnection(context)
		return nil
	})
	enablement.AppendPostStartHooksOrDie(genericConfig)

	return nil
}

func makeJSONRESTConfig(config *rest.Config) *rest.Config {
	c := rest.CopyConfig(config)
	c.AcceptContentTypes = "application/json"
	c.ContentType = "application/json"
	return c
}

func nodeFor() string {
	node := os.Getenv("HOST_IP")
	if hostname, err := os.Hostname(); err != nil {
		node = hostname
	}
	return node
}

// newInformers is only exposed for the build's integration testing until it can be fixed more appropriately.
func newInformers(loopbackClientConfig *rest.Config) (*kubeAPIServerInformers, error) {
	// ClusterResourceQuota is served using CRD resource any status update must use JSON
	jsonLoopbackClientConfig := makeJSONRESTConfig(loopbackClientConfig)

	securityClient, err := securityv1client.NewForConfig(jsonLoopbackClientConfig)
	if err != nil {
		return nil, err
	}

	// TODO find a single place to create and start informers.  During the 1.7 rebase this will come more naturally in a config object,
	// before then we should try to eliminate our direct to storage access.  It's making us do weird things.
	const defaultInformerResyncPeriod = 10 * time.Minute

	ret := &kubeAPIServerInformers{
		OpenshiftSecurityInformers: securityv1informer.NewSharedInformerFactory(securityClient, defaultInformerResyncPeriod),
	}

	return ret, nil
}

type kubeAPIServerInformers struct {
	OpenshiftSecurityInformers securityv1informer.SharedInformerFactory
}

func (i *kubeAPIServerInformers) getOpenshiftSecurityInformers() securityv1informer.SharedInformerFactory {
	return i.OpenshiftSecurityInformers
}

func (i *kubeAPIServerInformers) Start(stopCh <-chan struct{}) {
	i.OpenshiftSecurityInformers.Start(stopCh)
}
