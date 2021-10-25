package ingress

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/json"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	networkingv1informers "k8s.io/client-go/informers/networking/v1"
	kv1core "k8s.io/client-go/kubernetes/typed/core/v1"
	kv1networking "k8s.io/client-go/kubernetes/typed/networking/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	networkingv1listers "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	routev1 "github.com/openshift/api/route/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	routeinformers "github.com/openshift/client-go/route/informers/externalversions/route/v1"
	routelisters "github.com/openshift/client-go/route/listers/route/v1"
)

// Controller ensures that zero or more routes exist to match any supported ingress. The
// controller creates a controller owner reference from the route to the parent ingress,
// allowing users to orphan their ingress. All owned routes have specific spec fields
// managed (those attributes present on the ingress), while any other fields may be
// modified by the user.
//
// Invariants:
//
// 1. For every ingress path rule with a non-empty backend statement, a route should
//    exist that points to that backend.
// 2. For every TLS hostname that has a corresponding path rule and points to a secret
//    that exists, a route should exist with a valid TLS config from that secret.
// 3. For every service referenced by the ingress path rule, the route should have
//    a target port based on the service.
// 4. A route owned by an ingress that is not described by any of the three invariants
//    above should be deleted.
//
// The controller also relies on the use of expectations to remind itself whether there
// are route creations it has not yet observed, which prevents the controller from
// creating more objects than it needs. The expectations are reset when the ingress
// object is modified. It is possible that expectations could leak if an ingress is
// deleted and its deletion is not observed by the cache, but such leaks are only expected
// if there is a bug in the informer cache which must be fixed anyway.
//
// Unsupported attributes:
//
// * the ingress class attribute
// * nginx annotations
// * the empty backend
// * paths with empty backends
// * creating a dynamic route spec.host
//
type Controller struct {
	eventRecorder record.EventRecorder

	routeClient   routeclient.RoutesGetter
	ingressClient kv1networking.IngressesGetter

	ingressLister      networkingv1listers.IngressLister
	ingressclassLister networkingv1listers.IngressClassLister
	secretLister       corelisters.SecretLister
	routeLister        routelisters.RouteLister
	serviceLister      corelisters.ServiceLister

	// syncs are the items that must return true before the queue can be processed
	syncs []cache.InformerSynced

	// queue is the list of namespace keys that must be synced.
	queue workqueue.RateLimitingInterface

	// expectations track upcoming route creations that we have not yet observed
	expectations *expectations
	// expectationDelay controls how long the controller waits to observe its
	// own creates. Exposed only for testing.
	expectationDelay time.Duration
}

// expectations track an upcoming change to a named resource related
// to an ingress. This is a thread safe object but callers assume
// responsibility for ensuring expectations do not leak.
type expectations struct {
	lock   sync.Mutex
	expect map[queueKey]sets.String
}

// newExpectations returns a tracking object for upcoming events
// that the controller may expect to happen.
func newExpectations() *expectations {
	return &expectations{
		expect: make(map[queueKey]sets.String),
	}
}

// Expect that an event will happen in the future for the given ingress
// and a named resource related to that ingress.
func (e *expectations) Expect(namespace, ingressName, name string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	key := queueKey{namespace: namespace, name: ingressName}
	set, ok := e.expect[key]
	if !ok {
		set = sets.NewString()
		e.expect[key] = set
	}
	set.Insert(name)
}

// Satisfied clears the expectation for the given resource name on an
// ingress.
func (e *expectations) Satisfied(namespace, ingressName, name string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	key := queueKey{namespace: namespace, name: ingressName}
	set := e.expect[key]
	set.Delete(name)
	if set.Len() == 0 {
		delete(e.expect, key)
	}
}

// Expecting returns true if the provided ingress is still waiting to
// see changes.
func (e *expectations) Expecting(namespace, ingressName string) bool {
	e.lock.Lock()
	defer e.lock.Unlock()
	key := queueKey{namespace: namespace, name: ingressName}
	return e.expect[key].Len() > 0
}

// Clear indicates that all expectations for the given ingress should
// be cleared.
func (e *expectations) Clear(namespace, ingressName string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	key := queueKey{namespace: namespace, name: ingressName}
	delete(e.expect, key)
}

type queueKey struct {
	namespace string
	name      string
}

