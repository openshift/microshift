package bootstrappolicy

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/apis/apps"
	kauthenticationapi "k8s.io/kubernetes/pkg/apis/authentication"
	kauthorizationapi "k8s.io/kubernetes/pkg/apis/authorization"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/certificates"
	"k8s.io/kubernetes/pkg/apis/coordination"
	kapi "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/policy"
	"k8s.io/kubernetes/pkg/apis/rbac"
	rbacv1helpers "k8s.io/kubernetes/pkg/apis/rbac/v1"
	"k8s.io/kubernetes/pkg/apis/storage"
	"k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac/bootstrappolicy"

	oapps "github.com/openshift/api/apps"
	"github.com/openshift/api/authorization"
	"github.com/openshift/api/build"
	"github.com/openshift/api/config"
	"github.com/openshift/api/image"
	"github.com/openshift/api/network"
	"github.com/openshift/api/oauth"
	"github.com/openshift/api/project"
	"github.com/openshift/api/quota"
	"github.com/openshift/api/route"
	"github.com/openshift/api/security"
	"github.com/openshift/api/template"
	"github.com/openshift/api/user"
)

var (
	readWrite = []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"}
	read      = []string{"get", "list", "watch"}

	kapiGroup                  = kapi.GroupName
	admissionRegistrationGroup = "admissionregistration.k8s.io"
	appsGroup                  = apps.GroupName
	autoscalingGroup           = autoscaling.GroupName
	apiExtensionsGroup         = "apiextensions.k8s.io"
	eventsGroup                = "events.k8s.io"
	apiRegistrationGroup       = "apiregistration.k8s.io"
	batchGroup                 = batch.GroupName
	certificatesGroup          = certificates.GroupName
	coordinationGroup          = coordination.GroupName
	extensionsGroup            = extensions.GroupName
	networkingGroup            = "networking.k8s.io"
	nodeGroup                  = "node.k8s.io"
	policyGroup                = policy.GroupName
	rbacGroup                  = rbac.GroupName
	storageGroup               = storage.GroupName
	schedulingGroup            = "scheduling.k8s.io"
	kAuthzGroup                = kauthorizationapi.GroupName
	kAuthnGroup                = kauthenticationapi.GroupName
	discoveryGroup             = "discovery.k8s.io"

	deployGroup         = oapps.GroupName
	authzGroup          = authorization.GroupName
	buildGroup          = build.GroupName
	configGroup         = config.GroupName
	imageGroup          = image.GroupName
	networkGroup        = network.GroupName
	oauthGroup          = oauth.GroupName
	projectGroup        = project.GroupName
	quotaGroup          = quota.GroupName
	routeGroup          = route.GroupName
	securityGroup       = security.GroupName
	templateGroup       = template.GroupName
	userGroup           = user.GroupName
	legacyAuthzGroup    = ""
	legacyBuildGroup    = ""
	legacyDeployGroup   = ""
	legacyImageGroup    = ""
	legacyProjectGroup  = ""
	legacyQuotaGroup    = ""
	legacyRouteGroup    = ""
	legacyTemplateGroup = ""
	legacyUserGroup     = ""
	legacyOauthGroup    = ""
	legacyNetworkGroup  = ""
	legacySecurityGroup = ""

	userResource        = "users"
	groupResource       = "groups"
	systemUserResource  = "systemusers"
	systemGroupResource = "systemgroups"

	// discoveryRule is a rule that allows a client to discover the API resources available on this server
	discoveryRule = rbacv1.PolicyRule{
		Verbs: []string{"get"},
		NonResourceURLs: []string{
			// Server version checking
			"/version", "/version/*",

			// API discovery/negotiation
			"/api", "/api/*",
			"/apis", "/apis/*",
			"/oapi", "/oapi/*",
			"/openapi/v2",
			"/swaggerapi", "/swaggerapi/*", "/swagger.json", "/swagger-2.0.0.pb-v1",
			"/osapi", "/osapi/", // these cannot be removed until we can drop support for pre 3.1 clients
			"/.well-known", "/.well-known/oauth-authorization-server",

			// we intentionally allow all to here
			"/",
		},
	}

	// serviceBrokerRoot is the API root of the template service broker.
	serviceBrokerRoot = "/brokers/template.openshift.io"

	// openShiftDescription is a common, optional annotation that stores the description for a resource.
	openShiftDescription = "openshift.io/description"
)

