package psalabelsyncer

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller"
	psapi "k8s.io/pod-security-admission/api"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
)

const privilegedControllerName = "privileged-namespaces-psa-label-syncer"

type privilegedNamespacesPSALabelSyncer struct {
	nsClient      corev1client.NamespaceInterface
	nsLister      corev1listers.NamespaceLister
	nsCacheSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
}

func NewPrivilegedNamespacesPSALabelSyncer(
	ctx context.Context,
	namespaceClient corev1client.NamespaceInterface,
	namespaceInformer corev1informers.NamespaceInformer,
	eventRecorder events.Recorder,
) *privilegedNamespacesPSALabelSyncer {
	c := &privilegedNamespacesPSALabelSyncer{
		nsClient: namespaceClient,
		nsLister: namespaceInformer.Lister(),

		nsCacheSynced: namespaceInformer.Informer().HasSynced,

		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), privilegedControllerName),
	}

	logger := klog.FromContext(ctx)

	namespaceInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: factory.NamesFilter("default", "kube-system", "kube-public"),
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					c.enqueueNS(logger, obj)
				},
				UpdateFunc: func(_, newObj interface{}) {
					c.enqueueNS(logger, newObj)
				},
				DeleteFunc: nil,
			},
		},
	)

	return c
}

func (c *privilegedNamespacesPSALabelSyncer) Run(ctx context.Context, _ int) {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	logger := klog.FromContext(ctx)
	logger.Info("Starting", "controller", privilegedControllerName)
	defer logger.Info("Shutting down", "controller", privilegedControllerName)

	if !cache.WaitForNamedCacheSync(privilegedControllerName, ctx.Done(), c.nsCacheSynced) {
		return
	}

	go wait.UntilWithContext(ctx, c.worker, time.Second)

	<-ctx.Done()
}

func (c *privilegedNamespacesPSALabelSyncer) worker(ctx context.Context) {
	workerFunc := func(ctx context.Context) bool {
		key, quit := c.workqueue.Get()
		if quit {
			return true
		}
		defer c.workqueue.Done(key)

		logger := klog.FromContext(ctx)
		logger = klog.LoggerWithValues(logger, "queueKey", key)
		ctx = klog.NewContext(ctx, logger)

		err := c.sync(ctx, key.(string))
		if err == nil {
			c.workqueue.Forget(key)
			return false
		}

		utilruntime.HandleError(err)
		c.workqueue.AddRateLimited(key)

		return false
	}

	for {
		if quit := workerFunc(ctx); quit {
			return
		}
	}
}

func (c *privilegedNamespacesPSALabelSyncer) sync(ctx context.Context, nsName string) error {
	ns, err := c.nsLister.Get(nsName)
	if err != nil {
		return fmt.Errorf("failed to retrieve ns %q: %w", nsName, err)
	}

	if ns.Labels[psapi.EnforceLevelLabel] == "privileged" &&
		ns.Labels[psapi.WarnLevelLabel] == "privileged" &&
		ns.Labels[psapi.AuditLevelLabel] == "privileged" {
		return nil
	}

	nsApplyConfig := corev1apply.Namespace(ns.Name).WithLabels(map[string]string{
		psapi.EnforceLevelLabel: "privileged",
		psapi.WarnLevelLabel:    "privileged",
		psapi.AuditLevelLabel:   "privileged",
	})

	_, err = c.nsClient.Apply(ctx, nsApplyConfig, v1.ApplyOptions{FieldManager: privilegedControllerName, Force: true})

	return err
}

func (c *privilegedNamespacesPSALabelSyncer) enqueueNS(logger klog.Logger, obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		logger.Error(err, "Couldn't get key", "object", obj)
		return
	}
	c.workqueue.Add(key)
}
