package openshiftadmission

import (
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/plugin/resourcequota"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/kubernetes/plugin/pkg/admission/gc"

	"github.com/openshift/apiserver-library-go/pkg/admission/imagepolicy"
	quotaclusterresourcequota "github.com/openshift/apiserver-library-go/pkg/admission/quota/clusterresourcequota"
	buildsecretinjector "github.com/openshift/openshift-apiserver/pkg/build/apiserver/admission/secretinjector"
	buildstrategyrestrictions "github.com/openshift/openshift-apiserver/pkg/build/apiserver/admission/strategyrestrictions"
	imageadmission "github.com/openshift/openshift-apiserver/pkg/image/apiserver/admission/limitrange"
	projectrequestlimit "github.com/openshift/openshift-apiserver/pkg/project/apiserver/admission/requestlimit"
	requiredrouteannotations "github.com/openshift/openshift-apiserver/pkg/route/apiserver/admission/requiredrouteannotations"
)

// TODO register this per apiserver or at least per process
var OriginAdmissionPlugins = admission.NewPlugins()

func init() {
	RegisterAllAdmissionPlugins(OriginAdmissionPlugins)
}

// RegisterAllAdmissionPlugins registers all admission plugins
func RegisterAllAdmissionPlugins(plugins *admission.Plugins) {
	// kube admission plugins that we rely up.  These should move to generic
	gc.Register(plugins)
	resourcequota.Register(plugins)

	genericapiserver.RegisterAllAdmissionPlugins(plugins)
	RegisterOpenshiftAdmissionPlugins(plugins)
}

func RegisterOpenshiftAdmissionPlugins(plugins *admission.Plugins) {
	projectrequestlimit.Register(plugins)
	buildsecretinjector.Register(plugins)
	buildstrategyrestrictions.Register(plugins)
	imageadmission.Register(plugins)
	imagepolicy.Register(plugins)
	quotaclusterresourcequota.Register(plugins)
	requiredrouteannotations.Register(plugins)
}

var (
	// OpenShiftAdmissionPlugins gives the in-order default admission chain for openshift resources.
	OpenShiftAdmissionPlugins = []string{
		// these are from the kbue chain
		"NamespaceLifecycle",
		"OwnerReferencesPermissionEnforcement",

		// all custom admission goes here to simulate being part of a webhook
		"project.openshift.io/ProjectRequestLimit",
		"build.openshift.io/BuildConfigSecretInjector",
		"build.openshift.io/BuildByStrategy",
		"image.openshift.io/ImageLimitRange",
		"image.openshift.io/ImagePolicy",
		"quota.openshift.io/ClusterResourceQuota",
		"route.openshift.io/RequiredRouteAnnotations",

		// the rest of the kube chain goes here
		"MutatingAdmissionWebhook",
		"ValidatingAdmissionWebhook",
		"ResourceQuota",
	}
)