// NewController instantiates a Controller
func NewController(eventsClient kv1core.EventsGetter, routeClient routeclient.RoutesGetter, ingressClient kv1networking.IngressesGetter, ingresses networkingv1informers.IngressInformer, ingressclasses networkingv1informers.IngressClassInformer, secrets coreinformers.SecretInformer, services coreinformers.ServiceInformer, routes routeinformers.RouteInformer) *Controller {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	// TODO: remove the wrapper when every clients have moved to use the clientset.
	broadcaster.StartRecordingToSink(&kv1core.EventSinkImpl{Interface: eventsClient.Events("")})
	recorder := broadcaster.NewRecorder(legacyscheme.Scheme, corev1.EventSource{Component: "ingress-to-route-controller"})

	c := &Controller{
		eventRecorder: recorder,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingress-to-route"),

		expectations:     newExpectations(),
		expectationDelay: 2 * time.Second,

		routeClient:   routeClient,
		ingressClient: ingressClient,

		ingressLister:      ingresses.Lister(),
		ingressclassLister: ingressclasses.Lister(),
		secretLister:       secrets.Lister(),
		routeLister:        routes.Lister(),
		serviceLister:      services.Lister(),

		syncs: []cache.InformerSynced{
			ingresses.Informer().HasSynced,
			secrets.Informer().HasSynced,
			routes.Informer().HasSynced,
			services.Informer().HasSynced,
		},
	}

	// any change to a secret of type TLS in the namespace
	secrets.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *corev1.Secret:
				return t.Type == corev1.SecretTypeTLS
			}
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    c.processNamespace,
			DeleteFunc: c.processNamespace,
			UpdateFunc: func(oldObj, newObj interface{}) {
				c.processNamespace(newObj)
			},
		},
	})

	// any change to a service in the namespace
	services.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.processNamespace,
		DeleteFunc: c.processNamespace,
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.processNamespace(newObj)
		},
	})

	// any change to a route that has the controller relationship to an Ingress
	routes.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *routev1.Route:
				_, ok := hasIngressOwnerRef(t.OwnerReferences)
				return ok
			}
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    c.processRoute,
			DeleteFunc: c.processRoute,
			UpdateFunc: func(oldObj, newObj interface{}) {
				c.processRoute(newObj)
			},
		},
	})

	// changes to ingresses
	ingresses.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.processIngress,
		DeleteFunc: c.processIngress,
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.processIngress(newObj)
		},
	})

	return c
}

func (c *Controller) processNamespace(obj interface{}) {
	switch t := obj.(type) {
	case metav1.Object:
		ns := t.GetNamespace()
		if len(ns) == 0 {
			utilruntime.HandleError(fmt.Errorf("object %T has no namespace", obj))
			return
		}
		c.queue.Add(queueKey{namespace: ns})
	default:
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %T", obj))
	}
}

func (c *Controller) processRoute(obj interface{}) {
	switch t := obj.(type) {
	case *routev1.Route:
		ingressName, ok := hasIngressOwnerRef(t.OwnerReferences)
		if !ok {
			return
		}
		c.expectations.Satisfied(t.Namespace, ingressName, t.Name)
		c.queue.Add(queueKey{namespace: t.Namespace, name: ingressName})
	default:
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %T", obj))
	}
}

func (c *Controller) processIngress(obj interface{}) {
	switch t := obj.(type) {
	case *networkingv1.Ingress:
		// when we see a change to an ingress, reset our expectations
		// this also allows periodic purging of the expectation list in the event
		// we miss one or more events.
		c.expectations.Clear(t.Namespace, t.Name)
		c.queue.Add(queueKey{namespace: t.Namespace, name: t.Name})
	default:
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %T", obj))
	}
}

// Run begins watching and syncing.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting controller")

	if !cache.WaitForCacheSync(stopCh, c.syncs...) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
	klog.Infof("Shutting down controller")
}

func (c *Controller) worker() {
	for c.processNext() {
	}
	klog.V(4).Infof("Worker stopped")
}

func (c *Controller) processNext() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	klog.V(5).Infof("processing %v begin", key)
	err := c.sync(key.(queueKey))
	c.handleNamespaceErr(err, key)
	klog.V(5).Infof("processing %v end", key)

	return true
}

func (c *Controller) handleNamespaceErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	klog.V(4).Infof("Error syncing %v: %v", key, err)
	c.queue.AddRateLimited(key)
}

