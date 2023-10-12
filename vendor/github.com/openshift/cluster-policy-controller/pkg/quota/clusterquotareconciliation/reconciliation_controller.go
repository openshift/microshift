package clusterquotareconciliation

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kutilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	utilquota "k8s.io/apiserver/pkg/quota/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/cache"
	"k8s.io/controller-manager/pkg/informerfactory"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/resourcequota"

	quotav1 "github.com/openshift/api/quota/v1"
	quotatypedclient "github.com/openshift/client-go/quota/clientset/versioned/typed/quota/v1"
	quotainformer "github.com/openshift/client-go/quota/informers/externalversions/quota/v1"
	quotalister "github.com/openshift/client-go/quota/listers/quota/v1"
	"github.com/openshift/library-go/pkg/quota/clusterquotamapping"
	quotautil "github.com/openshift/library-go/pkg/quota/quotautil"
)

type ClusterQuotaReconcilationControllerOptions struct {
	ClusterQuotaInformer quotainformer.ClusterResourceQuotaInformer
	ClusterQuotaMapper   clusterquotamapping.ClusterQuotaMapper
	ClusterQuotaClient   quotatypedclient.ClusterResourceQuotaInterface

	// Knows how to calculate usage
	Registry utilquota.Registry
	// Controls full recalculation of quota usage
	ResyncPeriod time.Duration
	// Discover list of supported resources on the server.
	DiscoveryFunc resourcequota.NamespacedResourcesFunc
	// A function that returns the list of resources to ignore
	IgnoredResourcesFunc func() map[schema.GroupResource]struct{}
	// InformersStarted knows if informers were started.
	InformersStarted <-chan struct{}
	// InformerFactory interfaces with informers.
	InformerFactory informerfactory.InformerFactory
	// Controls full resync of objects monitored for replenihsment.
	ReplenishmentResyncPeriod controller.ResyncPeriodFunc
	// Filters update events so we only enqueue the ones where we know quota will change
	UpdateFilter resourcequota.UpdateFilter
}

type ClusterQuotaReconcilationController struct {
	clusterQuotaLister quotalister.ClusterResourceQuotaLister
	clusterQuotaMapper clusterquotamapping.ClusterQuotaMapper
	clusterQuotaClient quotatypedclient.ClusterResourceQuotaInterface
	// A list of functions that return true when their caches have synced
	informerSyncedFuncs []cache.InformerSynced

	resyncPeriod time.Duration

	// queue tracks which clusterquotas to update along with a list of namespaces for that clusterquota
	queue BucketingWorkQueue

	// knows how to calculate usage
	registry utilquota.Registry
	// knows how to monitor all the resources tracked by quota and trigger replenishment
	quotaMonitor *resourcequota.QuotaMonitor
	// controls the workers that process quotas
	// this lock is acquired to control write access to the monitors and ensures that all
	// monitors are synced before the controller can process quotas.
	workerLock sync.RWMutex
}

type workItem struct {
	namespaceName      string
	forceRecalculation bool
}

func NewClusterQuotaReconcilationController(ctx context.Context, options ClusterQuotaReconcilationControllerOptions) (*ClusterQuotaReconcilationController, error) {
	c := &ClusterQuotaReconcilationController{
		clusterQuotaLister:  options.ClusterQuotaInformer.Lister(),
		clusterQuotaMapper:  options.ClusterQuotaMapper,
		clusterQuotaClient:  options.ClusterQuotaClient,
		informerSyncedFuncs: []cache.InformerSynced{options.ClusterQuotaInformer.Informer().HasSynced},

		resyncPeriod: options.ResyncPeriod,
		registry:     options.Registry,

		queue: NewBucketingWorkQueue("controller_clusterquotareconcilationcontroller"),
	}

	// we need to trigger every time
	options.ClusterQuotaInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addClusterQuota,
		UpdateFunc: c.updateClusterQuota,
	})

	qm := resourcequota.NewMonitor(
		options.InformersStarted,
		options.InformerFactory,
		options.IgnoredResourcesFunc(),
		options.ReplenishmentResyncPeriod,
		c.replenishQuota,
		c.registry,
		options.UpdateFilter,
	)

	c.quotaMonitor = qm

	// do initial quota monitor setup.  If we have a discovery failure here, it's ok. We'll discover more resources when a later sync happens.
	resources, err := resourcequota.GetQuotableResources(options.DiscoveryFunc)
	if discovery.IsGroupDiscoveryFailedError(err) {
		utilruntime.HandleError(fmt.Errorf("initial discovery check failure, continuing and counting on future sync update: %v", err))
	} else if err != nil {
		return nil, err
	}

	if err = qm.SyncMonitors(ctx, resources); err != nil {
		utilruntime.HandleError(fmt.Errorf("initial monitor sync has error: %v", err))
	}

	// only start quota once all informers synced
	c.informerSyncedFuncs = append(c.informerSyncedFuncs, func() bool { return qm.IsSynced(ctx) })

	return c, nil
}

