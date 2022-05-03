package requiredrouteannotations

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/initializer"
	"k8s.io/client-go/informers"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	configv1 "github.com/openshift/api/config/v1"
	grouproute "github.com/openshift/api/route"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	configv1listers "github.com/openshift/client-go/config/listers/config/v1"
	routeinformers "github.com/openshift/client-go/route/informers/externalversions"
	routev1listers "github.com/openshift/client-go/route/listers/route/v1"
	openshiftapiserveradmission "github.com/openshift/openshift-apiserver/pkg/admission"
	routeapi "github.com/openshift/openshift-apiserver/pkg/route/apis/route"
)

const (
	pluginName = "route.openshift.io/RequiredRouteAnnotations"
	// To cover scenarios with 10,000 routes, need to wait up to 30 seconds for caches to sync
	timeToWaitForCacheSync = 20 * time.Second
	hstsAnnotation         = "haproxy.router.openshift.io/hsts_header"
)

func Register(plugins *admission.Plugins) {
	plugins.Register(pluginName,
		func(_ io.Reader) (admission.Interface, error) {
			return NewRequiredRouteAnnotations(), nil
		})
}

// cacheSync guards the isSynced variable
// Once isSynced is true, we don't care about setting it anymore
type cacheSync struct {
	isSyncedLock sync.RWMutex
	isSynced     bool
}

func (cs *cacheSync) hasSynced() bool {
	cs.isSyncedLock.RLock()
	defer cs.isSyncedLock.RUnlock()
	return cs.isSynced
}
func (cs *cacheSync) setSynced() {
	cs.isSyncedLock.Lock()
	defer cs.isSyncedLock.Unlock()
	cs.isSynced = true
}

type requiredRouteAnnotations struct {
	*admission.Handler
	routeLister   routev1listers.RouteLister
	nsLister      corev1listers.NamespaceLister
	ingressLister configv1listers.IngressLister
	cachesToSync  []cache.InformerSynced
	cacheSyncLock cacheSync
}

// Ensure that the required OpenShift admission interfaces are implemented.
var _ = initializer.WantsExternalKubeInformerFactory(&requiredRouteAnnotations{})
var _ = admission.ValidationInterface(&requiredRouteAnnotations{})
var _ = openshiftapiserveradmission.WantsOpenShiftConfigInformers(&requiredRouteAnnotations{})
var _ = openshiftapiserveradmission.WantsOpenShiftRouteInformers(&requiredRouteAnnotations{})

var maxAgeRegExp = regexp.MustCompile(`max-age=(\d+)`)

// Validate ensures that routes specify required annotations, and returns nil if valid.
// The admission handler ensures this is only called for Create/Update operations.
func (o *requiredRouteAnnotations) Validate(ctx context.Context, a admission.Attributes, _ admission.ObjectInterfaces) (err error) {
	if a.GetResource().GroupResource() != grouproute.Resource("routes") {
		return nil
	}
	newRoute, isRoute := a.GetObject().(*routeapi.Route)
	if !isRoute {
		return nil
	}

	// Determine if there are HSTS changes in this update
	if a.GetOperation() == admission.Update {
		wants, has := false, false
		var oldHSTS, newHSTS string

		newHSTS, wants = newRoute.Annotations[hstsAnnotation]

		oldObject := a.GetOldObject().(*routeapi.Route)
		oldHSTS, has = oldObject.Annotations[hstsAnnotation]

		// Skip the validation if we're not making a change to HSTS at this time
		if wants == has && newHSTS == oldHSTS {
			return nil
		}
	}

	// Cannot apply HSTS if route is not TLS.  Ignore silently to keep backward compatibility.
	tls := newRoute.Spec.TLS
	if tls == nil || (tls.Termination != routeapi.TLSTerminationEdge && tls.Termination != routeapi.TLSTerminationReencrypt) {
		// TODO - will address missing annotations on routes as route status in https://issues.redhat.com/browse/NE-678
		return nil
	}

	// Wait just once up to 20 seconds for all caches to sync
	if !o.waitForSyncedStore(ctx) {
		return admission.NewForbidden(a, errors.New(pluginName+": caches not synchronized"))
	}

	ingress, err := o.ingressLister.Get("cluster")
	if err != nil {
		return admission.NewForbidden(a, err)
	}

	namespace, err := o.nsLister.Get(newRoute.Namespace)
	if err != nil {
		return admission.NewForbidden(a, err)
	}

	if err = isRouteHSTSAllowed(ingress, newRoute, namespace); err != nil {
		return admission.NewForbidden(a, err)
	}
	return nil
}