func (c *Controller) sync(key queueKey) error {
	// sync all ingresses in the namespace
	if len(key.name) == 0 {
		ingresses, err := c.ingressLister.Ingresses(key.namespace).List(labels.Everything())
		if err != nil {
			return err
		}
		for _, ingress := range ingresses {
			c.queue.Add(queueKey{namespace: ingress.Namespace, name: ingress.Name})
		}
		return nil
	}
	// if we are waiting to observe the result of route creations, simply delay
	if c.expectations.Expecting(key.namespace, key.name) {
		c.queue.AddAfter(key, c.expectationDelay)
		klog.V(5).Infof("Ingress %s/%s has unsatisfied expectations", key.namespace, key.name)
		return nil
	}

	ingress, err := c.ingressLister.Ingresses(key.namespace).Get(key.name)
	if kerrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// If the ingress specifies an ingressclass and the ingressclass does
	// not specify openshift.io/ingress-to-route as its controller, ignore
	// the ingress.
	var ingressClassName *string
	if v, ok := ingress.Annotations["kubernetes.io/ingress.class"]; ok {
		ingressClassName = &v
	} else {
		ingressClassName = ingress.Spec.IngressClassName
	}
	if ingressClassName != nil {
		ingressclass, err := c.ingressclassLister.Get(*ingressClassName)
		if kerrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		// TODO Replace "openshift.io/ingress-to-route" with
		// routev1.IngressToRouteIngressClassControllerName once
		// openshift-controller-manager bumps openshift/api to a version
		// that defines it.
		if ingressclass.Spec.Controller != "openshift.io/ingress-to-route" {
			return nil
		}
	}

	// find all matching routes
	routes, err := c.routeLister.Routes(key.namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	old := routes[:0]
	for _, route := range routes {
		ingressName, ok := hasIngressOwnerRef(route.OwnerReferences)
		if !ok || ingressName != ingress.Name {
			continue
		}
		old = append(old, route)
	}

	// walk the ingress and identify whether any of the child routes need to be updated, deleted,
	// or created, as efficiently as possible.
	var creates, updates, matches []*routev1.Route
	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		if len(rule.Host) == 0 {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service == nil {
				// Non-Service backends are not implemented.
				continue
			}
			if len(path.Backend.Service.Name) == 0 {
				continue
			}
			if path.PathType != nil && *path.PathType == networkingv1.PathTypeExact {
				// Exact path type is not implemented.
				continue
			}

			var existing *routev1.Route
			old, existing = splitForPathAndHost(old, rule.Host, path.Path)
			if existing == nil {
				if r := newRouteForIngress(ingress, &rule, &path, c.secretLister, c.serviceLister); r != nil {
					creates = append(creates, r)
				}
				continue
			}

			if routeMatchesIngress(existing, ingress, &rule, &path, c.secretLister, c.serviceLister) {
				matches = append(matches, existing)
				continue
			}

			if r := newRouteForIngress(ingress, &rule, &path, c.secretLister, c.serviceLister); r != nil {
				// merge the relevant spec pieces
				preserveRouteAttributesFromExisting(r, existing)
				updates = append(updates, r)
			} else {
				// the route cannot be fully calculated, delete it
				old = append(old, existing)
			}
		}
	}

	var errs []error
	// add the new routes
	for _, route := range creates {
		if err := createRouteWithName(c.routeClient, ingress, route, c.expectations); err != nil {
			errs = append(errs, err)
		}
	}

	// update any existing routes in place
	for _, route := range updates {
		data, err := json.Marshal(&route.Spec)
		if err != nil {
			return err
		}
		annotations, err := json.Marshal(&route.Annotations)
		if err != nil {
			return err
		}
		ownerRefs, err := json.Marshal(&route.OwnerReferences)
		if err != nil {
			return err
		}
		data = []byte(fmt.Sprintf(`[{"op":"replace","path":"/spec","value":%s},`+
			`{"op":"replace","path":"/metadata/annotations","value":%s},`+
			`{"op":"replace","path":"/metadata/ownerReferences","value":%s}]`,
			data, annotations, ownerRefs))
		_, err = c.routeClient.Routes(route.Namespace).Patch(context.TODO(), route.Name, types.JSONPatchType, data, metav1.PatchOptions{})
		if err != nil {
			errs = append(errs, err)
		}
	}
	// purge any previously managed routes
	for _, route := range old {
		if err := c.routeClient.Routes(route.Namespace).Delete(context.TODO(), route.Name, metav1.DeleteOptions{}); err != nil && !kerrors.IsNotFound(err) {
			errs = append(errs, err)
		}
	}

	// reflect route status to ingress status
	//
	// We must preserve status that other controllers may have added, so we
	// cannot simply compute the new status from the routes associated with
	// the ingress; instead, we need to take the current status, remove
	// hosts for routes we've just deleted, and then add hosts for current
	// routes.  In sum, we compute
	// ingress.Status.LoadBalancer.Ingress[*].Hostname -
	// old[*].Status.Ingress[*].RouterCanonicalHostname +
	// matches[*].Status.Ingress[*].RouterCanonicalHostname.
	oldCanonicalHostnames := sets.NewString()
	for _, ingressIngress := range ingress.Status.LoadBalancer.Ingress {
		oldCanonicalHostnames.Insert(ingressIngress.Hostname)
	}
	newCanonicalHostnames := sets.NewString(oldCanonicalHostnames.List()...)
	for _, route := range old {
		for _, routeIngress := range route.Status.Ingress {
			for _, cond := range routeIngress.Conditions {
				if cond.Type == routev1.RouteAdmitted {
					if cond.Status == corev1.ConditionTrue {
						newCanonicalHostnames.Delete(routeIngress.RouterCanonicalHostname)
					}
					break
				}
			}
		}
	}
	for _, route := range matches {
		for _, routeIngress := range route.Status.Ingress {
			for _, cond := range routeIngress.Conditions {
				if cond.Type == routev1.RouteAdmitted {
					if cond.Status == corev1.ConditionTrue {
						newCanonicalHostnames.Insert(routeIngress.RouterCanonicalHostname)
					}
					break
				}
			}
		}
	}
	if !newCanonicalHostnames.Equal(oldCanonicalHostnames) {
		updatedIngress := ingress.DeepCopy()
		ingressIngresses := make([]corev1.LoadBalancerIngress, 0, newCanonicalHostnames.Len())
		for _, canonicalHostname := range newCanonicalHostnames.List() {
			ingressIngresses = append(ingressIngresses, corev1.LoadBalancerIngress{Hostname: canonicalHostname})
		}
		updatedIngress.Status.LoadBalancer.Ingress = ingressIngresses
		if _, err := c.ingressClient.Ingresses(key.namespace).UpdateStatus(context.TODO(), updatedIngress, metav1.UpdateOptions{}); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

// validOwnerRefAPIVersions is a set of recognized API versions for the ingress
// owner ref.
var validOwnerRefAPIVersions = sets.NewString(
	"networking.k8s.io/v1",
	"networking.k8s.io/v1beta1",
	"extensions.k8s.io/v1beta1",
)

func hasIngressOwnerRef(owners []metav1.OwnerReference) (string, bool) {
	for _, ref := range owners {
		if ref.Kind != "Ingress" || !validOwnerRefAPIVersions.Has(ref.APIVersion) || ref.Controller == nil || !*ref.Controller {
			continue
		}
		return ref.Name, true
	}
	return "", false
}

func newRouteForIngress(
	ingress *networkingv1.Ingress,
	rule *networkingv1.IngressRule,
	path *networkingv1.HTTPIngressPath,
	secretLister corelisters.SecretLister,
	serviceLister corelisters.ServiceLister,
) *routev1.Route {
	targetPort, err := targetPortForService(ingress.Namespace, path.Backend.Service, serviceLister)
	if err != nil {
		// no valid target port
		return nil
	}

	tlsSecret, hasInvalidTLSSecret := tlsSecretIfValid(ingress, rule, secretLister)
	if hasInvalidTLSSecret {
		return nil
	}

	var port *routev1.RoutePort
	if targetPort != nil {
		port = &routev1.RoutePort{TargetPort: *targetPort}
	}

	t := true
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: ingress.Name + "-",
			Namespace:    ingress.Namespace,
			Labels:       ingress.Labels,
			Annotations:  ingress.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Controller: &t, Name: ingress.Name, UID: ingress.UID},
			},
		},
		Spec: routev1.RouteSpec{
			Host: rule.Host,
			Path: path.Path,
			To: routev1.RouteTargetReference{
				Name: path.Backend.Service.Name,
			},
			Port: port,
			TLS:  tlsConfigForIngress(ingress, rule, tlsSecret),
		},
	}
}