// Run begins quota controller using the specified number of workers
func (c *ClusterQuotaReconcilationController) Run(workers int, ctx context.Context) {
	defer utilruntime.HandleCrash()

	klog.Infof("Starting the cluster quota reconciliation controller")

	// the controllers that replenish other resources to respond rapidly to state changes
	go c.quotaMonitor.Run(ctx)

	if !cache.WaitForCacheSync(ctx.Done(), c.informerSyncedFuncs...) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	klog.Infof("Caches are synced")
	// the workers that chug through the quota calculation backlog
	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, ctx.Done())
	}

	// the timer for how often we do a full recalculation across all quotas
	go wait.Until(func() { c.calculateAll() }, c.resyncPeriod, ctx.Done())

	<-ctx.Done()
	klog.Infof("Shutting down ClusterQuotaReconcilationController")
	c.queue.ShutDown()
}

// Sync periodically resyncs the controller when new resources are observed from discovery.
func (c *ClusterQuotaReconcilationController) Sync(discoveryFunc resourcequota.NamespacedResourcesFunc, period time.Duration, ctx context.Context) {
	// Something has changed, so track the new state and perform a sync.
	oldResources := make(map[schema.GroupVersionResource]struct{})
	wait.Until(func() {
		// Get the current resource list from discovery.
		newResources, err := resourcequota.GetQuotableResources(discoveryFunc)
		if err != nil {
			klog.V(2).Infof("error occurred GetQuotableResources err=%v", err)
			utilruntime.HandleError(err)

			if discovery.IsGroupDiscoveryFailedError(err) && len(newResources) > 0 {
				// In partial discovery cases, don't remove any existing informers, just add new ones
				for k, v := range oldResources {
					newResources[k] = v
				}
			} else {
				// short circuit in non-discovery error cases or if discovery returned zero resources
				return
			}
		}

		// Empty list of resources is a sign of incorrectness in the cluster.
		// Either the kube-apiserver incorrectly returns an empty list with no error.
		// Or, the client-code responsible for processing response from the discovery
		// is incorrectly interpretting the server error.
		//
		// If the list of resources is zero, all current monitors are cleared and removed.
		// Followed by invoking c.quotaMonitor.IsSynced which is always false when there's
		// no monitor running. In which case cache.WaitForCacheSync never returns
		// and loops indefinitely. Never resyncing the resources. Thus, none of the resources
		// mentioned on any ClusterResourceQuota's status are recomputed since there's no
		// resource monitor to provide current resource usage.
		if len(newResources) == 0 {
			klog.V(2).Infof("no resources discovered, skipping resource quota sync")
			return
		}

		// Decide whether discovery has reported a change.
		if reflect.DeepEqual(oldResources, newResources) {
			klog.V(4).Infof("no resource updates from discovery, skipping resource quota sync")
			return
		}

		klog.V(2).Infof("syncing resource quota controller with updated resources from discovery: %s", printDiff(oldResources, newResources))

		// Ensure workers are paused to avoid processing events before informers
		// have resynced.
		c.workerLock.Lock()
		defer c.workerLock.Unlock()

		// Perform the monitor resync and wait for controllers to report cache sync.
		if err := c.resyncMonitors(ctx, newResources); err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to sync resource monitors: %v", err))
			return
		}
		if c.quotaMonitor != nil && !cache.WaitForCacheSync(waitForStopOrTimeout(ctx.Done(), period), func() bool { return c.quotaMonitor.IsSynced(context.TODO()) }) {
			utilruntime.HandleError(fmt.Errorf("timed out waiting for quota monitor sync"))
		}

		oldResources = newResources
		klog.V(2).Infof("synced cluster resource quota controller")
	}, period, ctx.Done())
}

