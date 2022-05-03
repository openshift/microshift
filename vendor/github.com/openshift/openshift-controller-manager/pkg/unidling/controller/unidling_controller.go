package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/pager"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/component-base/metrics"
	"k8s.io/klog/v2"

	"github.com/openshift/api"
	unidlingapi "github.com/openshift/api/unidling/v1alpha1"
	appstypedclient "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	"github.com/openshift/library-go/pkg/unidling/unidlingclient"
	unidlingmetrics "github.com/openshift/openshift-controller-manager/pkg/unidling/metrics"
)

const MaxRetries = 5

type lastFiredCache struct {
	sync.RWMutex
	items map[types.NamespacedName]time.Time
}

func (c *lastFiredCache) Get(info types.NamespacedName) time.Time {
	c.RLock()
	defer c.RUnlock()

	return c.items[info]
}

func (c *lastFiredCache) Clear(info types.NamespacedName) {
	c.Lock()
	defer c.Unlock()

	delete(c.items, info)
}

func (c *lastFiredCache) AddIfNewer(info types.NamespacedName, newLastFired time.Time) bool {
	c.Lock()
	defer c.Unlock()

	if lastFired, hasLastFired := c.items[info]; !hasLastFired || lastFired.Before(newLastFired) {
		c.items[info] = newLastFired
		return true
	}

	return false
}

type UnidlingController struct {
	controller          cache.Controller
	scaleNamespacer     scale.ScalesGetter
	mapper              meta.RESTMapper
	endpointsNamespacer corev1client.EndpointsGetter
	servicesNamespacer  corev1client.ServicesGetter
	queue               workqueue.RateLimitingInterface
	lastFiredCache      *lastFiredCache
	eventsTotal         *metrics.Counter

	// TODO: remove these once we get the scale-source functionality in the scale endpoints
	dcNamespacer appstypedclient.DeploymentConfigsGetter
	rcNamespacer corev1client.ReplicationControllersGetter
}

func NewUnidlingController(scaleNS scale.ScalesGetter, mapper meta.RESTMapper, endptsNS corev1client.EndpointsGetter, servicesNS corev1client.ServicesGetter, evtNS corev1client.EventsGetter,
	dcNamespacer appstypedclient.DeploymentConfigsGetter, rcNamespacer corev1client.ReplicationControllersGetter,
	resyncPeriod time.Duration) *UnidlingController {
	fieldSet := fields.Set{}
	fieldSet["reason"] = unidlingapi.NeedPodsReason
	fieldSelector := fieldSet.AsSelector()

	unidlingController := &UnidlingController{
		scaleNamespacer:     scaleNS,
		mapper:              mapper,
		endpointsNamespacer: endptsNS,
		servicesNamespacer:  servicesNS,
		queue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "unidling"),
		lastFiredCache: &lastFiredCache{
			items: make(map[types.NamespacedName]time.Time),
		},
		eventsTotal: unidlingmetrics.GetEventsTotalCounter(),

		dcNamespacer: dcNamespacer,
		rcNamespacer: rcNamespacer,
	}

	_, controller := cache.NewInformer(
		&cache.ListWatch{
			// https://bugzilla.redhat.com/show_bug.cgi?id=1887745 -- making sure the ListFunc is
			// using chunking for bootstrap and resync periods, meaning resourceVersion=0 is replaced
			// by empty string, in order to have limits respected. Consider:
			//    https://github.com/kubernetes/apiserver/blob/0e46f0ea2bdc958827712205cf9a88140a28409d/pkg/storage/cacher/cacher.go#L655-L668
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				fn := func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
					return evtNS.Events(metav1.NamespaceAll).List(ctx, opts)
				}
				p := pager.New(pager.ListPageFunc(fn))

				options.FieldSelector = fieldSelector.String()
				options.Limit = 500

				// replacing rv=0 with empty string to make sure chunking takes place, using it as
				// zero makes the apiserver to return results from cache
				if options.ResourceVersion == "0" {
					options.ResourceVersion = ""
				}

				// TODO: wire a global context through the controller.
				list, _, err := p.List(context.TODO(), options)
				return list, err
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.FieldSelector = fieldSelector.String()
				return evtNS.Events(metav1.NamespaceAll).Watch(context.TODO(), options)
			},
		},
		&corev1.Event{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    unidlingController.addEvent,
			UpdateFunc: unidlingController.updateEvent,
			// this is just to clean up our cache of the last seen times
			DeleteFunc: unidlingController.checkAndClearFromCache,
		},
	)

	unidlingController.controller = controller

	return unidlingController
}

