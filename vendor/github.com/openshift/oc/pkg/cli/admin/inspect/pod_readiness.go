package inspect

import (
	"fmt"

	"k8s.io/api/core/v1"
)

// PodRunningReady checks whether pod p's phase is running and it has a ready
// condition of status true.
func PodRunningReady(p *v1.Pod) (bool, error) {
	if !hasReadyCondition(p) {
		return false, nil
	}

	// Check the phase is running.
	if p.Status.Phase != v1.PodRunning {
		return false, fmt.Errorf("want pod '%s' on '%s' to be '%v' but was '%v'",
			p.ObjectMeta.Name, p.Spec.NodeName, v1.PodRunning, p.Status.Phase)
	}
	// Check the ready condition is true.
	if !IsPodReady(p) {
		return false, fmt.Errorf("pod '%s' on '%s' didn't have condition {%v %v}; conditions: %v",
			p.ObjectMeta.Name, p.Spec.NodeName, v1.PodReady, v1.ConditionTrue, p.Status.Conditions)
	}
	return true, nil
}

func hasReadyCondition(pod *v1.Pod) bool {
	conditionReady := true
	for _, cond := range pod.Status.Conditions {
		if cond.Type != v1.PodReady {
			continue
		}
		conditionReady = cond.Status == v1.ConditionTrue
		break
	}
	return conditionReady
}

// IsPodReady returns true if a pod is ready; false otherwise.
func IsPodReady(pod *v1.Pod) bool {
	return IsPodReadyConditionTrue(pod.Status)
}

// IsPodReady returns true if a pod is ready; false otherwise.
func IsPodReadyConditionTrue(status v1.PodStatus) bool {
	condition := GetPodReadyCondition(status)
	return condition != nil && condition.Status == v1.ConditionTrue
}

// Extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
func GetPodReadyCondition(status v1.PodStatus) *v1.PodCondition {
	_, condition := GetPodCondition(&status, v1.PodReady)
	return condition
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetPodCondition(status *v1.PodStatus, conditionType v1.PodConditionType) (int, *v1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return GetPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func GetPodConditionFromList(conditions []v1.PodCondition, conditionType v1.PodConditionType) (int, *v1.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}