// printDiff returns a human-readable summary of what resources were added and removed
func printDiff(oldResources, newResources map[schema.GroupVersionResource]struct{}) string {
	removed := sets.NewString()
	for oldResource := range oldResources {
		if _, ok := newResources[oldResource]; !ok {
			removed.Insert(fmt.Sprintf("%+v", oldResource))
		}
	}
	added := sets.NewString()
	for newResource := range newResources {
		if _, ok := oldResources[newResource]; !ok {
			added.Insert(fmt.Sprintf("%+v", newResource))
		}
	}
	return fmt.Sprintf("added: %v, removed: %v", added.List(), removed.List())
}

// waitForStopOrTimeout returns a stop channel that closes when the provided stop channel closes or when the specified timeout is reached
func waitForStopOrTimeout(stopCh <-chan struct{}, timeout time.Duration) <-chan struct{} {
	stopChWithTimeout := make(chan struct{})
	go func() {
		defer close(stopChWithTimeout)
		select {
		case <-stopCh:
		case <-time.After(timeout):
		}
	}()
	return stopChWithTimeout
}

// resyncMonitors starts or stops quota monitors as needed to ensure that all
// (and only) those resources present in the map are monitored.
func (c *ClusterQuotaReconcilationController) resyncMonitors(ctx context.Context, resources map[schema.GroupVersionResource]struct{}) error {
	// SyncMonitors can only fail if there was no Informer for the given gvr
	err := c.quotaMonitor.SyncMonitors(ctx, resources)
	// this is no-op for already running monitors
	c.quotaMonitor.StartMonitors(ctx)
	return err
}

func (c *ClusterQuotaReconcilationController) calculate(quotaName string, namespaceNames ...string) {
	if len(namespaceNames) == 0 {
		klog.V(2).Infof("no namespace is passed for quota %s", quotaName)
		return
	}
	items := make([]interface{}, 0, len(namespaceNames))
	for _, name := range namespaceNames {
		items = append(items, workItem{namespaceName: name, forceRecalculation: false})
	}

	klog.V(2).Infof("calculating items for quota %s with namespaces %v", quotaName, items)
	c.queue.AddWithData(quotaName, items...)
}

func (c *ClusterQuotaReconcilationController) forceCalculation(quotaName string, namespaceNames ...string) {
	if len(namespaceNames) == 0 {
		return
	}
	items := make([]interface{}, 0, len(namespaceNames))
	for _, name := range namespaceNames {
		items = append(items, workItem{namespaceName: name, forceRecalculation: true})
	}

	klog.V(2).Infof("force calculating items for quota %s with namespaces %v", quotaName, items)
	c.queue.AddWithData(quotaName, items...)
}

