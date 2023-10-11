/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package podgc

import (
	"context"
	"sort"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller/podgc/metrics"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/kubelet/eviction"
	nodeutil "k8s.io/kubernetes/pkg/util/node"
	"k8s.io/kubernetes/pkg/util/taints"
)

const (
	// gcCheckPeriod defines frequency of running main controller loop
	gcCheckPeriod = 20 * time.Second
	// quarantineTime defines how long Orphaned GC waits for nodes to show up
	// in an informer before issuing a GET call to check if they are truly gone
	quarantineTime = 40 * time.Second

	// field manager used to add pod failure condition and change the pod phase
	fieldManager = "PodGC"
)

type PodGCController struct {
	kubeClient clientset.Interface

	podLister        corelisters.PodLister
	podListerSynced  cache.InformerSynced
	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced

	nodeQueue workqueue.DelayingInterface

	terminatedPodThreshold int
	gcCheckPeriod          time.Duration
	quarantineTime         time.Duration
}

func NewPodGC(ctx context.Context, kubeClient clientset.Interface, podInformer coreinformers.PodInformer,
	nodeInformer coreinformers.NodeInformer, terminatedPodThreshold int) *PodGCController {
	return NewPodGCInternal(ctx, kubeClient, podInformer, nodeInformer, terminatedPodThreshold, gcCheckPeriod, quarantineTime)
}

// This function is only intended for integration tests
func NewPodGCInternal(ctx context.Context, kubeClient clientset.Interface, podInformer coreinformers.PodInformer,
	nodeInformer coreinformers.NodeInformer, terminatedPodThreshold int, gcCheckPeriod, quarantineTime time.Duration) *PodGCController {
	gcc := &PodGCController{
		kubeClient:             kubeClient,
		terminatedPodThreshold: terminatedPodThreshold,
		podLister:              podInformer.Lister(),
		podListerSynced:        podInformer.Informer().HasSynced,
		nodeLister:             nodeInformer.Lister(),
		nodeListerSynced:       nodeInformer.Informer().HasSynced,
		nodeQueue:              workqueue.NewNamedDelayingQueue("orphaned_pods_nodes"),
		gcCheckPeriod:          gcCheckPeriod,
		quarantineTime:         quarantineTime,
	}

	// Register prometheus metrics
	metrics.RegisterMetrics()
	return gcc
}

func (gcc *PodGCController) Run(ctx context.Context) {
	logger := klog.FromContext(ctx)

	defer utilruntime.HandleCrash()

	logger.Info("Starting GC controller")
	defer gcc.nodeQueue.ShutDown()
	defer logger.Info("Shutting down GC controller")

	if !cache.WaitForNamedCacheSync("GC", ctx.Done(), gcc.podListerSynced, gcc.nodeListerSynced) {
		return
	}

	go wait.UntilWithContext(ctx, gcc.gc, gcc.gcCheckPeriod)

	<-ctx.Done()
}

func (gcc *PodGCController) gc(ctx context.Context) {
	pods, err := gcc.podLister.List(labels.Everything())
	if err != nil {
		klog.FromContext(ctx).Error(err, "Error while listing all pods")
		return
	}
	nodes, err := gcc.nodeLister.List(labels.Everything())
	if err != nil {
		klog.FromContext(ctx).Error(err, "Error while listing all nodes")
		return
	}
	if gcc.terminatedPodThreshold > 0 {
		gcc.gcTerminated(ctx, pods)
	}
	gcc.gcTerminating(ctx, pods)
	gcc.gcOrphaned(ctx, pods, nodes)
	gcc.gcUnscheduledTerminating(ctx, pods)
}

func isPodTerminated(pod *v1.Pod) bool {
	if phase := pod.Status.Phase; phase != v1.PodPending && phase != v1.PodRunning && phase != v1.PodUnknown {
		return true
	}
	return false
}

// isPodTerminating returns true if the pod is terminating.
func isPodTerminating(pod *v1.Pod) bool {
	return pod.ObjectMeta.DeletionTimestamp != nil
}

