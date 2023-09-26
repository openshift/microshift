/*
Copyright 2016 The Kubernetes Authors.

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

package statefulset

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/history"
	"k8s.io/kubernetes/pkg/features"
)

var patchCodec = scheme.Codecs.LegacyCodec(apps.SchemeGroupVersion)

// overlappingStatefulSets sorts a list of StatefulSets by creation timestamp, using their names as a tie breaker.
// Generally used to tie break between StatefulSets that have overlapping selectors.
type overlappingStatefulSets []*apps.StatefulSet

func (o overlappingStatefulSets) Len() int { return len(o) }

func (o overlappingStatefulSets) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func (o overlappingStatefulSets) Less(i, j int) bool {
	if o[i].CreationTimestamp.Equal(&o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].CreationTimestamp.Before(&o[j].CreationTimestamp)
}

// statefulPodRegex is a regular expression that extracts the parent StatefulSet and ordinal from the Name of a Pod
var statefulPodRegex = regexp.MustCompile("(.*)-([0-9]+)$")

// getParentNameAndOrdinal gets the name of pod's parent StatefulSet and pod's ordinal as extracted from its Name. If
// the Pod was not created by a StatefulSet, its parent is considered to be empty string, and its ordinal is considered
// to be -1.
func getParentNameAndOrdinal(pod *v1.Pod) (string, int) {
	parent := ""
	ordinal := -1
	subMatches := statefulPodRegex.FindStringSubmatch(pod.Name)
	if len(subMatches) < 3 {
		return parent, ordinal
	}
	parent = subMatches[1]
	if i, err := strconv.ParseInt(subMatches[2], 10, 32); err == nil {
		ordinal = int(i)
	}
	return parent, ordinal
}

// getParentName gets the name of pod's parent StatefulSet. If pod has not parent, the empty string is returned.
func getParentName(pod *v1.Pod) string {
	parent, _ := getParentNameAndOrdinal(pod)
	return parent
}

// getOrdinal gets pod's ordinal. If pod has no ordinal, -1 is returned.
func getOrdinal(pod *v1.Pod) int {
	_, ordinal := getParentNameAndOrdinal(pod)
	return ordinal
}

// getStartOrdinal gets the first possible ordinal (inclusive).
// Returns spec.ordinals.start if spec.ordinals is set, otherwise returns 0.
func getStartOrdinal(set *apps.StatefulSet) int {
	if utilfeature.DefaultFeatureGate.Enabled(features.StatefulSetStartOrdinal) {
		if set.Spec.Ordinals != nil {
			return int(set.Spec.Ordinals.Start)
		}
	}
	return 0
}

// getEndOrdinal gets the last possible ordinal (inclusive).
func getEndOrdinal(set *apps.StatefulSet) int {
	return getStartOrdinal(set) + int(*set.Spec.Replicas) - 1
}

// podInOrdinalRange returns true if the pod ordinal is within the allowed
// range of ordinals that this StatefulSet is set to control.
func podInOrdinalRange(pod *v1.Pod, set *apps.StatefulSet) bool {
	ordinal := getOrdinal(pod)
	return ordinal >= getStartOrdinal(set) && ordinal <= getEndOrdinal(set)
}

// getPodName gets the name of set's child Pod with an ordinal index of ordinal
func getPodName(set *apps.StatefulSet, ordinal int) string {
	return fmt.Sprintf("%s-%d", set.Name, ordinal)
}

// getPersistentVolumeClaimName gets the name of PersistentVolumeClaim for a Pod with an ordinal index of ordinal. claim
// must be a PersistentVolumeClaim from set's VolumeClaims template.
func getPersistentVolumeClaimName(set *apps.StatefulSet, claim *v1.PersistentVolumeClaim, ordinal int) string {
	// NOTE: This name format is used by the heuristics for zone spreading in ChooseZoneForVolume
	return fmt.Sprintf("%s-%s-%d", claim.Name, set.Name, ordinal)
}

// isMemberOf tests if pod is a member of set.
func isMemberOf(set *apps.StatefulSet, pod *v1.Pod) bool {
	return getParentName(pod) == set.Name
}

// identityMatches returns true if pod has a valid identity and network identity for a member of set.
func identityMatches(set *apps.StatefulSet, pod *v1.Pod) bool {
	parent, ordinal := getParentNameAndOrdinal(pod)
	return ordinal >= 0 &&
		set.Name == parent &&
		pod.Name == getPodName(set, ordinal) &&
		pod.Namespace == set.Namespace &&
		pod.Labels[apps.StatefulSetPodNameLabel] == pod.Name
}

// storageMatches returns true if pod's Volumes cover the set of PersistentVolumeClaims
func storageMatches(set *apps.StatefulSet, pod *v1.Pod) bool {
	ordinal := getOrdinal(pod)
	if ordinal < 0 {
		return false
	}
	volumes := make(map[string]v1.Volume, len(pod.Spec.Volumes))
	for _, volume := range pod.Spec.Volumes {
		volumes[volume.Name] = volume
	}
	for _, claim := range set.Spec.VolumeClaimTemplates {
		volume, found := volumes[claim.Name]
		if !found ||
			volume.VolumeSource.PersistentVolumeClaim == nil ||
			volume.VolumeSource.PersistentVolumeClaim.ClaimName !=
				getPersistentVolumeClaimName(set, &claim, ordinal) {
			return false
		}
	}
	return true
}

// getPersistentVolumeClaimPolicy returns the PVC policy for a StatefulSet, returning a retain policy if the set policy is nil.
func getPersistentVolumeClaimRetentionPolicy(set *apps.StatefulSet) apps.StatefulSetPersistentVolumeClaimRetentionPolicy {
	policy := apps.StatefulSetPersistentVolumeClaimRetentionPolicy{
		WhenDeleted: apps.RetainPersistentVolumeClaimRetentionPolicyType,
		WhenScaled:  apps.RetainPersistentVolumeClaimRetentionPolicyType,
	}
	if set.Spec.PersistentVolumeClaimRetentionPolicy != nil {
		policy = *set.Spec.PersistentVolumeClaimRetentionPolicy
	}
	return policy
}

// claimOwnerMatchesSetAndPod returns false if the ownerRefs of the claim are not set consistently with the
// PVC deletion policy for the StatefulSet.
func claimOwnerMatchesSetAndPod(logger klog.Logger, claim *v1.PersistentVolumeClaim, set *apps.StatefulSet, pod *v1.Pod) bool {
	policy := getPersistentVolumeClaimRetentionPolicy(set)
	const retain = apps.RetainPersistentVolumeClaimRetentionPolicyType
	const delete = apps.DeletePersistentVolumeClaimRetentionPolicyType
	switch {
	default:
		logger.Error(nil, "Unknown policy, treating as Retain", "policy", set.Spec.PersistentVolumeClaimRetentionPolicy)
		fallthrough
	case policy.WhenScaled == retain && policy.WhenDeleted == retain:
		if hasOwnerRef(claim, set) ||
			hasOwnerRef(claim, pod) {
			return false
		}
	case policy.WhenScaled == retain && policy.WhenDeleted == delete:
		if !hasOwnerRef(claim, set) ||
			hasOwnerRef(claim, pod) {
			return false
		}
	case policy.WhenScaled == delete && policy.WhenDeleted == retain:
		if hasOwnerRef(claim, set) {
			return false
		}
		podScaledDown := !podInOrdinalRange(pod, set)
		if podScaledDown != hasOwnerRef(claim, pod) {
			return false
		}
	case policy.WhenScaled == delete && policy.WhenDeleted == delete:
		podScaledDown := !podInOrdinalRange(pod, set)
		// If a pod is scaled down, there should be no set ref and a pod ref;
		// if the pod is not scaled down it's the other way around.
		if podScaledDown == hasOwnerRef(claim, set) {
			return false
		}
		if podScaledDown != hasOwnerRef(claim, pod) {
			return false
		}
	}
	return true
}

// updateClaimOwnerRefForSetAndPod updates the ownerRefs for the claim according to the deletion policy of
// the StatefulSet. Returns true if the claim was changed and should be updated and false otherwise.
func updateClaimOwnerRefForSetAndPod(logger klog.Logger, claim *v1.PersistentVolumeClaim, set *apps.StatefulSet, pod *v1.Pod) bool {
	needsUpdate := false
	// Sometimes the version and kind are not set {pod,set}.TypeMeta. These are necessary for the ownerRef.
	// This is the case both in real clusters and the unittests.
	// TODO: there must be a better way to do this other than hardcoding the pod version?
	updateMeta := func(tm *metav1.TypeMeta, kind string) {
		if tm.APIVersion == "" {
			if kind == "StatefulSet" {
				tm.APIVersion = "apps/v1"
			} else {
				tm.APIVersion = "v1"
			}
		}
		if tm.Kind == "" {
			tm.Kind = kind
		}
	}
	podMeta := pod.TypeMeta
	updateMeta(&podMeta, "Pod")
	setMeta := set.TypeMeta
	updateMeta(&setMeta, "StatefulSet")
	policy := getPersistentVolumeClaimRetentionPolicy(set)
	const retain = apps.RetainPersistentVolumeClaimRetentionPolicyType
	const delete = apps.DeletePersistentVolumeClaimRetentionPolicyType
	switch {
	default:
		logger.Error(nil, "Unknown policy, treating as Retain", "policy", set.Spec.PersistentVolumeClaimRetentionPolicy)
		fallthrough
	case policy.WhenScaled == retain && policy.WhenDeleted == retain:
		needsUpdate = removeOwnerRef(claim, set) || needsUpdate
		needsUpdate = removeOwnerRef(claim, pod) || needsUpdate
	case policy.WhenScaled == retain && policy.WhenDeleted == delete:
		needsUpdate = setOwnerRef(claim, set, &setMeta) || needsUpdate
		needsUpdate = removeOwnerRef(claim, pod) || needsUpdate
	case policy.WhenScaled == delete && policy.WhenDeleted == retain:
		needsUpdate = removeOwnerRef(claim, set) || needsUpdate
		podScaledDown := !podInOrdinalRange(pod, set)
		if podScaledDown {
			needsUpdate = setOwnerRef(claim, pod, &podMeta) || needsUpdate
		}
		if !podScaledDown {
			needsUpdate = removeOwnerRef(claim, pod) || needsUpdate
		}
	case policy.WhenScaled == delete && policy.WhenDeleted == delete:
		podScaledDown := !podInOrdinalRange(pod, set)
		if podScaledDown {
			needsUpdate = removeOwnerRef(claim, set) || needsUpdate
			needsUpdate = setOwnerRef(claim, pod, &podMeta) || needsUpdate
		}
		if !podScaledDown {
			needsUpdate = setOwnerRef(claim, set, &setMeta) || needsUpdate
			needsUpdate = removeOwnerRef(claim, pod) || needsUpdate
		}
	}
	return needsUpdate
}

// hasOwnerRef returns true if target has an ownerRef to owner.
func hasOwnerRef(target, owner metav1.Object) bool {
	ownerUID := owner.GetUID()
	for _, ownerRef := range target.GetOwnerReferences() {
		if ownerRef.UID == ownerUID {
			return true
		}
	}
	return false
}

// hasStaleOwnerRef returns true if target has a ref to owner that appears to be stale.
func hasStaleOwnerRef(target, owner metav1.Object) bool {
	for _, ownerRef := range target.GetOwnerReferences() {
		if ownerRef.Name == owner.GetName() && ownerRef.UID != owner.GetUID() {
			return true
		}
	}
	return false
}

// setOwnerRef adds owner to the ownerRefs of target, if necessary. Returns true if target needs to be
// updated and false otherwise.
func setOwnerRef(target, owner metav1.Object, ownerType *metav1.TypeMeta) bool {
	if hasOwnerRef(target, owner) {
		return false
	}
	ownerRefs := append(
		target.GetOwnerReferences(),
		metav1.OwnerReference{
			APIVersion: ownerType.APIVersion,
			Kind:       ownerType.Kind,
			Name:       owner.GetName(),
			UID:        owner.GetUID(),
		})
	target.SetOwnerReferences(ownerRefs)
	return true
}

// removeOwnerRef removes owner from the ownerRefs of target, if necessary. Returns true if target needs
// to be updated and false otherwise.
func removeOwnerRef(target, owner metav1.Object) bool {
	if !hasOwnerRef(target, owner) {
		return false
	}
	ownerUID := owner.GetUID()
	oldRefs := target.GetOwnerReferences()
	newRefs := make([]metav1.OwnerReference, len(oldRefs)-1)
	skip := 0
	for i := range oldRefs {
		if oldRefs[i].UID == ownerUID {
			skip = -1
		} else {
			newRefs[i+skip] = oldRefs[i]
		}
	}
	target.SetOwnerReferences(newRefs)
	return true
}

// getPersistentVolumeClaims gets a map of PersistentVolumeClaims to their template names, as defined in set. The
// returned PersistentVolumeClaims are each constructed with a the name specific to the Pod. This name is determined
// by getPersistentVolumeClaimName.
func getPersistentVolumeClaims(set *apps.StatefulSet, pod *v1.Pod) map[string]v1.PersistentVolumeClaim {
	ordinal := getOrdinal(pod)
	templates := set.Spec.VolumeClaimTemplates
	claims := make(map[string]v1.PersistentVolumeClaim, len(templates))
	for i := range templates {
		claim := templates[i]
		claim.Name = getPersistentVolumeClaimName(set, &claim, ordinal)
		claim.Namespace = set.Namespace
		if claim.Labels != nil {
			for key, value := range set.Spec.Selector.MatchLabels {
				claim.Labels[key] = value
			}
		} else {
			claim.Labels = set.Spec.Selector.MatchLabels
		}
		claims[templates[i].Name] = claim
	}
	return claims
}

// updateStorage updates pod's Volumes to conform with the PersistentVolumeClaim of set's templates. If pod has
// conflicting local Volumes these are replaced with Volumes that conform to the set's templates.
func updateStorage(set *apps.StatefulSet, pod *v1.Pod) {
	currentVolumes := pod.Spec.Volumes
	claims := getPersistentVolumeClaims(set, pod)
	newVolumes := make([]v1.Volume, 0, len(claims))
	for name, claim := range claims {
		newVolumes = append(newVolumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: claim.Name,
					// TODO: Use source definition to set this value when we have one.
					ReadOnly: false,
				},
			},
		})
	}
	for i := range currentVolumes {
		if _, ok := claims[currentVolumes[i].Name]; !ok {
			newVolumes = append(newVolumes, currentVolumes[i])
		}
	}
	pod.Spec.Volumes = newVolumes
}

func initIdentity(set *apps.StatefulSet, pod *v1.Pod) {
	updateIdentity(set, pod)
	// Set these immutable fields only on initial Pod creation, not updates.
	pod.Spec.Hostname = pod.Name
	pod.Spec.Subdomain = set.Spec.ServiceName
}

// updateIdentity updates pod's name, hostname, and subdomain, and StatefulSetPodNameLabel to conform to set's name
// and headless service.
func updateIdentity(set *apps.StatefulSet, pod *v1.Pod) {
	ordinal := getOrdinal(pod)
	pod.Name = getPodName(set, ordinal)
	pod.Namespace = set.Namespace
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[apps.StatefulSetPodNameLabel] = pod.Name
	if utilfeature.DefaultFeatureGate.Enabled(features.PodIndexLabel) {
		pod.Labels[apps.PodIndexLabel] = strconv.Itoa(ordinal)
	}
}

// isRunningAndReady returns true if pod is in the PodRunning Phase, if it has a condition of PodReady.
func isRunningAndReady(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodRunning && podutil.IsPodReady(pod)
}

func isRunningAndAvailable(pod *v1.Pod, minReadySeconds int32) bool {
	return podutil.IsPodAvailable(pod, minReadySeconds, metav1.Now())
}

// isCreated returns true if pod has been created and is maintained by the API server
func isCreated(pod *v1.Pod) bool {
	return pod.Status.Phase != ""
}

// isPending returns true if pod has a Phase of PodPending
func isPending(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodPending
}

// isFailed returns true if pod has a Phase of PodFailed
func isFailed(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodFailed
}

// isTerminating returns true if pod's DeletionTimestamp has been set
func isTerminating(pod *v1.Pod) bool {
	return pod.DeletionTimestamp != nil
}

// isHealthy returns true if pod is running and ready and has not been terminated
func isHealthy(pod *v1.Pod) bool {
	return isRunningAndReady(pod) && !isTerminating(pod)
}

// allowsBurst is true if the alpha burst annotation is set.
func allowsBurst(set *apps.StatefulSet) bool {
	return set.Spec.PodManagementPolicy == apps.ParallelPodManagement
}

// setPodRevision sets the revision of Pod to revision by adding the StatefulSetRevisionLabel
func setPodRevision(pod *v1.Pod, revision string) {
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[apps.StatefulSetRevisionLabel] = revision
}

// getPodRevision gets the revision of Pod by inspecting the StatefulSetRevisionLabel. If pod has no revision the empty
// string is returned.
func getPodRevision(pod *v1.Pod) string {
	if pod.Labels == nil {
		return ""
	}
	return pod.Labels[apps.StatefulSetRevisionLabel]
}

// newStatefulSetPod returns a new Pod conforming to the set's Spec with an identity generated from ordinal.
func newStatefulSetPod(set *apps.StatefulSet, ordinal int) *v1.Pod {
	pod, _ := controller.GetPodFromTemplate(&set.Spec.Template, set, metav1.NewControllerRef(set, controllerKind))
	pod.Name = getPodName(set, ordinal)
	initIdentity(set, pod)
	updateStorage(set, pod)
	return pod
}

// newVersionedStatefulSetPod creates a new Pod for a StatefulSet. currentSet is the representation of the set at the
// current revision. updateSet is the representation of the set at the updateRevision. currentRevision is the name of
// the current revision. updateRevision is the name of the update revision. ordinal is the ordinal of the Pod. If the
// returned error is nil, the returned Pod is valid.
func newVersionedStatefulSetPod(currentSet, updateSet *apps.StatefulSet, currentRevision, updateRevision string, ordinal int) *v1.Pod {
	if currentSet.Spec.UpdateStrategy.Type == apps.RollingUpdateStatefulSetStrategyType &&
		(currentSet.Spec.UpdateStrategy.RollingUpdate == nil && ordinal < (getStartOrdinal(currentSet)+int(currentSet.Status.CurrentReplicas))) ||
		(currentSet.Spec.UpdateStrategy.RollingUpdate != nil && ordinal < (getStartOrdinal(currentSet)+int(*currentSet.Spec.UpdateStrategy.RollingUpdate.Partition))) {
		pod := newStatefulSetPod(currentSet, ordinal)
		setPodRevision(pod, currentRevision)
		return pod
	}
	pod := newStatefulSetPod(updateSet, ordinal)
	setPodRevision(pod, updateRevision)
	return pod
}

// getPatch returns a strategic merge patch that can be applied to restore a StatefulSet to a
// previous version. If the returned error is nil the patch is valid. The current state that we save is just the
// PodSpecTemplate. We can modify this later to encompass more state (or less) and remain compatible with previously
// recorded patches.
func getPatch(set *apps.StatefulSet) ([]byte, error) {
	data, err := runtime.Encode(patchCodec, set)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	objCopy := make(map[string]interface{})
	specCopy := make(map[string]interface{})
	spec := raw["spec"].(map[string]interface{})
	template := spec["template"].(map[string]interface{})
	specCopy["template"] = template
	template["$patch"] = "replace"
	objCopy["spec"] = specCopy
	patch, err := json.Marshal(objCopy)
	return patch, err
}

// newRevision creates a new ControllerRevision containing a patch that reapplies the target state of set.
// The Revision of the returned ControllerRevision is set to revision. If the returned error is nil, the returned
// ControllerRevision is valid. StatefulSet revisions are stored as patches that re-apply the current state of set
// to a new StatefulSet using a strategic merge patch to replace the saved state of the new StatefulSet.
func newRevision(set *apps.StatefulSet, revision int64, collisionCount *int32) (*apps.ControllerRevision, error) {
	patch, err := getPatch(set)
	if err != nil {
		return nil, err
	}
	cr, err := history.NewControllerRevision(set,
		controllerKind,
		set.Spec.Template.Labels,
		runtime.RawExtension{Raw: patch},
		revision,
		collisionCount)
	if err != nil {
		return nil, err
	}
	if cr.ObjectMeta.Annotations == nil {
		cr.ObjectMeta.Annotations = make(map[string]string)
	}
	for key, value := range set.Annotations {
		cr.ObjectMeta.Annotations[key] = value
	}
	return cr, nil
}

// ApplyRevision returns a new StatefulSet constructed by restoring the state in revision to set. If the returned error
// is nil, the returned StatefulSet is valid.
func ApplyRevision(set *apps.StatefulSet, revision *apps.ControllerRevision) (*apps.StatefulSet, error) {
	clone := set.DeepCopy()
	patched, err := strategicpatch.StrategicMergePatch([]byte(runtime.EncodeOrDie(patchCodec, clone)), revision.Data.Raw, clone)
	if err != nil {
		return nil, err
	}
	restoredSet := &apps.StatefulSet{}
	err = json.Unmarshal(patched, restoredSet)
	if err != nil {
		return nil, err
	}
	return restoredSet, nil
}

// nextRevision finds the next valid revision number based on revisions. If the length of revisions
// is 0 this is 1. Otherwise, it is 1 greater than the largest revision's Revision. This method
// assumes that revisions has been sorted by Revision.
func nextRevision(revisions []*apps.ControllerRevision) int64 {
	count := len(revisions)
	if count <= 0 {
		return 1
	}
	return revisions[count-1].Revision + 1
}

// inconsistentStatus returns true if the ObservedGeneration of status is greater than set's
// Generation or if any of the status's fields do not match those of set's status.
func inconsistentStatus(set *apps.StatefulSet, status *apps.StatefulSetStatus) bool {
	return status.ObservedGeneration > set.Status.ObservedGeneration ||
		status.Replicas != set.Status.Replicas ||
		status.CurrentReplicas != set.Status.CurrentReplicas ||
		status.ReadyReplicas != set.Status.ReadyReplicas ||
		status.UpdatedReplicas != set.Status.UpdatedReplicas ||
		status.CurrentRevision != set.Status.CurrentRevision ||
		status.AvailableReplicas != set.Status.AvailableReplicas ||
		status.UpdateRevision != set.Status.UpdateRevision
}

// completeRollingUpdate completes a rolling update when all of set's replica Pods have been updated
// to the updateRevision. status's currentRevision is set to updateRevision and its' updateRevision
// is set to the empty string. status's currentReplicas is set to updateReplicas and its updateReplicas
// are set to 0.
func completeRollingUpdate(set *apps.StatefulSet, status *apps.StatefulSetStatus) {
	if set.Spec.UpdateStrategy.Type == apps.RollingUpdateStatefulSetStrategyType &&
		status.UpdatedReplicas == status.Replicas &&
		status.ReadyReplicas == status.Replicas {
		status.CurrentReplicas = status.UpdatedReplicas
		status.CurrentRevision = status.UpdateRevision
	}
}

// ascendingOrdinal is a sort.Interface that Sorts a list of Pods based on the ordinals extracted
// from the Pod. Pod's that have not been constructed by StatefulSet's have an ordinal of -1, and are therefore pushed
// to the front of the list.
type ascendingOrdinal []*v1.Pod

func (ao ascendingOrdinal) Len() int {
	return len(ao)
}

func (ao ascendingOrdinal) Swap(i, j int) {
	ao[i], ao[j] = ao[j], ao[i]
}

func (ao ascendingOrdinal) Less(i, j int) bool {
	return getOrdinal(ao[i]) < getOrdinal(ao[j])
}

// descendingOrdinal is a sort.Interface that Sorts a list of Pods based on the ordinals extracted
// from the Pod. Pod's that have not been constructed by StatefulSet's have an ordinal of -1, and are therefore pushed
// to the end of the list.
type descendingOrdinal []*v1.Pod

func (do descendingOrdinal) Len() int {
	return len(do)
}

func (do descendingOrdinal) Swap(i, j int) {
	do[i], do[j] = do[j], do[i]
}

func (do descendingOrdinal) Less(i, j int) bool {
	return getOrdinal(do[i]) > getOrdinal(do[j])
}

// getStatefulSetMaxUnavailable calculates the real maxUnavailable number according to the replica count
// and maxUnavailable from rollingUpdateStrategy. The number defaults to 1 if the maxUnavailable field is
// not set, and it will be round down to at least 1 if the maxUnavailable value is a percentage.
// Note that API validation has already guaranteed the maxUnavailable field to be >1 if it is an integer
// or 0% < value <= 100% if it is a percentage, so we don't have to consider other cases.
func getStatefulSetMaxUnavailable(maxUnavailable *intstr.IntOrString, replicaCount int) (int, error) {
	maxUnavailableNum, err := intstr.GetScaledValueFromIntOrPercent(intstr.ValueOrDefault(maxUnavailable, intstr.FromInt32(1)), replicaCount, false)
	if err != nil {
		return 0, err
	}
	// maxUnavailable might be zero for small percentage with round down.
	// So we have to enforce it not to be less than 1.
	if maxUnavailableNum < 1 {
		maxUnavailableNum = 1
	}
	return maxUnavailableNum, nil
}
