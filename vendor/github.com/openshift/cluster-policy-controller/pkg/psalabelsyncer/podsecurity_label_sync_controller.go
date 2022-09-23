package psalabelsyncer

import (
	"context"
	"fmt"
	"strings"

	"github.com/openshift/cluster-policy-controller/pkg/psalabelsyncer/nsexemptions"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/sets"
	corev1informers "k8s.io/client-go/informers/core/v1"
	rbacv1informers "k8s.io/client-go/informers/rbac/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	psapi "k8s.io/pod-security-admission/api"

	securityv1 "github.com/openshift/api/security/v1"
	securityv1informers "github.com/openshift/client-go/security/informers/externalversions/security/v1"
	securityv1listers "github.com/openshift/client-go/security/listers/security/v1"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
)

const (
	controllerName        = "pod-security-admission-label-synchronization-controller"
	labelSyncControlLabel = "security.openshift.io/scc.podSecurityLabelSync"
	currentPSaVersion     = "v1.24"
)

// PodSecurityAdmissionLabelSynchronizationController watches over namespaces labelled with
// "security.openshift.io/scc.podSecurityLabelSync: true" and configures the PodSecurity
// admission namespace label to match the user account privileges in terms of being able
// to use SCCs
type PodSecurityAdmissionLabelSynchronizationController struct {
	namespaceClient corev1client.NamespaceInterface

	namespaceLister      corev1listers.NamespaceLister
	serviceAccountLister corev1listers.ServiceAccountLister
	sccLister            securityv1listers.SecurityContextConstraintsLister

	nsLabelSelector labels.Selector

	workQueue     workqueue.RateLimitingInterface
	saToSCCsCache SAToSCCCache
}

func NewPodSecurityAdmissionLabelSynchronizationController(
	namespaceClient corev1client.NamespaceInterface,
	namespaceInformer corev1informers.NamespaceInformer,
	rbacInformers rbacv1informers.Interface,
	serviceAccountInformer corev1informers.ServiceAccountInformer,
	sccInformer securityv1informers.SecurityContextConstraintsInformer,
	eventRecorder events.Recorder,
) (factory.Controller, error) {

	// add the indexers that are used in the SAToSCC cache
	if err := rbacInformers.RoleBindings().Informer().AddIndexers(
		cache.Indexers{BySAIndexName: BySAIndexKeys},
	); err != nil {
		return nil, err
	}

	if err := rbacInformers.ClusterRoleBindings().Informer().AddIndexers(
		cache.Indexers{BySAIndexName: BySAIndexKeys},
	); err != nil {
		return nil, err
	}

	controlledNamespacesLabelSelector, err := controlledNamespacesLabelSelector()
	if err != nil {
		return nil, err
	}

	syncCtx := factory.NewSyncContext(controllerName, eventRecorder.WithComponentSuffix(controllerName))
	c := &PodSecurityAdmissionLabelSynchronizationController{
		namespaceClient: namespaceClient,

		namespaceLister:      namespaceInformer.Lister(),
		serviceAccountLister: serviceAccountInformer.Lister(),
		sccLister:            sccInformer.Lister(),

		nsLabelSelector: controlledNamespacesLabelSelector,

		workQueue: syncCtx.Queue(),
	}

	saToSCCCache := NewSAToSCCCache(rbacInformers, sccInformer).WithExternalQueueEnqueue(c.saToSCCCAcheEnqueueFunc)

	c.saToSCCsCache = saToSCCCache

	return factory.New().
		WithSync(c.sync).
		WithSyncContext(syncCtx).
		WithFilteredEventsInformersQueueKeysFunc(
			c.queueKeysRuntimeForObj,
			c.saToSCCsCache.IsRoleBindingRelevant,
			rbacInformers.RoleBindings().Informer(),
			rbacInformers.ClusterRoleBindings().Informer(),
		).
		WithFilteredEventsInformersQueueKeysFunc(
			nameToKey,
			func(obj interface{}) bool {
				ns, ok := obj.(*corev1.Namespace)
				if !ok {
					// we don't care if the NS is being deleted so we're not checking
					// for a tombstone
					return false
				}

				return isNSControlled(ns)
			},
			namespaceInformer.Informer(),
		).
		WithInformersQueueKeysFunc(
			c.queueKeysRuntimeForObj,
			serviceAccountInformer.Informer(),
		).
		ToController(
			controllerName,
			eventRecorder.WithComponentSuffix(controllerName),
		), nil
}