func (gcc *PodGCController) gcTerminating(ctx context.Context, pods []*v1.Pod) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("GC'ing terminating pods that are on out-of-service nodes")
	terminatingPods := []*v1.Pod{}
	for _, pod := range pods {
		if isPodTerminating(pod) {
			node, err := gcc.nodeLister.Get(pod.Spec.NodeName)
			if err != nil {
				logger.Error(err, "Failed to get node", "node", klog.KRef("", pod.Spec.NodeName))
				continue
			}
			// Add this pod to terminatingPods list only if the following conditions are met:
			// 1. Node is not ready.
			// 2. Node has `node.kubernetes.io/out-of-service` taint.
			if !nodeutil.IsNodeReady(node) && taints.TaintKeyExists(node.Spec.Taints, v1.TaintNodeOutOfService) {
				logger.V(4).Info("Garbage collecting pod that is terminating", "pod", klog.KObj(pod), "phase", pod.Status.Phase)
				terminatingPods = append(terminatingPods, pod)
			}
		}
	}

	deleteCount := len(terminatingPods)
	if deleteCount == 0 {
		return
	}

	logger.V(4).Info("Garbage collecting pods that are terminating on node tainted with node.kubernetes.io/out-of-service", "numPods", deleteCount)
	// sort only when necessary
	sort.Sort(byEvictionAndCreationTimestamp(terminatingPods))
	var wait sync.WaitGroup
	for i := 0; i < deleteCount; i++ {
		wait.Add(1)
		go func(pod *v1.Pod) {
			defer wait.Done()
			metrics.DeletingPodsTotal.WithLabelValues(pod.Namespace, metrics.PodGCReasonTerminatingOutOfService).Inc()
			if err := gcc.markFailedAndDeletePod(ctx, pod); err != nil {
				// ignore not founds
				utilruntime.HandleError(err)
				metrics.DeletingPodsErrorTotal.WithLabelValues(pod.Namespace, metrics.PodGCReasonTerminatingOutOfService).Inc()
			}
		}(terminatingPods[i])
	}
	wait.Wait()
}

func (gcc *PodGCController) gcTerminated(ctx context.Context, pods []*v1.Pod) {
	terminatedPods := []*v1.Pod{}
	for _, pod := range pods {
		if isPodTerminated(pod) {
			terminatedPods = append(terminatedPods, pod)
		}
	}

	terminatedPodCount := len(terminatedPods)
	deleteCount := terminatedPodCount - gcc.terminatedPodThreshold

	if deleteCount <= 0 {
		return
	}

	logger := klog.FromContext(ctx)
	logger.Info("Garbage collecting pods", "numPods", deleteCount)
	// sort only when necessary
	sort.Sort(byEvictionAndCreationTimestamp(terminatedPods))
	var wait sync.WaitGroup
	for i := 0; i < deleteCount; i++ {
		wait.Add(1)
		go func(pod *v1.Pod) {
			defer wait.Done()
			if err := gcc.markFailedAndDeletePod(ctx, pod); err != nil {
				// ignore not founds
				defer utilruntime.HandleError(err)
				metrics.DeletingPodsErrorTotal.WithLabelValues(pod.Namespace, metrics.PodGCReasonTerminated).Inc()
			}
			metrics.DeletingPodsTotal.WithLabelValues(pod.Namespace, metrics.PodGCReasonTerminated).Inc()
		}(terminatedPods[i])
	}
	wait.Wait()
}

// gcOrphaned deletes pods that are bound to nodes that don't exist.
func (gcc *PodGCController) gcOrphaned(ctx context.Context, pods []*v1.Pod, nodes []*v1.Node) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("GC'ing orphaned")
	existingNodeNames := sets.NewString()
	for _, node := range nodes {
		existingNodeNames.Insert(node.Name)
	}
	// Add newly found unknown nodes to quarantine
	for _, pod := range pods {
		if pod.Spec.NodeName != "" && !existingNodeNames.Has(pod.Spec.NodeName) {
			gcc.nodeQueue.AddAfter(pod.Spec.NodeName, gcc.quarantineTime)
		}
	}
	// Check if nodes are still missing after quarantine period
	deletedNodesNames, quit := gcc.discoverDeletedNodes(ctx, existingNodeNames)
	if quit {
		return
	}
	// Delete orphaned pods
	for _, pod := range pods {
		if !deletedNodesNames.Has(pod.Spec.NodeName) {
			continue
		}
		logger.V(2).Info("Found orphaned Pod assigned to the Node, deleting", "pod", klog.KObj(pod), "node", klog.KRef("", pod.Spec.NodeName))
		condition := corev1apply.PodCondition().
			WithType(v1.DisruptionTarget).
			WithStatus(v1.ConditionTrue).
			WithReason("DeletionByPodGC").
			WithMessage("PodGC: node no longer exists").
			WithLastTransitionTime(metav1.Now())
		if err := gcc.markFailedAndDeletePodWithCondition(ctx, pod, condition); err != nil {
			utilruntime.HandleError(err)
			metrics.DeletingPodsErrorTotal.WithLabelValues(pod.Namespace, metrics.PodGCReasonOrphaned).Inc()
		} else {
			logger.Info("Forced deletion of orphaned Pod succeeded", "pod", klog.KObj(pod))
		}
		metrics.DeletingPodsTotal.WithLabelValues(pod.Namespace, metrics.PodGCReasonOrphaned).Inc()
	}
}