func (c *UnidlingController) addEvent(obj interface{}) {
	evt, ok := obj.(*corev1.Event)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("got non-Event object in event action: %v", obj))
		return
	}

	c.enqueueEvent(evt)
}

func (c *UnidlingController) updateEvent(oldObj, newObj interface{}) {
	evt, ok := newObj.(*corev1.Event)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("got non-Event object in event action: %v", newObj))
		return
	}

	c.enqueueEvent(evt)
}

func (c *UnidlingController) checkAndClearFromCache(obj interface{}) {
	evt, objIsEvent := obj.(*corev1.Event)
	if !objIsEvent {
		tombstone, objIsTombstone := obj.(cache.DeletedFinalStateUnknown)
		if !objIsTombstone {
			utilruntime.HandleError(fmt.Errorf("got non-event, non-tombstone object in event action: %v", obj))
			return
		}

		evt, objIsEvent = tombstone.Obj.(*corev1.Event)
		if !objIsEvent {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not an Event in event action: %v", obj))
			return
		}
	}

	c.clearEventFromCache(evt)
}

// clearEventFromCache removes the entry for the given event from the lastFiredCache.
func (c *UnidlingController) clearEventFromCache(event *corev1.Event) {
	if event.Reason != unidlingapi.NeedPodsReason {
		return
	}

	info := types.NamespacedName{
		Namespace: event.InvolvedObject.Namespace,
		Name:      event.InvolvedObject.Name,
	}
	c.lastFiredCache.Clear(info)
}

// equeueEvent checks if the given event is relevant (i.e. if it's a NeedPods event),
// and, if so, extracts relevant information, and enqueues that information in the
// processing queue.
func (c *UnidlingController) enqueueEvent(event *corev1.Event) {
	if event.Reason != unidlingapi.NeedPodsReason {
		return
	}

	info := types.NamespacedName{
		Namespace: event.InvolvedObject.Namespace,
		Name:      event.InvolvedObject.Name,
	}

	// only add things to the queue if they're newer than what we already have
	if c.lastFiredCache.AddIfNewer(info, event.LastTimestamp.Time) {
		c.eventsTotal.Inc()
		c.queue.Add(info)
	}
}

func (c *UnidlingController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	go c.controller.Run(stopCh)
	go wait.Until(c.processRequests, time.Second, stopCh)
}

// processRequests calls awaitRequest repeatedly, until told to stop by
// the return value of awaitRequest.
func (c *UnidlingController) processRequests() {
	for {
		if !c.awaitRequest() {
			return
		}
	}
}

// awaitRequest awaits a new request on the queue, and sends it off for processing.
// If more requests on the queue should be processed, it returns true.  If we should
// stop processing, it returns false.
func (c *UnidlingController) awaitRequest() bool {
	infoRaw, stop := c.queue.Get()
	if stop {
		return false
	}

	defer c.queue.Done(infoRaw)

	info := infoRaw.(types.NamespacedName)
	lastFired := c.lastFiredCache.Get(info)

	var retry bool
	var err error
	if retry, err = c.handleRequest(info, lastFired); err == nil {
		// if there was no error, we succeeded in the unidling, and we need to
		// tell the rate limitter to stop tracking this request
		c.queue.Forget(infoRaw)
		return true
	}

	// check to see if we think the error was transient (e.g. server error on the update request),
	// and if not, do not retry
	if !retry {
		utilruntime.HandleError(fmt.Errorf("Unable to process unidling event for %s/%s at (%s), will not retry: %v", info.Namespace, info.Name, lastFired, err))
		return true
	}

	// Otherwise, if we have an error, we were at least partially unsuccessful in unidling, so
	// we requeue the event to process later

	// don't try to process failing requests forever
	if c.queue.NumRequeues(infoRaw) > MaxRetries {
		utilruntime.HandleError(fmt.Errorf("Unable to process unidling event for %s/%s (at %s), will not retry again: %v", info.Namespace, info.Name, lastFired, err))
		c.queue.Forget(infoRaw)
		return true
	}

	klog.V(4).Infof("Unable to fully process unidling request for %s/%s (at %s), will retry: %v", info.Namespace, info.Name, lastFired, err)
	c.queue.AddRateLimited(infoRaw)
	return true
}