func (c *PodSecurityAdmissionLabelSynchronizationController) sync(ctx context.Context, controllerContext factory.SyncContext) error {
	const errFmt = "failed to synchronize namespace %q: %w"

	qKey := controllerContext.QueueKey()
	ns, err := c.namespaceLister.Get(qKey)
	if err != nil {
		return fmt.Errorf(errFmt, qKey, err)
	}

	if !isNSControlled(ns) {
		return nil
	}

	if ns.Status.Phase == corev1.NamespaceTerminating {
		klog.Infof("skipping synchronizing namespace %q because it is terminating", ns.Name)
		return nil
	}

	if err := c.syncNamespace(ctx, controllerContext, ns); err != nil {
		return fmt.Errorf(errFmt, qKey, err)
	}

	return nil
}

func (c *PodSecurityAdmissionLabelSynchronizationController) syncNamespace(ctx context.Context, controllerContext factory.SyncContext, ns *corev1.Namespace) error {
	// We cannot safely determine the SCC level for an NS until it gets the UID annotation.
	// No need to care about re-queueing the key, we should get the NS once it is updated
	// with the annotation.
	if len(ns.Annotations[securityv1.UIDRangeAnnotation]) == 0 {
		return nil
	}

	serviceAccounts, err := c.serviceAccountLister.ServiceAccounts(ns.Name).List(labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to list service accounts for %s: %w", ns.Name, err)
	}

	if len(serviceAccounts) == 0 {
		klog.Infof("no service accounts were found in the %q NS", ns.Name)
		return nil
	}

	nsSCCs := sets.NewString()
	for _, sa := range serviceAccounts {
		allowedSCCs, err := c.saToSCCsCache.SCCsFor(sa)
		if err != nil {
			return fmt.Errorf("failed to determine SCCs for ServiceAccount '%s/%s': %w", sa.Namespace, sa.Name, err)
		}
		nsSCCs.Insert(allowedSCCs.UnsortedList()...)
	}

	var currentNSLevel uint8
	for _, sccName := range nsSCCs.UnsortedList() {
		scc, err := c.sccLister.Get(sccName)
		if err != nil {
			return fmt.Errorf("failed to retrieve an SCC: %w", err)
		}
		sccPSaLevel, err := convertSCCToPSALevel(ns, scc)
		if err != nil {
			return fmt.Errorf("couldn't convert SCC %q to PSa level: %w", scc.Name, err)
		}

		if sccPSaLevel > currentNSLevel {
			currentNSLevel = sccPSaLevel
		}
		// can't go more privileged
		if currentNSLevel == privileged {
			break
		}
	}

	psaLevel := internalRestrictivnessToPSaLevel(currentNSLevel)
	if len(psaLevel) == 0 {
		return fmt.Errorf("unknown PSa level for namespace %q", ns.Name)
	}

	nsCopy := ns.DeepCopy()

	var changed bool
	if nsCopy.Labels[psapi.EnforceLevelLabel] != string(psaLevel) || nsCopy.Labels[psapi.EnforceVersionLabel] != currentPSaVersion {
		changed = true
		if nsCopy.Labels == nil {
			nsCopy.Labels = map[string]string{}
		}

		nsCopy.Labels[psapi.EnforceLevelLabel] = string(psaLevel)
		nsCopy.Labels[psapi.EnforceVersionLabel] = currentPSaVersion
	}

	// cleanup audit and warn labels from version 4.11
	// TODO: This can be removed in 4.13 and allow users set these as they wish
	for typeLabel, versionLabel := range map[string]string{
		psapi.WarnLevelLabel:  psapi.WarnVersionLabel,
		psapi.AuditLevelLabel: psapi.AuditVersionLabel,
	} {
		if _, ok := nsCopy.Labels[typeLabel]; ok {
			delete(nsCopy.Labels, typeLabel)
			changed = true
		}
		if _, ok := nsCopy.Labels[versionLabel]; ok {
			delete(nsCopy.Labels, versionLabel)
			changed = true
		}
	}

	if changed {
		_, err := c.namespaceClient.Update(ctx, nsCopy, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update the namespace: %w", err)
		}
	}

	return nil
}

