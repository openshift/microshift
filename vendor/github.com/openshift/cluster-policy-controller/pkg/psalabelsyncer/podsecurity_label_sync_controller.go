package psalabelsyncer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openshift/cluster-policy-controller/pkg/psalabelsyncer/nsexemptions"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/sets"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
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

var (
	allPSaLabels = map[string]string{
		psapi.EnforceLevelLabel: psapi.EnforceVersionLabel,
		psapi.WarnLevelLabel:    psapi.WarnVersionLabel,
		psapi.AuditLevelLabel:   psapi.AuditVersionLabel,
	}
	loggingLabels = map[string]string{
		psapi.WarnLevelLabel:  psapi.WarnVersionLabel,
		psapi.AuditLevelLabel: psapi.AuditVersionLabel,
	}
)

// PodSecurityAdmissionLabelSynchronizationController watches over namespaces labelled with
// "security.openshift.io/scc.podSecurityLabelSync: true" and configures the PodSecurity
// admission namespace label to match the user account privileges in terms of being able
// to use SCCs
type PodSecurityAdmissionLabelSynchronizationController struct {
	syncedLabels map[string]string

	namespaceClient corev1client.NamespaceInterface

	namespaceLister      corev1listers.NamespaceLister
	serviceAccountLister corev1listers.ServiceAccountLister
	sccLister            securityv1listers.SecurityContextConstraintsLister

	nsLabelSelector labels.Selector

	workQueue     workqueue.RateLimitingInterface
	saToSCCsCache SAToSCCCache
}

func NewEnforcingPodSecurityAdmissionLabelSynchronizationController(
	namespaceClient corev1client.NamespaceInterface,
	namespaceInformer corev1informers.NamespaceInformer,
	rbacInformers rbacv1informers.Interface,
	serviceAccountInformer corev1informers.ServiceAccountInformer,
	sccInformer securityv1informers.SecurityContextConstraintsInformer,
	eventRecorder events.Recorder,
) (factory.Controller, error) {
	return newPodSecurityAdmissionLabelSynchronizationController(
		allPSaLabels,
		namespaceClient,
		namespaceInformer,
		rbacInformers,
		serviceAccountInformer,
		sccInformer,
		eventRecorder,
	)
}

func NewAdvisingPodSecurityAdmissionLabelSynchronizationController(
	namespaceClient corev1client.NamespaceInterface,
	namespaceInformer corev1informers.NamespaceInformer,
	rbacInformers rbacv1informers.Interface,
	serviceAccountInformer corev1informers.ServiceAccountInformer,
	sccInformer securityv1informers.SecurityContextConstraintsInformer,
	eventRecorder events.Recorder,
) (factory.Controller, error) {
	return newPodSecurityAdmissionLabelSynchronizationController(
		loggingLabels,
		namespaceClient,
		namespaceInformer,
		rbacInformers,
		serviceAccountInformer,
		sccInformer,
		eventRecorder,
	)
}

func newPodSecurityAdmissionLabelSynchronizationController(
	syncedLabels map[string]string,
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
		syncedLabels: syncedLabels,

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
		if apierrors.IsNotFound(err) {
			klog.V(4).Infof("namespace %q was not found, nothing to do", qKey)
			return nil
		}
		return fmt.Errorf(errFmt, qKey, err)
	}

	if !isNSControlled(ns) {
		return nil
	}

	if ns.Status.Phase == corev1.NamespaceTerminating {
		klog.Infof("skipping synchronizing namespace %q because it is terminating", ns.Name)
		return nil
	}

	ns, err = forceHistoricalLabelsOwnership(ctx, c.namespaceClient, ns)
	if err != nil {
		return fmt.Errorf("failed to force ownership from cluster-policy-controller to %s: %w", controllerName, err)
	}

	if err := c.syncNamespace(ctx, controllerContext, ns); err != nil {
		return fmt.Errorf(errFmt, qKey, err)
	}

	return nil
}