func preserveRouteAttributesFromExisting(r, existing *routev1.Route) {
	r.Name = existing.Name
	r.GenerateName = ""
	r.Spec.To.Weight = existing.Spec.To.Weight
	if r.Spec.TLS != nil && existing.Spec.TLS != nil {
		r.Spec.TLS.CACertificate = existing.Spec.TLS.CACertificate
		r.Spec.TLS.DestinationCACertificate = existing.Spec.TLS.DestinationCACertificate
		r.Spec.TLS.InsecureEdgeTerminationPolicy = existing.Spec.TLS.InsecureEdgeTerminationPolicy
	}
}

func routeMatchesIngress(
	route *routev1.Route,
	ingress *networkingv1.Ingress,
	rule *networkingv1.IngressRule,
	path *networkingv1.HTTPIngressPath,
	secretLister corelisters.SecretLister,
	serviceLister corelisters.ServiceLister,
) bool {
	match := route.Spec.Host == rule.Host &&
		route.Spec.Path == path.Path &&
		route.Spec.To.Name == path.Backend.Service.Name &&
		route.Spec.WildcardPolicy == routev1.WildcardPolicyNone &&
		len(route.Spec.AlternateBackends) == 0 &&
		reflect.DeepEqual(route.Annotations, ingress.Annotations) &&
		route.OwnerReferences[0].APIVersion == "networking.k8s.io/v1"
	if !match {
		return false
	}

	targetPort, err := targetPortForService(ingress.Namespace, path.Backend.Service, serviceLister)
	if err != nil {
		// not valid
		return false
	}
	if targetPort == nil && route.Spec.Port != nil {
		return false
	}
	if targetPort != nil && (route.Spec.Port == nil || *targetPort != route.Spec.Port.TargetPort) {
		return false
	}

	tlsSecret, hasInvalidTLSSecret := tlsSecretIfValid(ingress, rule, secretLister)
	if hasInvalidTLSSecret {
		return false
	}
	tlsConfig := tlsConfigForIngress(ingress, rule, tlsSecret)
	if route.Spec.TLS != nil && tlsConfig != nil {
		tlsConfig.InsecureEdgeTerminationPolicy = route.Spec.TLS.InsecureEdgeTerminationPolicy
	}
	return reflect.DeepEqual(tlsConfig, route.Spec.TLS)
}