// handleRequest handles a single request to unidle.  After checking the validity of the request,
// it will examine the service in question to determine which scalables to scale, and will scale
// them and remove them from the services' list of idled scalables.  If it is unable to properly
// process the request, it will return a boolean indicating whether or not we should retry later,
// as well as an error (e.g. if we're unable to parse an annotation, retrying later won't help,
// so it will return false).
func (c *UnidlingController) handleRequest(info types.NamespacedName, lastFired time.Time) (bool, error) {
	// fetch the service in question
	targetService, err := c.servicesNamespacer.Services(info.Namespace).Get(context.TODO(), info.Name, metav1.GetOptions{})
	if err != nil {
		return true, fmt.Errorf("unable to retrieve service: %v", err)
	}

	// make sure we actually were idled...
	idledTimeRaw, wasIdled := targetService.Annotations[unidlingapi.IdledAtAnnotation]
	if !wasIdled {
		klog.V(5).Infof("UnidlingController received a NeedPods event for a service that was not idled, ignoring")
		return false, nil
	}

	// ...and make sure this request was to wake up from the most recent idling, and not a previous one
	idledTime, err := time.Parse(time.RFC3339, idledTimeRaw)
	if err != nil {
		// retrying here won't help, we're just stuck as idle since we can't get parse the idled time
		return false, fmt.Errorf("unable to check idled-at time: %v", err)
	}
	if lastFired.Before(idledTime) {
		klog.V(5).Infof("UnidlingController received an out-of-date NeedPods event, ignoring")
		return false, nil
	}

	// TODO: ew, this is metav1.  Such is life when working with annotations.
	var targetScalables []unidlingapi.RecordedScaleReference
	if targetScalablesStr, hasTargetScalables := targetService.Annotations[unidlingapi.UnidleTargetAnnotation]; hasTargetScalables {
		if err = json.Unmarshal([]byte(targetScalablesStr), &targetScalables); err != nil {
			// retrying here won't help, we're just stuck as idled since we can't parse the idled scalables list
			return false, fmt.Errorf("unable to unmarshal target scalable references: %v", err)
		}
	} else {
		klog.V(4).Infof("Service %s/%s had no scalables to unidle", info.Namespace, info.Name)
		targetScalables = []unidlingapi.RecordedScaleReference{}
	}

	targetScalablesSet := make(map[unidlingapi.RecordedScaleReference]struct{}, len(targetScalables))
	for _, v := range targetScalables {
		targetScalablesSet[v] = struct{}{}
	}

	deleteIdlingAnnotations := func(_ int32, annotations map[string]string) {
		delete(annotations, unidlingapi.IdledAtAnnotation)
		delete(annotations, unidlingapi.PreviousScaleAnnotation)
	}

	scaleAnnotater := unidlingclient.NewScaleAnnotater(c.scaleNamespacer, c.mapper, c.dcNamespacer, c.rcNamespacer, deleteIdlingAnnotations)

	for _, scalableRef := range targetScalables {
		var scale *autoscalingv1.Scale
		var obj runtime.Object

		obj, scale, err = scaleAnnotater.GetObjectWithScale(info.Namespace, scalableRef.CrossGroupObjectReference)
		if err != nil {
			if errors.IsNotFound(err) {
				utilruntime.HandleError(fmt.Errorf("%s %q does not exist, removing from list of scalables while unidling service %s/%s: %v", scalableRef.Kind, scalableRef.Name, info.Namespace, info.Name, err))
				delete(targetScalablesSet, scalableRef)
			} else {
				utilruntime.HandleError(fmt.Errorf("Unable to get scale for %s %q while unidling service %s/%s, will try again later: %v", scalableRef.Kind, scalableRef.Name, info.Namespace, info.Name, err))
			}
			continue
		}

		if scale.Spec.Replicas > 0 {
			klog.V(4).Infof("%s %q is not idle, skipping while unidling service %s/%s", scalableRef.Kind, scalableRef.Name, info.Namespace, info.Name)
			continue
		}

		scale.Spec.Replicas = scalableRef.Replicas

		updater := unidlingclient.NewScaleUpdater(codecs.LegacyCodec(scheme.PrioritizedVersionsAllGroups()...), info.Namespace, c.dcNamespacer, c.rcNamespacer)
		if err = scaleAnnotater.UpdateObjectScale(updater, info.Namespace, scalableRef.CrossGroupObjectReference, obj, scale); err != nil {
			if errors.IsNotFound(err) {
				utilruntime.HandleError(fmt.Errorf("%s %q does not exist, removing from list of scalables while unidling service %s/%s: %v", scalableRef.Kind, scalableRef.Name, info.Namespace, info.Name, err))
				delete(targetScalablesSet, scalableRef)
			} else {
				utilruntime.HandleError(fmt.Errorf("Unable to scale up %s %q while unidling service %s/%s: %v", scalableRef.Kind, scalableRef.Name, info.Namespace, info.Name, err))
			}
			continue
		} else {
			klog.V(4).Infof("Scaled up %s %q while unidling service %s/%s", scalableRef.Kind, scalableRef.Name, info.Namespace, info.Name)
		}

		delete(targetScalablesSet, scalableRef)
	}

	newAnnotationList := make([]unidlingapi.RecordedScaleReference, 0, len(targetScalablesSet))
	for k := range targetScalablesSet {
		newAnnotationList = append(newAnnotationList, k)
	}

	if len(newAnnotationList) == 0 {
		delete(targetService.Annotations, unidlingapi.UnidleTargetAnnotation)
		delete(targetService.Annotations, unidlingapi.IdledAtAnnotation)
	} else {
		var newAnnotationBytes []byte
		newAnnotationBytes, err = json.Marshal(newAnnotationList)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("unable to update/remove idle annotations from %s/%s: unable to marshal list of remaining scalables, removing list entirely: %v", info.Namespace, info.Name, err))

			delete(targetService.Annotations, unidlingapi.UnidleTargetAnnotation)
			delete(targetService.Annotations, unidlingapi.IdledAtAnnotation)
		} else {
			targetService.Annotations[unidlingapi.UnidleTargetAnnotation] = string(newAnnotationBytes)
		}
	}

	if _, err = c.servicesNamespacer.Services(info.Namespace).Update(context.TODO(), targetService, metav1.UpdateOptions{}); err != nil {
		return true, fmt.Errorf("unable to update/remove idle annotations from %s/%s: %v", info.Namespace, info.Name, err)
	}

	// oc idle still annotates endpoints for backwards
	// compatibilty. We need to remove any idled annotations on
	// the endpoints.

	// fetch the endpoints in question
	targetEndpoints, err := c.endpointsNamespacer.Endpoints(info.Namespace).Get(context.TODO(), info.Name, metav1.GetOptions{})
	if err != nil {
		return true, fmt.Errorf("unable to retrieve endpoints: %v", err)
	}

	// make sure we actually were idled...
	idledTimeRaw, wasIdled = targetEndpoints.Annotations[unidlingapi.IdledAtAnnotation]
	if !wasIdled {
		klog.V(5).Infof("UnidlingController received a NeedPods event for a service that was not idled, ignoring")
		return false, nil
	}

	// ...and make sure this request was to wake up from the most recent idling, and not a previous one
	idledTime, err = time.Parse(time.RFC3339, idledTimeRaw)
	if err != nil {
		// retrying here won't help, we're just stuck as idle since we can't get parse the idled time
		return false, fmt.Errorf("unable to check idled-at time: %v", err)
	}
	if lastFired.Before(idledTime) {
		klog.V(5).Infof("UnidlingController received an out-of-date NeedPods event, ignoring")
		return false, nil
	}

	_, unidledAnnotation := targetEndpoints.Annotations[unidlingapi.UnidleTargetAnnotation]
	_, idledAtAnnotation := targetEndpoints.Annotations[unidlingapi.IdledAtAnnotation]

	if unidledAnnotation || idledAtAnnotation {
		delete(targetEndpoints.Annotations, unidlingapi.UnidleTargetAnnotation)
		delete(targetEndpoints.Annotations, unidlingapi.IdledAtAnnotation)

		if _, err = c.endpointsNamespacer.Endpoints(info.Namespace).Update(context.TODO(), targetEndpoints, metav1.UpdateOptions{}); err != nil {
			return true, fmt.Errorf("unable to update/remove idle annotations from %s/%s: %v", info.Namespace, info.Name, err)
		}
	}

	return false, nil
}

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

func init() {
	utilruntime.Must(api.Install(scheme))
	utilruntime.Must(api.InstallKube(scheme))
}