// nameToKey turns a meta object into a key by using the object's name.
func nameToKey(obj runtime.Object) []string {
	metaObj, ok := obj.(metav1.ObjectMetaAccessor)
	if !ok {
		klog.Errorf("the object is not a metav1.ObjectMetaAccessor: %T", obj)
		return []string{}
	}
	return []string{metaObj.GetObjectMeta().GetName()}
}

// queueKeysRuntimeForObj is an adapter on top of queueKeysForObj to be used in
// factory.Controller queueing functions
func (c *PodSecurityAdmissionLabelSynchronizationController) queueKeysRuntimeForObj(obj runtime.Object) []string {
	return c.queueKeysForObj(obj)
}

// queueKeysForObj returns slice with:
// - a namespace name for a namespaced object
// - all cluster namespaces names for cluster-wide objects
func (c *PodSecurityAdmissionLabelSynchronizationController) queueKeysForObj(obj interface{}) []string {
	metaObj, ok := obj.(metav1.ObjectMetaAccessor)
	if !ok {
		klog.Error("unable to access metadata of the handled object: %v", obj)
		return nil
	}

	if ns := metaObj.GetObjectMeta().GetNamespace(); len(ns) > 0 { // it is a namespaced object
		controlled, err := c.isNSControlled(ns)
		if err != nil {
			klog.Errorf("failed to determine whether namespace %q should be enqueued: %v", ns, err)
			return nil
		}

		if !controlled {
			return nil
		}

		return []string{ns}
	}

	return c.allWatchedNamespacesAsQueueKeys()
}

// allWatchedNamespacesAsQueueKeys returns all namespace names slice irregardles of the retrieved object.
func (c *PodSecurityAdmissionLabelSynchronizationController) allWatchedNamespacesAsQueueKeys() []string {
	namespaces, err := c.namespaceLister.List(c.nsLabelSelector)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("failed to list namespaces while handling an informer event: %v", err)
	}

	qKeys := make([]string, 0, len(namespaces))
	for _, n := range namespaces {
		if controlled := isNSControlled(n); controlled {
			qKeys = append(qKeys, n.Name)
		}
	}

	return qKeys
}

// saToSCCCAcheEnqueueFunc is a function that allows the SAToSCCCache enqueue keys
// in the queue of the `PodSecurityAdmissionLabelSynchronizationController` struct
func (c *PodSecurityAdmissionLabelSynchronizationController) saToSCCCAcheEnqueueFunc(obj interface{}) {
	for _, qkey := range c.queueKeysForObj(obj) {
		c.workQueue.Add(qkey)
	}
}

func (c *PodSecurityAdmissionLabelSynchronizationController) isNSControlled(nsName string) (bool, error) {
	ns, err := c.namespaceLister.Get(nsName)
	if err != nil {
		return false, err
	}

	return isNSControlled(ns), nil

}

func isNSControlled(ns *corev1.Namespace) bool {
	nsName := ns.Name

	// never sync exempted namespaces
	if nsexemptions.IsNamespacePSALabelSyncExemptedInVendoredOCPVersion(nsName) {
		return false
	}

	// while "openshift-" namespaces should be considered controlled, there are some
	// edge cases where users can also create them. Consider these a special case
	// and delegate the decision to sync on the user who should know what they are
	// doing when creating a NS that appears to be system-controlled.
	if strings.HasPrefix(nsName, "openshift-") {
		return ns.Labels[labelSyncControlLabel] == "true"
	}

	return ns.Labels[labelSyncControlLabel] != "false"
}

// controlledNamespacesLabelSelector returns label selector to be used with the
// PodSecurityAdmissionLabelSynchronizationController.
func controlledNamespacesLabelSelector() (labels.Selector, error) {
	labelRequirement, err := labels.NewRequirement(labelSyncControlLabel, selection.NotEquals, []string{"false"})
	if err != nil {
		return nil, fmt.Errorf("failed to create a label requirement to list only opted-in namespaces: %w", err)
	}

	return labels.NewSelector().Add(*labelRequirement), nil
}