func (gcc *PodGCController) discoverDeletedNodes(ctx context.Context, existingNodeNames sets.String) (sets.String, bool) {
	deletedNodesNames := sets.NewString()
	for gcc.nodeQueue.Len() > 0 {
		item, quit := gcc.nodeQueue.Get()
		if quit {
			return nil, true
		}
		nodeName := item.(string)
		if !existingNodeNames.Has(nodeName) {
			exists, err := gcc.checkIfNodeExists(ctx, nodeName)
			switch {
			case err != nil:
				klog.FromContext(ctx).Error(err, "Error while getting node", "node", klog.KRef("", nodeName))
				// Node will be added back to the queue in the subsequent loop if still needed
			case !exists:
				deletedNodesNames.Insert(nodeName)
			}
		}
		gcc.nodeQueue.Done(item)
	}
	return deletedNodesNames, false
}

func (gcc *PodGCController) checkIfNodeExists(ctx context.Context, name string) (bool, error) {
	_, fetchErr := gcc.kubeClient.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if errors.IsNotFound(fetchErr) {
		return false, nil
	}
	return fetchErr == nil, fetchErr
}

// gcUnscheduledTerminating deletes pods that are terminating and haven't been scheduled to a particular node.
func (gcc *PodGCController) gcUnscheduledTerminating(ctx context.Context, pods []*v1.Pod) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("GC'ing unscheduled pods which are terminating")

	for _, pod := range pods {
		if pod.DeletionTimestamp == nil || len(pod.Spec.NodeName) > 0 {
			continue
		}

		logger.V(2).Info("Found unscheduled terminating Pod not assigned to any Node, deleting", "pod", klog.KObj(pod))
		if err := gcc.markFailedAndDeletePod(ctx, pod); err != nil {
			utilruntime.HandleError(err)
			metrics.DeletingPodsErrorTotal.WithLabelValues(pod.Namespace, metrics.PodGCReasonTerminatingUnscheduled).Inc()
		} else {
			logger.Info("Forced deletion of unscheduled terminating Pod succeeded", "pod", klog.KObj(pod))
		}
		metrics.DeletingPodsTotal.WithLabelValues(pod.Namespace, metrics.PodGCReasonTerminatingUnscheduled).Inc()
	}
}

// byEvictionAndCreationTimestamp sorts a list by Evicted status and then creation timestamp,
// using their names as a tie breaker.
// Evicted pods will be deleted first to avoid impact on terminated pods created by controllers.
type byEvictionAndCreationTimestamp []*v1.Pod

func (o byEvictionAndCreationTimestamp) Len() int      { return len(o) }
func (o byEvictionAndCreationTimestamp) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func (o byEvictionAndCreationTimestamp) Less(i, j int) bool {
	iEvicted, jEvicted := eviction.PodIsEvicted(o[i].Status), eviction.PodIsEvicted(o[j].Status)
	// Evicted pod is smaller
	if iEvicted != jEvicted {
		return iEvicted
	}
	if o[i].CreationTimestamp.Equal(&o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].CreationTimestamp.Before(&o[j].CreationTimestamp)
}

func (gcc *PodGCController) markFailedAndDeletePod(ctx context.Context, pod *v1.Pod) error {
	return gcc.markFailedAndDeletePodWithCondition(ctx, pod, nil)
}

func (gcc *PodGCController) markFailedAndDeletePodWithCondition(ctx context.Context, pod *v1.Pod, condition *corev1apply.PodConditionApplyConfiguration) error {
	logger := klog.FromContext(ctx)
	logger.Info("PodGC is force deleting Pod", "pod", klog.KObj(pod))
	// Patch the pod to make sure it is transitioned to the Failed phase before deletion.
	// This is needed for the JobPodReplacementPolicy feature to make sure Job replacement pods are created.
	// See https://github.com/kubernetes/enhancements/tree/master/keps/sig-apps/3939-allow-replacement-when-fully-terminated#risks-and-mitigations
	// for more details.
	if utilfeature.DefaultFeatureGate.Enabled(features.PodDisruptionConditions) || utilfeature.DefaultFeatureGate.Enabled(features.JobPodReplacementPolicy) {

		// Mark the pod as failed - this is especially important in case the pod
		// is orphaned, in which case the pod would remain in the Running phase
		// forever as there is no kubelet running to change the phase.
		if pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
			podApply := corev1apply.Pod(pod.Name, pod.Namespace).WithStatus(corev1apply.PodStatus())
			// we don't need to extract the pod apply configuration and can send
			// only phase and the DisruptionTarget condition as PodGC would not
			// own other fields. If the DisruptionTarget condition is owned by
			// PodGC it means that it is in the Failed phase, so sending the
			// condition will not be re-attempted.
			podApply.Status.WithPhase(v1.PodFailed)
			if condition != nil && utilfeature.DefaultFeatureGate.Enabled(features.PodDisruptionConditions) {
				podApply.Status.WithConditions(condition)
			}
			if _, err := gcc.kubeClient.CoreV1().Pods(pod.Namespace).ApplyStatus(ctx, podApply, metav1.ApplyOptions{FieldManager: fieldManager, Force: true}); err != nil {
				return err
			}
		}
	}
	return gcc.kubeClient.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, *metav1.NewDeleteOptions(0))
}