func (o *requiredRouteAnnotations) SetExternalKubeInformerFactory(kubeInformers informers.SharedInformerFactory) {
	o.nsLister = kubeInformers.Core().V1().Namespaces().Lister()
	o.cachesToSync = append(o.cachesToSync, kubeInformers.Core().V1().Namespaces().Informer().HasSynced)
}

// waitForSyncedStore calls cache.WaitForCacheSync, which will wait up to timeToWaitForCacheSync
// for the cachesToSync to synchronize.
func (o *requiredRouteAnnotations) waitForSyncedStore(ctx context.Context) bool {
	syncCtx, cancelFn := context.WithTimeout(ctx, timeToWaitForCacheSync)
	defer cancelFn()
	if !o.cacheSyncLock.hasSynced() {
		if !cache.WaitForCacheSync(syncCtx.Done(), o.cachesToSync...) {
			return false
		}
		o.cacheSyncLock.setSynced()
	}
	return true
}

func (o *requiredRouteAnnotations) ValidateInitialization() error {
	if o.ingressLister == nil {
		return fmt.Errorf(pluginName + " plugin needs an ingress lister")
	}
	if o.routeLister == nil {
		return fmt.Errorf(pluginName + " plugin needs a route lister")
	}
	if o.nsLister == nil {
		return fmt.Errorf(pluginName + " plugin needs a namespace lister")
	}
	if len(o.cachesToSync) < 3 {
		return fmt.Errorf(pluginName + " plugin missing informer synced functions")
	}
	return nil
}