func forceHistoricalLabelsOwnership(ctx context.Context, nsClient corev1client.NamespaceInterface, ns *corev1.Namespace) (*corev1.Namespace, error) {
	cpcOwnedLabelKeys := sets.New[string]()
	for _, f := range ns.ManagedFields {
		if f.Manager != "cluster-policy-controller" {
			continue
		}

		newCPCLabels, err := managedLabels(f)
		if err != nil {
			return nil, err
		}

		cpcOwnedLabelKeys = cpcOwnedLabelKeys.Union(newCPCLabels)
	}

	if cpcOwnedLabelKeys.Len() == 0 {
		return ns, nil
	}

	cpcOwnedPSaLabels := map[string]string{}
	// filter out all the labels not owned by this controller
	for labelType, labelVersion := range allPSaLabels {
		if cpcOwnedLabelKeys.Has(labelType) {
			cpcOwnedPSaLabels[labelType] = ns.Labels[labelType]
		}
		if cpcOwnedLabelKeys.Has(labelVersion) {
			cpcOwnedPSaLabels[labelVersion] = ns.Labels[labelVersion]
		}
	}

	// none of the labels CPC is managing are interesting to us
	if len(cpcOwnedPSaLabels) == 0 {
		return ns, nil
	}

	// we need to extract all our managed fields not to delete them on the apply below
	ourOwned, err := corev1apply.ExtractNamespace(ns, controllerName)
	if err != nil {
		return nil, err
	}

	// add the PSa labels that were previously owned by CPC under this manager
	ourOwned.WithLabels(cpcOwnedPSaLabels)

	nsCopy := ns.DeepCopy()
	for labelKey := range cpcOwnedPSaLabels {
		delete(nsCopy.Labels, labelKey)
	}

	// previously, we were using Update to set the labels, Kube does not consider that as actually owning the fields, even
	// though it shows up in managedFields and would cause conflicts on value change. Eh. Ugly.
	//
	// Writing custom logic that checks which fields are _really_ managed by a manager, caches the unstructured object and then
	// conditionally removes the label from all those unstructured fields and shoves them back in the proper place
	// in the object managedFields is tedious, ugly and super error-prone.
	//
	// Just remove the fields as the previous owner and quickly readd them as the new one.
	if _, err = nsClient.Update(ctx, nsCopy, metav1.UpdateOptions{FieldManager: "cluster-policy-controller"}); err != nil {
		return nil, fmt.Errorf("failed to share PSa label ownership with the previous owner: %w", err)
	}

	// take ownership of the fields since they should all be clear now
	return nsClient.Apply(ctx, ourOwned, metav1.ApplyOptions{FieldManager: controllerName})
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

	managedNamespaces, err := extractNSFieldsPerManager(ns)
	if err != nil {
		return fmt.Errorf("ns extraction failed: %w", err)
	}

	// we must extract the NS in case only some of the labels we're setting need
	// updating to avoid hotlo
	nsApplyConfig := corev1apply.Namespace(ns.Name)
	if err != nil {
		return fmt.Errorf("failed to extract field ownership for NS %q: %w", ns.Name, err)
	}

	var shouldUpdate bool
	mustSync := ns.Labels[labelSyncControlLabel] == "true"
	for typeLabel, versionLabel := range c.syncedLabels {
		if manager := managedNamespaces.getManagerForLabel(typeLabel); len(manager) == 0 || manager == controllerName || mustSync {
			nsApplyConfig.WithLabels(map[string]string{
				typeLabel: string(psaLevel),
			})
			if ns.Labels[typeLabel] != string(psaLevel) {
				shouldUpdate = true
			}
		}
		if manager := managedNamespaces.getManagerForLabel(versionLabel); len(manager) == 0 || manager == controllerName || mustSync {
			nsApplyConfig.WithLabels(map[string]string{
				versionLabel: currentPSaVersion,
			})
			if ns.Labels[versionLabel] != currentPSaVersion {
				shouldUpdate = true
			}
		}
	}

	if !shouldUpdate {
		return nil
	}

	_, err = c.namespaceClient.Apply(ctx, nsApplyConfig, metav1.ApplyOptions{FieldManager: controllerName, Force: mustSync})
	if err != nil {
		if apierrors.IsConflict(err) {
			klog.Warning("someone else is already managing the PSa labels: %v", err)
			return nil
		}
		return fmt.Errorf("failed to update the namespace: %w", err)
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

	if ns.Labels[labelSyncControlLabel] == "true" {
		return true
	}

	// while "openshift-" namespaces should be considered controlled, there are some
	// edge cases where users can also create them. Consider these a special case
	// and delegate the decision to sync on the user who should know what they are
	// doing when creating a NS that appears to be system-controlled.
	if strings.HasPrefix(nsName, "openshift-") {
		return false
	}

	extractedPerManager, err := extractNSFieldsPerManager(ns)
	if err != nil {
		klog.Errorf("ns extraction failed: %v", err)
		return false
	}

	var owningAtLeastOneLabel bool
	for _, labelName := range []string{
		psapi.EnforceLevelLabel, psapi.EnforceVersionLabel,
		psapi.WarnLevelLabel, psapi.WarnVersionLabel,
		psapi.AuditLevelLabel, psapi.AuditVersionLabel,
	} {
		if _, ok := ns.Labels[labelName]; ok {
			manager := extractedPerManager.getManagerForLabel(labelName)
			if len(manager) > 0 && manager != "cluster-policy-controller" && manager != controllerName {
				continue
			}
		}
		// a label is either not set or is directly owned by us
		owningAtLeastOneLabel = true

	}

	if !owningAtLeastOneLabel {
		return false
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

// extractedNamespaces serves as a cache so that we don't have to re-extract the namespaces
// for each label. It helps us prevent performance overhead from multiple deserializations.
//
// Maps a set of managed metadata.labels to their manager name.
type extractedNamespaces map[string]sets.Set[string]

// extractNSFieldsPerManager parses all the FieldsV1 in a Namespace `ns`,
// extracts the information about label ownership and returns a structure that
// maps all these labels in a set to their manager
func extractNSFieldsPerManager(ns *corev1.Namespace) (extractedNamespaces, error) {
	ret := extractedNamespaces{}
	for _, fieldEntry := range ns.ManagedFields {
		managedLabels, err := managedLabels(fieldEntry)
		if err != nil {
			return nil, fmt.Errorf("failed to extract managed fields for NS %q: %v", ns.Name, err)
		}
		if current, ok := ret[fieldEntry.Manager]; ok {
			ret[fieldEntry.Manager] = current.Union(managedLabels)
		} else {
			ret[fieldEntry.Manager] = managedLabels
		}
	}
	return ret, nil
}

func (n extractedNamespaces) getManagerForLabel(labelName string) string {
	for manager, extractedNS := range n {
		if _, managed := extractedNS[labelName]; managed {
			return manager
		}
	}
	return ""
}

// managedLabels extract the metadata.labels from the JSON in the managedEntry.FieldsV1
// that describes the object's field ownership
func managedLabels(fieldsEntry metav1.ManagedFieldsEntry) (sets.Set[string], error) {
	managedUnstructured := map[string]interface{}{}
	err := json.Unmarshal(fieldsEntry.FieldsV1.Raw, &managedUnstructured)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal managed fields: %w", err)
	}

	labels, found, err := unstructured.NestedMap(managedUnstructured, "f:metadata", "f:labels")
	if err != nil {
		return nil, fmt.Errorf("failed to get labels from the managed fields: %w", err)
	}

	ret := sets.New[string]()
	if !found {
		return ret, nil
	}

	for l := range labels {
		ret.Insert(strings.Replace(l, "f:", "", 1))
	}

	return ret, nil
}