func GetOpenshiftBootstrapClusterRoles() []rbacv1.ClusterRole {
	// four resource can be a single line
	// up to ten-ish resources per line otherwise
	clusterRoles := []rbacv1.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: SudoerRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("impersonate").Groups(userGroup, legacyUserGroup).Resources(systemUserResource, userResource).Names(SystemAdminUsername).RuleOrDie(),
				rbacv1helpers.NewRule("impersonate").Groups(userGroup, legacyUserGroup).Resources(systemGroupResource, groupResource).Names(MastersGroup).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ScopeImpersonationRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("impersonate").Groups(kAuthnGroup).Resources("userextras/scopes.authorization.openshift.io").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ClusterReaderRoleName,
			},
			AggregationRule: &rbacv1.AggregationRule{
				ClusterRoleSelectors: []metav1.LabelSelector{
					{
						MatchLabels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-cluster-reader": "true"},
					},
					{
						MatchLabels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-view": "true"},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   AggregatedClusterReaderRoleName,
				Labels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-cluster-reader": "true"},
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule(read...).Groups(kapiGroup).Resources("componentstatuses", "nodes", "nodes/status", "persistentvolumeclaims/status", "persistentvolumes",
					"persistentvolumes/status", "pods/binding", "pods/eviction", "podtemplates", "securitycontextconstraints", "services/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(admissionRegistrationGroup).Resources("mutatingwebhookconfigurations", "validatingwebhookconfigurations").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(appsGroup).Resources("statefulsets/status", "deployments/status", "controllerrevisions", "daemonsets/status",
					"replicasets/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(apiExtensionsGroup).Resources("customresourcedefinitions", "customresourcedefinitions/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(apiRegistrationGroup).Resources("apiservices", "apiservices/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(autoscalingGroup).Resources("horizontalpodautoscalers/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(batchGroup).Resources("jobs/status", "cronjobs/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(coordinationGroup).Resources("leases").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(eventsGroup).Resources("events").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(networkingGroup).Resources("ingresses/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(nodeGroup).Resources("runtimeclasses").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(policyGroup).Resources("podsecuritypolicies", "poddisruptionbudgets/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(rbacGroup).Resources("roles", "rolebindings", "clusterroles", "clusterrolebindings").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(storageGroup).Resources("csidrivers", "csinodes", "storageclasses", "volumeattachments", "volumeattachments/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(schedulingGroup).Resources("priorityclasses").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(certificatesGroup).Resources("certificatesigningrequests", "certificatesigningrequests/approval",
					"certificatesigningrequests/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(authzGroup, legacyAuthzGroup).Resources("clusterroles", "clusterrolebindings", "roles", "rolebindings",
					"rolebindingrestrictions").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("builds/details").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("images", "imagesignatures").RuleOrDie(),
				// pull images
				rbacv1helpers.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(oauthGroup, legacyOauthGroup).Resources("oauthclientauthorizations").RuleOrDie(),

				// "get" comes in from aggregate-to-view role
				rbacv1helpers.NewRule("list", "watch").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(projectGroup, legacyProjectGroup).Resources("projectrequests").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(quotaGroup, legacyQuotaGroup).Resources("clusterresourcequotas", "clusterresourcequotas/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(networkGroup, legacyNetworkGroup).Resources("clusternetworks", "egressnetworkpolicies", "hostsubnets",
					"netnamespaces").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(securityGroup, legacySecurityGroup).Resources("securitycontextconstraints").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(securityGroup).Resources("rangeallocations").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(templateGroup, legacyTemplateGroup).Resources("brokertemplateinstances", "templateinstances/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(userGroup, legacyUserGroup).Resources("groups", "identities", "useridentitymappings", "users").RuleOrDie(),

				// permissions to check access.  These creates are non-mutating
				rbacv1helpers.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("localresourceaccessreviews", "localsubjectaccessreviews",
					"resourceaccessreviews", "selfsubjectrulesreviews", "subjectrulesreviews", "subjectaccessreviews").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(kAuthzGroup).Resources("selfsubjectaccessreviews", "subjectaccessreviews", "selfsubjectrulesreviews",
					"localsubjectaccessreviews").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(kAuthnGroup).Resources("tokenreviews").RuleOrDie(),
				// permissions to check PSP, these creates are non-mutating
				rbacv1helpers.NewRule("create").Groups(securityGroup, legacySecurityGroup).Resources("podsecuritypolicysubjectreviews",
					"podsecuritypolicyselfsubjectreviews", "podsecuritypolicyreviews").RuleOrDie(),
				// Allow read access to node metrics
				rbacv1helpers.NewRule("get").Groups(kapiGroup).Resources("nodes/"+NodeMetricsSubresource, "nodes/"+NodeSpecSubresource).RuleOrDie(),
				// Allow read access to stats
				// Node stats requests are submitted as POSTs.  These creates are non-mutating
				rbacv1helpers.NewRule("get", "create").Groups(kapiGroup).Resources("nodes/" + NodeStatsSubresource).RuleOrDie(),

				rbacv1helpers.NewRule("get").URLs(rbac.NonResourceAll).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ClusterDebuggerRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("get").URLs("/metrics", "/debug/pprof", "/debug/pprof/*").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: BuildStrategyDockerRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources(DockerBuildResource, OptimizedDockerBuildResource).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: BuildStrategyCustomRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources(CustomBuildResource).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: BuildStrategySourceRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources(SourceBuildResource).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: BuildStrategyJenkinsPipelineRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources(JenkinsPipelineBuildResource).RuleOrDie(),
			},
		},
		{
			// Aggregated: cluster-csi-snapshot-controller-operator needs to grant storage-admin permissions to manage
			// VolumeSnapshotContents (=~PV) and VolumeSnapshotClasses (=~StorageClass).
			ObjectMeta: metav1.ObjectMeta{
				Name: StorageAdminRoleName,
			},
			AggregationRule: &rbacv1.AggregationRule{
				ClusterRoleSelectors: []metav1.LabelSelector{
					{
						MatchLabels: map[string]string{"storage.openshift.io/aggregate-to-storage-admin": "true"},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   AggregatedStorageAdminRoleName,
				Labels: map[string]string{"storage.openshift.io/aggregate-to-storage-admin": "true"},
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule(readWrite...).Groups(kapiGroup).Resources("persistentvolumes").RuleOrDie(),
				rbacv1helpers.NewRule(readWrite...).Groups(storageGroup).Resources("storageclasses").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(kapiGroup).Resources("persistentvolumeclaims", "events").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(kapiGroup).Resources("pods").RuleOrDie(),
			},
		},
		{
			// a role for a namespace level admin.  It is `edit` plus the power to grant permissions to other users.
			ObjectMeta: metav1.ObjectMeta{Name: AggregatedAdminRoleName, Labels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-admin": "true"}},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule(readWrite...).Groups(authzGroup, legacyAuthzGroup).Resources("roles", "rolebindings").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("localresourceaccessreviews", "localsubjectaccessreviews", "subjectrulesreviews").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(securityGroup, legacySecurityGroup).Resources("podsecuritypolicysubjectreviews", "podsecuritypolicyselfsubjectreviews", "podsecuritypolicyreviews").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(authzGroup, legacyAuthzGroup).Resources("rolebindingrestrictions").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(buildGroup, legacyBuildGroup).Resources("builds", "buildconfigs", "buildconfigs/webhooks").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("builds/log").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources("buildconfigs/instantiate", "buildconfigs/instantiatebinary", "builds/clone").RuleOrDie(),
				rbacv1helpers.NewRule("update").Groups(buildGroup, legacyBuildGroup).Resources("builds/details").RuleOrDie(),
				// access to jenkins.  multiple values to ensure that covers relationships
				rbacv1helpers.NewRule("admin", "edit", "view").Groups(build.GroupName).Resources("jenkins").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs", "deploymentconfigs/scale").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigrollbacks", "deploymentconfigs/rollback", "deploymentconfigs/instantiate").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs/log", "deploymentconfigs/status").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams", "imagestreammappings", "imagestreamtags", "imagetags", "imagestreamimages", "imagestreams/secrets").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams/status").RuleOrDie(),
				// push and pull images
				rbacv1helpers.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreamimports").RuleOrDie(),

				rbacv1helpers.NewRule("get", "patch", "update", "delete").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(quotaGroup, legacyQuotaGroup).Resources("appliedclusterresourcequotas").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(routeGroup, legacyRouteGroup).Resources("routes").RuleOrDie(),
				// admins can create routes with custom hosts
				rbacv1helpers.NewRule("create").Groups(routeGroup, legacyRouteGroup).Resources("routes/custom-host").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(routeGroup, legacyRouteGroup).Resources("routes/status").RuleOrDie(),
				// an admin can run routers that write back conditions to the route
				rbacv1helpers.NewRule("update").Groups(routeGroup, legacyRouteGroup).Resources("routes/status").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(templateGroup, legacyTemplateGroup).Resources("templates", "templateconfigs", "processedtemplates", "templateinstances").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(networkingGroup).Resources("networkpolicies").RuleOrDie(),

				// backwards compatibility
				rbacv1helpers.NewRule(readWrite...).Groups(buildGroup, legacyBuildGroup).Resources("buildlogs").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(kapiGroup).Resources("resourcequotausages").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("resourceaccessreviews", "subjectaccessreviews").RuleOrDie(),
			},
		},
		{
			// a role for a namespace level editor.  It grants access to all user level actions in a namespace.
			// It does not grant powers for "privileged" resources which are domain of the system: `/status`
			// subresources or `quota`/`limits` which are used to control namespaces
			ObjectMeta: metav1.ObjectMeta{Name: AggregatedEditRoleName, Labels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-edit": "true"}},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule(readWrite...).Groups(buildGroup, legacyBuildGroup).Resources("builds", "buildconfigs", "buildconfigs/webhooks").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("builds/log").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources("buildconfigs/instantiate", "buildconfigs/instantiatebinary", "builds/clone").RuleOrDie(),
				rbacv1helpers.NewRule("update").Groups(buildGroup, legacyBuildGroup).Resources("builds/details").RuleOrDie(),
				// access to jenkins.  multiple values to ensure that covers relationships
				rbacv1helpers.NewRule("edit", "view").Groups(buildGroup).Resources("jenkins").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs", "deploymentconfigs/scale").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigrollbacks", "deploymentconfigs/rollback", "deploymentconfigs/instantiate").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs/log", "deploymentconfigs/status").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams", "imagestreammappings", "imagestreamtags", "imagetags", "imagestreamimages", "imagestreams/secrets").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams/status").RuleOrDie(),
				// push and pull images
				rbacv1helpers.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreamimports").RuleOrDie(),

				rbacv1helpers.NewRule("get").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(quotaGroup, legacyQuotaGroup).Resources("appliedclusterresourcequotas").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(routeGroup, legacyRouteGroup).Resources("routes").RuleOrDie(),
				// editors can create routes with custom hosts
				rbacv1helpers.NewRule("create").Groups(routeGroup, legacyRouteGroup).Resources("routes/custom-host").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(routeGroup, legacyRouteGroup).Resources("routes/status").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(templateGroup, legacyTemplateGroup).Resources("templates", "templateconfigs", "processedtemplates", "templateinstances").RuleOrDie(),

				rbacv1helpers.NewRule(readWrite...).Groups(networkingGroup).Resources("networkpolicies").RuleOrDie(),

				// backwards compatibility
				rbacv1helpers.NewRule(readWrite...).Groups(buildGroup, legacyBuildGroup).Resources("buildlogs").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(kapiGroup).Resources("resourcequotausages").RuleOrDie(),
			},
		},
		{
			// a role for namespace level viewing.  It grants Read-only access to non-escalating resources in
			// a namespace.
			ObjectMeta: metav1.ObjectMeta{Name: AggregatedViewRoleName, Labels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-view": "true"}},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("builds", "buildconfigs", "buildconfigs/webhooks").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("builds/log").RuleOrDie(),
				// access to jenkins
				rbacv1helpers.NewRule("view").Groups(buildGroup).Resources("jenkins").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs", "deploymentconfigs/scale").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs/log", "deploymentconfigs/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams", "imagestreammappings", "imagestreamtags", "imagetags", "imagestreamimages").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams/status").RuleOrDie(),
				// TODO let them pull images?
				// pull images
				// rbacv1helpers.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),

				rbacv1helpers.NewRule("get").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(quotaGroup, legacyQuotaGroup).Resources("appliedclusterresourcequotas").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(routeGroup, legacyRouteGroup).Resources("routes").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(routeGroup, legacyRouteGroup).Resources("routes/status").RuleOrDie(),

				rbacv1helpers.NewRule(read...).Groups(templateGroup, legacyTemplateGroup).Resources("templates", "templateconfigs", "processedtemplates", "templateinstances").RuleOrDie(),

				// backwards compatibility
				rbacv1helpers.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("buildlogs").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(kapiGroup).Resources("resourcequotausages").RuleOrDie(),
			},
		},
		{
			// Aggregated: cluster-csi-snapshot-controller-operator needs to grant basic-user permissions to read
			// non-namespaced VolumeSnapshotClasses.
			ObjectMeta: metav1.ObjectMeta{
				Name: BasicUserRoleName,
				Annotations: map[string]string{
					openShiftDescription: "A user that can get basic information about projects.",
				},
			},
			AggregationRule: &rbacv1.AggregationRule{
				ClusterRoleSelectors: []metav1.LabelSelector{
					{
						MatchLabels: map[string]string{"authorization.openshift.io/aggregate-to-basic-user": "true"},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   AggregatedBasicUserRoleName,
				Labels: map[string]string{"authorization.openshift.io/aggregate-to-basic-user": "true"},
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("get").Groups(userGroup, legacyUserGroup).Resources("users").Names("~").RuleOrDie(),
				rbacv1helpers.NewRule("list").Groups(projectGroup, legacyProjectGroup).Resources("projectrequests").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list").Groups(authzGroup, legacyAuthzGroup).Resources("clusterroles").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(rbacGroup).Resources("clusterroles").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list").Groups(storageGroup).Resources("storageclasses").RuleOrDie(),
				rbacv1helpers.NewRule("list", "watch").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("selfsubjectrulesreviews").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(kAuthzGroup).Resources("selfsubjectaccessreviews").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: SelfAccessReviewerRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("selfsubjectrulesreviews").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(kAuthzGroup).Resources("selfsubjectaccessreviews").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: SelfProvisionerRoleName,
				Annotations: map[string]string{
					openShiftDescription: "A user that can request projects.",
				},
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("create").Groups(projectGroup, legacyProjectGroup).Resources("projectrequests").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: StatusCheckerRoleName,
				Annotations: map[string]string{
					openShiftDescription: "A user that can get basic cluster status information.",
				},
			},
			Rules: []rbacv1.PolicyRule{
				// Health
				rbacv1helpers.NewRule("get").URLs("/healthz", "/healthz/").RuleOrDie(),
				discoveryRule,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ImageAuditorRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("get", "list", "watch", "patch", "update").Groups(imageGroup, legacyImageGroup).Resources("images").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ImagePullerRoleName,
				Annotations: map[string]string{
					openShiftDescription: "Grants the right to pull images from within a project.",
				},
			},
			Rules: []rbacv1.PolicyRule{
				// pull images
				rbacv1helpers.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
			},
		},
		{
			// This role looks like a duplicate of ImageBuilderRole, but the ImageBuilder role is specifically for our builder service accounts
			// if we found another permission needed by them, we'd add it there so the intent is different if you used the ImageBuilderRole
			// you could end up accidentally granting more permissions than you intended.  This is intended to only grant enough powers to
			// push an image to our registry
			ObjectMeta: metav1.ObjectMeta{
				Name: ImagePusherRoleName,
				Annotations: map[string]string{
					openShiftDescription: "Grants the right to push and pull images from within a project.",
				},
			},
			Rules: []rbacv1.PolicyRule{
				// push and pull images
				rbacv1helpers.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ImageBuilderRoleName,
				Annotations: map[string]string{
					openShiftDescription: "Grants the right to build, push and pull images from within a project.  Used primarily with service accounts for builds.",
				},
				// TODO after the 1.12 rebase, we don't need to separate aggregate to admin
				Labels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-admin": "true", "rbac.authorization.k8s.io/aggregate-to-edit": "true"},
			},
			Rules: []rbacv1.PolicyRule{
				// push and pull images
				rbacv1helpers.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
				// allow auto-provisioning when pushing an image that doesn't have an imagestream yet
				rbacv1helpers.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreams").RuleOrDie(),
				rbacv1helpers.NewRule("update").Groups(buildGroup, legacyBuildGroup).Resources("builds/details").RuleOrDie(),
				rbacv1helpers.NewRule("get").Groups(buildGroup, legacyBuildGroup).Resources("builds").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ImagePrunerRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("get", "list").Groups(kapiGroup).Resources("pods", "replicationcontrollers").RuleOrDie(),
				rbacv1helpers.NewRule("list").Groups(kapiGroup).Resources("limitranges").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list").Groups(buildGroup, legacyBuildGroup).Resources("buildconfigs", "builds").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list").Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list").Groups(batchGroup).Resources("jobs").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list").Groups(batchGroup).Resources("cronjobs").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list").Groups(appsGroup).Resources("daemonsets").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list").Groups(appsGroup).Resources("deployments").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list").Groups(appsGroup).Resources("replicasets").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list").Groups(appsGroup).Resources("statefulsets").RuleOrDie(),

				rbacv1helpers.NewRule("delete").Groups(imageGroup, legacyImageGroup).Resources("images").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list", "watch").Groups(imageGroup, legacyImageGroup).Resources("images", "imagestreams").RuleOrDie(),
				rbacv1helpers.NewRule("update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/status").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ImageSignerRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("images", "imagestreams/layers").RuleOrDie(),
				rbacv1helpers.NewRule("create", "delete").Groups(imageGroup, legacyImageGroup).Resources("imagesignatures").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: DeployerRoleName,
				Annotations: map[string]string{
					openShiftDescription: "Grants the right to deploy within a project.  Used primarily with service accounts for automated deployments.",
				},
			},
			Rules: []rbacv1.PolicyRule{
				// "delete" is required here for compatibility with older deployer images
				// (see https://github.com/openshift/origin/pull/14322#issuecomment-303968976)
				// TODO: remove "delete" rule few releases after 3.6
				rbacv1helpers.NewRule("delete").Groups(kapiGroup).Resources("replicationcontrollers").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list", "watch", "update").Groups(kapiGroup).Resources("replicationcontrollers").RuleOrDie(),
				rbacv1helpers.NewRule("get", "update").Groups(kapiGroup).Resources("replicationcontrollers/scale").RuleOrDie(),
				rbacv1helpers.NewRule("get", "list", "watch", "create").Groups(kapiGroup).Resources("pods").RuleOrDie(),
				rbacv1helpers.NewRule("get").Groups(kapiGroup).Resources("pods/log").RuleOrDie(),
				rbacv1helpers.NewRule("create", "list").Groups(kapiGroup).Resources("events").RuleOrDie(),

				rbacv1helpers.NewRule("create", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreamtags", "imagetags").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: MasterRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule(rbac.VerbAll).Groups(rbac.APIGroupAll).Resources(rbac.ResourceAll).RuleOrDie(),
				rbacv1helpers.NewRule(rbac.VerbAll).URLs(rbac.NonResourceAll).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: OAuthTokenDeleterRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("delete").Groups(oauthGroup, legacyOauthGroup).Resources("oauthaccesstokens", "oauthauthorizetokens").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: RouterRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("list", "watch").Groups(discoveryGroup).Resources("endpointslices").RuleOrDie(),
				rbacv1helpers.NewRule("list", "watch").Groups(kapiGroup).Resources("endpoints").RuleOrDie(),
				rbacv1helpers.NewRule("list", "watch").Groups(kapiGroup).Resources("services").RuleOrDie(),

				rbacv1helpers.NewRule("create").Groups(kAuthnGroup).Resources("tokenreviews").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(kAuthzGroup).Resources("subjectaccessreviews").RuleOrDie(),

				rbacv1helpers.NewRule("list", "watch").Groups(routeGroup, legacyRouteGroup).Resources("routes").RuleOrDie(),
				rbacv1helpers.NewRule("update").Groups(routeGroup, legacyRouteGroup).Resources("routes/status").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: NodeAdminRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				// Allow read-only access to the API objects
				rbacv1helpers.NewRule(read...).Groups(kapiGroup).Resources("nodes").RuleOrDie(),
				// Allow all API calls to the nodes
				rbacv1helpers.NewRule("proxy").Groups(kapiGroup).Resources("nodes").RuleOrDie(),
				rbacv1helpers.NewRule("*").Groups(kapiGroup).Resources("nodes/proxy", "nodes/"+NodeMetricsSubresource, "nodes/"+NodeSpecSubresource, "nodes/"+NodeStatsSubresource, "nodes/"+NodeLogSubresource).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: NodeReaderRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				// Allow read-only access to the API objects
				rbacv1helpers.NewRule(read...).Groups(kapiGroup).Resources("nodes").RuleOrDie(),
				// Allow read access to node metrics
				rbacv1helpers.NewRule("get").Groups(kapiGroup).Resources("nodes/"+NodeMetricsSubresource, "nodes/"+NodeSpecSubresource).RuleOrDie(),
				// Allow read access to stats
				// Node stats requests are submitted as POSTs.  These creates are non-mutating
				rbacv1helpers.NewRule("get", "create").Groups(kapiGroup).Resources("nodes/" + NodeStatsSubresource).RuleOrDie(),
				// TODO: expose other things like /healthz on the node once we figure out non-resource URL policy across systems
			},
		},

		{
			ObjectMeta: metav1.ObjectMeta{
				Name: SDNReaderRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule(read...).Groups(networkGroup, legacyNetworkGroup).Resources("egressnetworkpolicies", "hostsubnets", "netnamespaces").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(kapiGroup).Resources("nodes", "namespaces").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(networkingGroup).Resources("networkpolicies").RuleOrDie(),
				rbacv1helpers.NewRule("get").Groups(networkGroup, legacyNetworkGroup).Resources("clusternetworks").RuleOrDie(),
				rbacv1helpers.NewRule("create", "update", "patch").Groups(kapiGroup).Resources("events").RuleOrDie(),
			},
		},

		{
			ObjectMeta: metav1.ObjectMeta{
				Name: SDNManagerRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("get", "list", "watch", "create", "delete").Groups(networkGroup, legacyNetworkGroup).Resources("hostsubnets", "netnamespaces").RuleOrDie(),
				rbacv1helpers.NewRule("get", "create").Groups(networkGroup, legacyNetworkGroup).Resources("clusternetworks").RuleOrDie(),
				rbacv1helpers.NewRule(read...).Groups(kapiGroup).Resources("nodes").RuleOrDie(),
			},
		},

		{
			ObjectMeta: metav1.ObjectMeta{
				Name: WebHooksRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("get", "create").Groups(buildGroup, legacyBuildGroup).Resources("buildconfigs/webhooks").RuleOrDie(),
			},
		},

		{
			ObjectMeta: metav1.ObjectMeta{
				Name: DiscoveryRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				discoveryRule,
			},
		},

		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   RegistryAdminRoleName,
				Labels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-admin": "true"},
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule(readWrite...).Groups(kapiGroup).Resources("serviceaccounts", "secrets").RuleOrDie(),
				rbacv1helpers.NewRule(readWrite...).Groups(imageGroup, legacyImageGroup).Resources("imagestreamimages", "imagestreammappings", "imagestreams", "imagestreams/secrets", "imagestreamtags", "imagetags").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreamimports").RuleOrDie(),
				rbacv1helpers.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
				rbacv1helpers.NewRule(readWrite...).Groups(authzGroup, legacyAuthzGroup).Resources("rolebindings", "roles").RuleOrDie(),
				rbacv1helpers.NewRule(readWrite...).Groups(rbacGroup).Resources("roles", "rolebindings").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("localresourceaccessreviews", "localsubjectaccessreviews", "subjectrulesreviews").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(kAuthzGroup).Resources("localsubjectaccessreviews").RuleOrDie(),

				rbacv1helpers.NewRule("get").Groups(kapiGroup).Resources("namespaces").RuleOrDie(),
				rbacv1helpers.NewRule("get", "delete").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),

				// backwards compatibility
				rbacv1helpers.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("resourceaccessreviews", "subjectaccessreviews").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   RegistryEditorRoleName,
				Labels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-edit": "true"},
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule(readWrite...).Groups(kapiGroup).Resources("serviceaccounts", "secrets").RuleOrDie(),
				rbacv1helpers.NewRule(readWrite...).Groups(imageGroup, legacyImageGroup).Resources("imagestreamimages", "imagestreammappings", "imagestreams", "imagestreams/secrets", "imagestreamtags", "imagetags").RuleOrDie(),
				rbacv1helpers.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreamimports").RuleOrDie(),
				rbacv1helpers.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),

				rbacv1helpers.NewRule("get").Groups(kapiGroup).Resources("namespaces").RuleOrDie(),
				rbacv1helpers.NewRule("get").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   RegistryViewerRoleName,
				Labels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-view": "true"},
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("imagestreamimages", "imagestreammappings", "imagestreams", "imagestreamtags", "imagetags").RuleOrDie(),
				rbacv1helpers.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),

				rbacv1helpers.NewRule("get").Groups(kapiGroup).Resources("namespaces").RuleOrDie(),
				rbacv1helpers.NewRule("get").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: TemplateServiceBrokerClientRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("get", "put", "update", "delete").URLs(serviceBrokerRoot + "/*").RuleOrDie(),
			},
		},
	}
	for i := range clusterRoles {
		clusterRole := &clusterRoles[i]
		addDefaultMetadata(clusterRole)
	}
	return clusterRoles
}