// targetPortForService returns a target port for a Route based on the given
// Ingress backend service.  If the Ingress references a Service or port that
// cannot be found, targetPortForService returns an error.  If the Ingress
// references a port that has no name, a nil value is returned for the target
// port.  Otherwise, the port name is returned as the target port.
//
// Note that an Ingress specifies a port on a Service whereas a Route specifies
// a port on an Endpoints resource.  The ports on a Service resource and the
// ports on its corresponding Endpoints resource have the same names but may
// have different numbers.  If there is only one port, it may be nameless, but
// in this case, the Route need have no port specification because omitting the
// port specification causes the Route to target every port (in this case, the
// only port) on the Endpoints resource.
func targetPortForService(namespace string, backendService *networkingv1.IngressServiceBackend, serviceLister corelisters.ServiceLister) (*intstr.IntOrString, error) {
	service, err := serviceLister.Services(namespace).Get(backendService.Name)
	if err != nil {
		// service doesn't exist yet, wait
		return nil, err
	}
	if len(backendService.Port.Name) != 0 {
		expect := backendService.Port.Name
		for _, port := range service.Spec.Ports {
			if port.Name == expect {
				targetPort := intstr.FromString(port.Name)
				return &targetPort, nil
			}
		}
	} else {
		expect := backendService.Port.Number
		for _, port := range service.Spec.Ports {
			if port.Port == expect {
				if len(port.Name) == 0 {
					return nil, nil
				}
				targetPort := intstr.FromString(port.Name)
				return &targetPort, nil
			}
		}
	}
	return nil, errors.New("no port found")
}

func splitForPathAndHost(routes []*routev1.Route, host, path string) ([]*routev1.Route, *routev1.Route) {
	for i, route := range routes {
		if route.Spec.Host == host && route.Spec.Path == path {
			last := len(routes) - 1
			routes[i], routes[last] = routes[last], route
			return routes[:last], route
		}
	}
	return routes, nil
}

func referencesSecret(ingress *networkingv1.Ingress, host string) (string, bool) {
	for _, tls := range ingress.Spec.TLS {
		for _, tlsHost := range tls.Hosts {
			if tlsHost == host {
				return tls.SecretName, true
			}
		}
	}
	return "", false
}