func (c *ClusterQuotaReconcilationController) calculateAll() {
	quotas, err := c.clusterQuotaLister.List(labels.Everything())
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	for _, quota := range quotas {
		// If we have namespaces we map to, force calculating those namespaces
		namespaces, _ := c.clusterQuotaMapper.GetNamespacesFor(quota.Name)
		if len(namespaces) > 0 {
			klog.V(2).Infof("syncing quota %s with namespaces %v", quota.Name, namespaces)
			c.forceCalculation(quota.Name, namespaces...)
			continue
		}

		// If the quota status has namespaces when our mapper doesn't think it should,
		// add it directly to the queue without any work items
		if len(quota.Status.Namespaces) > 0 {
			c.queue.AddWithData(quota.Name)
			continue
		}
	}
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (c *ClusterQuotaReconcilationController) worker() {
	workFunc := func() bool {
		uncastKey, uncastData, quit := c.queue.GetWithData()
		if quit {
			return true
		}
		defer c.queue.Done(uncastKey)

		c.workerLock.RLock()
		defer c.workerLock.RUnlock()

		quotaName := uncastKey.(string)
		quota, err := c.clusterQuotaLister.Get(quotaName)
		if apierrors.IsNotFound(err) {
			klog.V(4).Infof("queued quota %s not found in quota lister", quotaName)
			c.queue.Forget(uncastKey)
			return false
		}
		if err != nil {
			utilruntime.HandleError(err)
			c.queue.AddWithDataRateLimited(uncastKey, uncastData...)
			return false
		}

		workItems := make([]workItem, 0, len(uncastData))
		for _, dataElement := range uncastData {
			workItems = append(workItems, dataElement.(workItem))
		}
		err, retryItems := c.syncQuotaForNamespaces(quota, workItems)
		if err == nil {
			c.queue.Forget(uncastKey)
			return false
		}
		utilruntime.HandleError(err)

		items := make([]interface{}, 0, len(retryItems))
		for _, item := range retryItems {
			items = append(items, item)
		}
		c.queue.AddWithDataRateLimited(uncastKey, items...)
		return false
	}

	for {
		if quit := workFunc(); quit {
			klog.Infof("resource quota controller worker shutting down")
			return
		}
	}
}

// syncResourceQuotaFromKey syncs a quota key
func (c *ClusterQuotaReconcilationController) syncQuotaForNamespaces(originalQuota *quotav1.ClusterResourceQuota, workItems []workItem) (error, []workItem /* to retry */) {
	quota := originalQuota.DeepCopy()

	// get the list of namespaces that match this cluster quota
	matchingNamespaceNamesList, quotaSelector := c.clusterQuotaMapper.GetNamespacesFor(quota.Name)
	if !equality.Semantic.DeepEqual(quotaSelector, quota.Spec.Selector) {
		return fmt.Errorf("mapping not up to date, have=%v need=%v", quotaSelector, quota.Spec.Selector), workItems
	}
	matchingNamespaceNames := sets.NewString(matchingNamespaceNamesList...)
	klog.V(2).Infof("syncing for quota %s with set of namespaces %v", quota.Name, matchingNamespaceNames)

	reconcilationErrors := []error{}
	retryItems := []workItem{}
	for _, item := range workItems {
		namespaceName := item.namespaceName
		namespaceTotals, namespaceLoaded := quotautil.GetResourceQuotasStatusByNamespace(quota.Status.Namespaces, namespaceName)
		if !matchingNamespaceNames.Has(namespaceName) {
			if namespaceLoaded {
				// remove this item from all totals
				quota.Status.Total.Used = utilquota.Subtract(quota.Status.Total.Used, namespaceTotals.Used)
				quotautil.RemoveResourceQuotasStatusByNamespace(&quota.Status.Namespaces, namespaceName)
			}
			continue
		}

		// if there's no work for us to do, do nothing
		if !item.forceRecalculation && namespaceLoaded && equality.Semantic.DeepEqual(namespaceTotals.Hard, quota.Spec.Quota.Hard) {
			continue
		}

		actualUsage, err := quotaUsageCalculationFunc(namespaceName, quota.Spec.Quota.Scopes, quota.Spec.Quota.Hard, c.registry, quota.Spec.Quota.ScopeSelector)
		if err != nil {
			// tally up errors, but calculate everything you can
			reconcilationErrors = append(reconcilationErrors, err)
			retryItems = append(retryItems, item)
			continue
		}
		recalculatedStatus := corev1.ResourceQuotaStatus{
			Used: actualUsage,
			Hard: quota.Spec.Quota.Hard,
		}

		// subtract old usage, add new usage
		quota.Status.Total.Used = utilquota.Subtract(quota.Status.Total.Used, namespaceTotals.Used)
		quota.Status.Total.Used = utilquota.Add(quota.Status.Total.Used, recalculatedStatus.Used)
		quotautil.InsertResourceQuotasStatus(&quota.Status.Namespaces, quotav1.ResourceQuotaStatusByNamespace{
			Namespace: namespaceName,
			Status:    recalculatedStatus,
		})
	}

	// Remove any namespaces from quota.status that no longer match.
	// Needed because we will never get workitems for namespaces that no longer exist if we missed the delete event (e.g. on startup)
	// range on a copy so that we don't mutate our original
	statusCopy := quota.Status.Namespaces.DeepCopy()
	for _, namespaceTotals := range statusCopy {
		namespaceName := namespaceTotals.Namespace
		if !matchingNamespaceNames.Has(namespaceName) {
			quota.Status.Total.Used = utilquota.Subtract(quota.Status.Total.Used, namespaceTotals.Status.Used)
			quotautil.RemoveResourceQuotasStatusByNamespace(&quota.Status.Namespaces, namespaceName)
		}
	}

	quota.Status.Total.Hard = quota.Spec.Quota.Hard

	// if there's no change, no update, return early.  NewAggregate returns nil on empty input
	if equality.Semantic.DeepEqual(quota, originalQuota) {
		return kutilerrors.NewAggregate(reconcilationErrors), retryItems
	}

	if _, err := c.clusterQuotaClient.UpdateStatus(context.TODO(), quota, metav1.UpdateOptions{}); err != nil {
		return kutilerrors.NewAggregate(append(reconcilationErrors, err)), workItems
	}

	return kutilerrors.NewAggregate(reconcilationErrors), retryItems
}

// replenishQuota is a replenishment function invoked by a controller to notify that a quota should be recalculated
func (c *ClusterQuotaReconcilationController) replenishQuota(ctx context.Context, groupResource schema.GroupResource, namespace string) {
	// check if the quota controller can evaluate this kind, if not, ignore it altogether...
	releventEvaluators := []utilquota.Evaluator{}
	evaluators := c.registry.List()
	for i := range evaluators {
		evaluator := evaluators[i]
		if evaluator.GroupResource() == groupResource {
			releventEvaluators = append(releventEvaluators, evaluator)
		}
	}
	if len(releventEvaluators) == 0 {
		return
	}

	quotaNames, _ := c.clusterQuotaMapper.GetClusterQuotasFor(namespace)
	if len(quotaNames) > 0 {
		klog.V(2).Infof("replenish quotas %v for namespace %s", quotaNames, namespace)
	}

	// only queue those quotas that are tracking a resource associated with this kind.
	for _, quotaName := range quotaNames {
		quota, err := c.clusterQuotaLister.Get(quotaName)
		if err != nil {
			// replenishment will be delayed, but we'll get back around to it later if it matters
			continue
		}

		resourceQuotaResources := utilquota.ResourceNames(quota.Status.Total.Hard)
		for _, evaluator := range releventEvaluators {
			matchedResources := evaluator.MatchingResources(resourceQuotaResources)
			if len(matchedResources) > 0 {
				// TODO: make this support targeted replenishment to a specific kind, right now it does a full recalc on that quota.
				c.forceCalculation(quotaName, namespace)
				break
			}
		}
	}
}

func (c *ClusterQuotaReconcilationController) addClusterQuota(cur interface{}) {
	c.enqueueClusterQuota(cur)
}
func (c *ClusterQuotaReconcilationController) updateClusterQuota(old, cur interface{}) {
	c.enqueueClusterQuota(cur)
}
func (c *ClusterQuotaReconcilationController) enqueueClusterQuota(obj interface{}) {
	quota, ok := obj.(*quotav1.ClusterResourceQuota)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("not a ClusterResourceQuota %v", obj))
		return
	}

	namespaces, _ := c.clusterQuotaMapper.GetNamespacesFor(quota.Name)
	c.calculate(quota.Name, namespaces...)
}

func (c *ClusterQuotaReconcilationController) AddMapping(quotaName, namespaceName string) {
	c.calculate(quotaName, namespaceName)

}
func (c *ClusterQuotaReconcilationController) RemoveMapping(quotaName, namespaceName string) {
	c.calculate(quotaName, namespaceName)
}

// quotaUsageCalculationFunc is a function to calculate quota usage.  It is only configurable for easy unit testing
// NEVER CHANGE THIS OUTSIDE A TEST
var quotaUsageCalculationFunc = utilquota.CalculateUsage