func GetBootstrapClusterRoles() []rbacv1.ClusterRole {
	return append(GetOpenshiftBootstrapClusterRoles(), ControllerRoles()...)
}

func newOriginRoleBinding(bindingName, roleName, namespace string) *rbacv1helpers.RoleBindingBuilder {
	builder := rbacv1helpers.NewRoleBinding(roleName, namespace)
	builder.RoleBinding.Name = bindingName
	return builder
}

func newOriginClusterBinding(bindingName, roleName string) *rbacv1helpers.ClusterRoleBindingBuilder {
	builder := rbacv1helpers.NewClusterBinding(roleName)
	builder.ClusterRoleBinding.Name = bindingName
	return builder
}

func GetOpenshiftBootstrapClusterRoleBindings() []rbacv1.ClusterRoleBinding {
	clusterRoleBindings := []rbacv1.ClusterRoleBinding{
		newOriginClusterBinding(MasterRoleBindingName, MasterRoleName).
			Groups(MastersGroup).
			BindingOrDie(),
		newOriginClusterBinding(NodeAdminRoleBindingName, NodeAdminRoleName).
			Users(LegacyMasterKubeletAdminClientUsername).
			Groups(NodeAdminsGroup).
			BindingOrDie(),
		newOriginClusterBinding(ClusterAdminRoleBindingName, ClusterAdminRoleName).
			Groups(ClusterAdminGroup).
			// add system:admin to this binding so that members of the
			// sudoer group can use --as=system:admin to run a command
			// as a cluster-admin
			Users(SystemAdminUsername).
			BindingOrDie(),
		newOriginClusterBinding(ClusterReaderRoleBindingName, ClusterReaderRoleName).
			Groups(ClusterReaderGroup).
			BindingOrDie(),
		newOriginClusterBinding(BasicUserRoleBindingName, BasicUserRoleName).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(SelfAccessReviewerRoleBindingName, SelfAccessReviewerRoleName).
			Groups(AuthenticatedGroup, UnauthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(SelfProvisionerRoleBindingName, SelfProvisionerRoleName).
			Groups(AuthenticatedOAuthGroup).
			BindingOrDie(),
		newOriginClusterBinding(OAuthTokenDeleterRoleBindingName, OAuthTokenDeleterRoleName).
			Groups(AuthenticatedGroup, UnauthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(StatusCheckerRoleBindingName, StatusCheckerRoleName).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(NodeProxierRoleBindingName, "system:node-proxier").
			// Allow node identities to run node proxies
			Groups(NodesGroup).
			BindingOrDie(),
		newOriginClusterBinding(SDNReaderRoleBindingName, SDNReaderRoleName).
			// Allow node identities to run SDN plugins
			Groups(NodesGroup).
			BindingOrDie(),
		newOriginClusterBinding(WebHooksRoleBindingName, WebHooksRoleName).
			Groups(AuthenticatedGroup, UnauthenticatedGroup).
			BindingOrDie(),
		rbacv1helpers.NewClusterBinding(DiscoveryRoleName).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
		// Allow all build strategies by default.
		// These are in separate bindings so that cluster admins can remove the subjects
		// and use the annotation to prevent reconciliation on server start.
		newOriginClusterBinding(BuildStrategyDockerRoleBindingName, BuildStrategyDockerRoleName).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(BuildStrategySourceRoleBindingName, BuildStrategySourceRoleName).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(BuildStrategyJenkinsPipelineRoleBindingName, BuildStrategyJenkinsPipelineRoleName).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
		// Allow node-bootstrapper SA to bootstrap nodes by default.
		rbacv1helpers.NewClusterBinding(NodeBootstrapRoleName).
			SAs(DefaultOpenShiftInfraNamespace, InfraNodeBootstrapServiceAccountName).
			BindingOrDie(),
		// Everyone should be able to add a scope to their impersonation request.  It is purely tightening.
		// This does not grant access to impersonate in general, only tighten if you already have permission.
		rbacv1helpers.NewClusterBinding(ScopeImpersonationRoleName).
			Groups(AuthenticatedGroup, UnauthenticatedGroup).
			BindingOrDie(),
	}
	for i := range clusterRoleBindings {
		clusterRoleBinding := &clusterRoleBindings[i]
		addDefaultMetadata(clusterRoleBinding)
	}
	return clusterRoleBindings
}

func GetBootstrapClusterRoleBindings() []rbacv1.ClusterRoleBinding {
	return append(GetOpenshiftBootstrapClusterRoleBindings(), ControllerRoleBindings()...)
}

// TODO we need to remove the global mutable state from all roles / bindings so we know this function is safe to call
func addDefaultMetadata(obj runtime.Object) {
	metadata, err := meta.Accessor(obj)
	if err != nil {
		// if this happens, then some static code is broken
		panic(err)
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	for k, v := range bootstrappolicy.Annotation {
		annotations[k] = v
	}
	metadata.SetAnnotations(annotations)
}

func GetBootstrapClusterRolesToAggregate() map[string]string {
	return map[string]string{
		ClusterReaderRoleName: AggregatedClusterReaderRoleName,
		BasicUserRoleName:     AggregatedBasicUserRoleName,
		StorageAdminRoleName:  AggregatedStorageAdminRoleName,
	}
}