// createRouteWithName performs client side name generation so we can set a predictable expectation.
// If we fail multiple times in a row we will return an error.
// TODO: future optimization, check the local cache for the name first
func createRouteWithName(client routeclient.RoutesGetter, ingress *networkingv1.Ingress, route *routev1.Route, expect *expectations) error {
	base := route.GenerateName
	var lastErr error
	// only retry a limited number of times
	for i := 0; i < 3; i++ {
		if len(base) > 0 {
			route.GenerateName = ""
			route.Name = generateRouteName(base)
		}

		// Set the expectation before we talk to the server in order to
		// prevent racing with the route cache.
		expect.Expect(ingress.Namespace, ingress.Name, route.Name)

		_, err := client.Routes(route.Namespace).Create(context.TODO(), route, metav1.CreateOptions{})
		if err == nil {
			return nil
		}

		// We either collided with another randomly generated name, or another
		// error between us and the server prevented observing the success
		// of the result. In either case we are not expecting a new route. This
		// is safe because expectations are an optimization to avoid churn rather
		// than to prevent true duplicate creation.
		expect.Satisfied(ingress.Namespace, ingress.Name, route.Name)

		// if we aren't generating names (or if we got any other type of error)
		// return right away
		if len(base) == 0 || !kerrors.IsAlreadyExists(err) {
			return err
		}
		lastErr = err
	}
	return lastErr
}

const (
	maxNameLength          = 63
	randomLength           = 5
	maxGeneratedNameLength = maxNameLength - randomLength
)

func generateRouteName(base string) string {
	if len(base) > maxGeneratedNameLength {
		base = base[:maxGeneratedNameLength]
	}
	return fmt.Sprintf("%s%s", base, utilrand.String(randomLength))
}

func tlsConfigForIngress(
	ingress *networkingv1.Ingress,
	rule *networkingv1.IngressRule,
	potentiallyNilTLSSecret *corev1.Secret,
) *routev1.TLSConfig {
	if !tlsEnabled(ingress, rule, potentiallyNilTLSSecret) {
		return nil
	}
	// Edge: May have cert
	// Re-Encrypt: May have cert
	// Passthrough: Must not have cert
	terminationPolicy := terminationPolicyForIngress(ingress)
	tlsConfig := &routev1.TLSConfig{
		Termination:                   terminationPolicy,
		InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
	}
	if terminationPolicy != routev1.TLSTerminationPassthrough && potentiallyNilTLSSecret != nil {
		tlsConfig.Certificate = string(potentiallyNilTLSSecret.Data[corev1.TLSCertKey])
		tlsConfig.Key = string(potentiallyNilTLSSecret.Data[corev1.TLSPrivateKeyKey])
	}
	return tlsConfig
}

var emptyTLS = networkingv1.IngressTLS{}

func tlsEnabled(ingress *networkingv1.Ingress, rule *networkingv1.IngressRule, potentiallyNilTLSSecret *corev1.Secret) bool {
	switch ingress.Annotations[terminationPolicyAnnotationKey] {
	case string(routev1.TLSTerminationPassthrough), string(routev1.TLSTerminationReencrypt), string(routev1.TLSTerminationEdge):
		return true
	}
	if potentiallyNilTLSSecret != nil {
		return true
	}
	for _, tls := range ingress.Spec.TLS {
		if reflect.DeepEqual(tls, emptyTLS) {
			return true
		}
	}
	return false
}

func tlsSecretIfValid(ingress *networkingv1.Ingress, rule *networkingv1.IngressRule, secretLister corelisters.SecretLister) (_ *corev1.Secret, hasInvalidSecret bool) {
	name, ok := referencesSecret(ingress, rule.Host)
	if !ok {
		return nil, false
	}
	secret, err := secretLister.Secrets(ingress.Namespace).Get(name)
	if err != nil {
		// secret doesn't exist yet, wait
		return nil, true
	}
	if secret.Type != corev1.SecretTypeTLS {
		// secret is the wrong type
		return nil, true
	}
	if _, ok := secret.Data[corev1.TLSCertKey]; !ok {
		return nil, true
	}
	if _, ok := secret.Data[corev1.TLSPrivateKeyKey]; !ok {
		return nil, true
	}
	return secret, false
}

var terminationPolicyAnnotationKey = routev1.GroupName + "/termination"

func terminationPolicyForIngress(ingress *networkingv1.Ingress) routev1.TLSTerminationType {
	switch {
	case ingress.Annotations[terminationPolicyAnnotationKey] == string(routev1.TLSTerminationPassthrough):
		return routev1.TLSTerminationPassthrough
	case ingress.Annotations[terminationPolicyAnnotationKey] == string(routev1.TLSTerminationReencrypt):
		return routev1.TLSTerminationReencrypt
	default:
		return routev1.TLSTerminationEdge
	}
}