func NewRequiredRouteAnnotations() *requiredRouteAnnotations {
	return &requiredRouteAnnotations{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}

func (o *requiredRouteAnnotations) SetOpenShiftRouteInformers(informers routeinformers.SharedInformerFactory) {
	o.cachesToSync = append(o.cachesToSync, informers.Route().V1().Routes().Informer().HasSynced)
	o.routeLister = informers.Route().V1().Routes().Lister()
}

func (o *requiredRouteAnnotations) SetOpenShiftConfigInformers(informers configinformers.SharedInformerFactory) {
	o.cachesToSync = append(o.cachesToSync, informers.Config().V1().Ingresses().Informer().HasSynced)
	o.ingressLister = informers.Config().V1().Ingresses().Lister()
}

// isRouteHSTSAllowed returns nil if the route is allowed.  Otherwise, returns details and a suggestion in the error
func isRouteHSTSAllowed(ingress *configv1.Ingress, newRoute *routeapi.Route, namespace *corev1.Namespace) error {
	requirements := ingress.Spec.RequiredHSTSPolicies
	for _, requirement := range requirements {
		// Check if the required namespaceSelector (if any) and the domainPattern match
		if matches, err := requiredNamespaceDomainMatchesRoute(requirement, newRoute, namespace); err != nil {
			return err
		} else if !matches {
			// If one of either the namespaceSelector or domain didn't match, we will continue to look
			continue
		}

		routeHSTS, err := hstsConfigFromRoute(newRoute)
		if err != nil {
			return err
		}

		// If there is no annotation but there needs to be one, return error
		if routeHSTS != nil {
			if err = routeHSTS.meetsRequirements(requirement); err != nil {
				return err
			}
		}

		// Validation only checks the first matching required HSTS rule.
		return nil
	}

	// None of the requirements matched this route's domain/namespace, it is automatically allowed
	return nil
}

type hstsConfig struct {
	maxAge            int32
	preload           bool
	includeSubDomains bool
}

const (
	HSTSMaxAgeMissingOrWrongError     = "HSTS max-age must be set correctly in HSTS annotation"
	HSTSMaxAgeGreaterError            = "HSTS max-age is greater than maximum age %ds"
	HSTSMaxAgeLessThanError           = "HSTS max-age is less than minimum age %ds"
	HSTSPreloadMustError              = "HSTS preload must be specified"
	HSTSPreloadMustNotError           = "HSTS preload must not be specified"
	HSTSIncludeSubDomainsMustError    = "HSTS includeSubDomains must be specified"
	HSTSIncludeSubDomainsMustNotError = "HSTS includeSubDomains must not be specified"
)

// Parse out the hstsConfig fields from the annotation
// Unrecognized fields are ignored
func hstsConfigFromRoute(route *routeapi.Route) (*hstsConfig, error) {
	var ret hstsConfig

	trimmed := strings.ToLower(strings.ReplaceAll(route.Annotations[hstsAnnotation], " ", ""))
	tokens := strings.Split(trimmed, ";")
	for _, token := range tokens {
		if strings.EqualFold(token, "includeSubDomains") {
			ret.includeSubDomains = true
		}
		if strings.EqualFold(token, "preload") {
			ret.preload = true
		}
		// unrecognized tokens are ignored
	}

	if match := maxAgeRegExp.FindStringSubmatch(trimmed); match != nil && len(match) > 1 {
		age, err := strconv.ParseInt(match[1], 10, 32)
		if err != nil {
			return nil, err
		}
		ret.maxAge = int32(age)
	} else {
		return nil, fmt.Errorf(HSTSMaxAgeMissingOrWrongError)
	}

	return &ret, nil
}

// Make sure the given requirement meets the configured HSTS policy, validating:
// - range for maxAge (existence already established)
// - preloadPolicy
// - includeSubDomainsPolicy
func (c *hstsConfig) meetsRequirements(requirement configv1.RequiredHSTSPolicy) error {
	if requirement.MaxAge.LargestMaxAge != nil && c.maxAge > *requirement.MaxAge.LargestMaxAge {
		return fmt.Errorf(HSTSMaxAgeGreaterError, *requirement.MaxAge.LargestMaxAge)
	}
	if requirement.MaxAge.SmallestMaxAge != nil && c.maxAge < *requirement.MaxAge.SmallestMaxAge {
		return fmt.Errorf(HSTSMaxAgeLessThanError, *requirement.MaxAge.SmallestMaxAge)
	}

	switch requirement.PreloadPolicy {
	case configv1.NoOpinionPreloadPolicy:
	// anything is allowed, do nothing
	case configv1.RequirePreloadPolicy:
		if !c.preload {
			return fmt.Errorf(HSTSPreloadMustError)
		}
	case configv1.RequireNoPreloadPolicy:
		if c.preload {
			return fmt.Errorf(HSTSPreloadMustNotError)
		}
	}

	switch requirement.IncludeSubDomainsPolicy {
	case configv1.NoOpinionIncludeSubDomains:
	// anything is allowed, do nothing
	case configv1.RequireIncludeSubDomains:
		if !c.includeSubDomains {
			return fmt.Errorf(HSTSIncludeSubDomainsMustError)
		}
	case configv1.RequireNoIncludeSubDomains:
		if c.includeSubDomains {
			return fmt.Errorf(HSTSIncludeSubDomainsMustNotError)
		}
	}

	return nil
}

// Check if the route matches the required domain/namespace in the HSTS Policy
func requiredNamespaceDomainMatchesRoute(requirement configv1.RequiredHSTSPolicy, route *routeapi.Route, namespace *corev1.Namespace) (bool, error) {
	matchesNamespace, err := matchesNamespaceSelector(requirement.NamespaceSelector, namespace)
	if err != nil {
		return false, err
	}

	routeDomains := []string{route.Spec.Host}
	for _, ingress := range route.Status.Ingress {
		routeDomains = append(routeDomains, ingress.Host)
	}
	matchesDom := matchesDomain(requirement.DomainPatterns, routeDomains)

	return matchesNamespace && matchesDom, nil
}

// Check all of the required domainMatcher patterns against all provided domains,
// first match returns true.  If none match, return false.
func matchesDomain(domainMatchers []string, domains []string) bool {
	for _, pattern := range domainMatchers {
		for _, candidate := range domains {
			matched, err := filepath.Match(pattern, candidate)
			if err != nil {
				klog.Warningf("Ignoring HSTS Policy domain match for %s, error parsing: %v", candidate, err)
				continue
			}
			if matched {
				return true
			}
		}
	}

	return false
}

func matchesNamespaceSelector(nsSelector *metav1.LabelSelector, namespace *corev1.Namespace) (bool, error) {
	if nsSelector == nil {
		return true, nil
	}
	selector, err := getParsedNamespaceSelector(nsSelector)
	if err != nil {
		klog.Warningf("Ignoring HSTS Policy namespace match for %s, error parsing: %v", namespace, err)
		return false, err
	}
	return selector.Matches(labels.Set(namespace.Labels)), nil
}

func getParsedNamespaceSelector(nsSelector *metav1.LabelSelector) (labels.Selector, error) {
	// TODO cache this result to save time
	return metav1.LabelSelectorAsSelector(nsSelector)
}
