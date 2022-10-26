/*
Copyright 2021 The Kubernetes Authors.

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

package job

import (
	"fmt"

	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

// matchPodFailurePolicy returns information about matching a given failed pod
// against the pod failure policy rules. The information is represented as an
// optional job failure message (present in case the pod matched a 'FailJob'
// rule) and a boolean indicating if the failure should be counted towards
// backoffLimit (it should not be counted if the pod matched an 'Ignore' rule).
func matchPodFailurePolicy(podFailurePolicy *batch.PodFailurePolicy, failedPod *v1.Pod) (*string, bool) {
	if podFailurePolicy == nil {
		return nil, true
	}
	for index, podFailurePolicyRule := range podFailurePolicy.Rules {
		if podFailurePolicyRule.OnExitCodes != nil {
			if containerStatus := matchOnExitCodes(&failedPod.Status, podFailurePolicyRule.OnExitCodes); containerStatus != nil {
				switch podFailurePolicyRule.Action {
				case batch.PodFailurePolicyActionIgnore:
					return nil, false
				case batch.PodFailurePolicyActionCount:
					return nil, true
				case batch.PodFailurePolicyActionFailJob:
					msg := fmt.Sprintf("Container %s for pod %s/%s failed with exit code %v matching %v rule at index %d",
						containerStatus.Name, failedPod.Namespace, failedPod.Name, containerStatus.State.Terminated.ExitCode, podFailurePolicyRule.Action, index)
					return &msg, true
				}
			}
		} else if podFailurePolicyRule.OnPodConditions != nil {
			if podCondition := matchOnPodConditions(&failedPod.Status, podFailurePolicyRule.OnPodConditions); podCondition != nil {
				switch podFailurePolicyRule.Action {
				case batch.PodFailurePolicyActionIgnore:
					return nil, false
				case batch.PodFailurePolicyActionCount:
					return nil, true
				case batch.PodFailurePolicyActionFailJob:
					msg := fmt.Sprintf("Pod %s/%s has condition %v matching %v rule at index %d",
						failedPod.Namespace, failedPod.Name, podCondition.Type, podFailurePolicyRule.Action, index)
					return &msg, true
				}
			}
		}
	}
	return nil, true
}

func matchOnExitCodes(podStatus *v1.PodStatus, requirement *batch.PodFailurePolicyOnExitCodesRequirement) *v1.ContainerStatus {
	if containerStatus := getMatchingContainerFromList(podStatus.ContainerStatuses, requirement); containerStatus != nil {
		return containerStatus
	}
	return getMatchingContainerFromList(podStatus.InitContainerStatuses, requirement)
}

func matchOnPodConditions(podStatus *v1.PodStatus, requirement []batch.PodFailurePolicyOnPodConditionsPattern) *v1.PodCondition {
	for _, podCondition := range podStatus.Conditions {
		for _, pattern := range requirement {
			if podCondition.Type == pattern.Type && podCondition.Status == pattern.Status {
				return &podCondition
			}
		}
	}
	return nil
}

func getMatchingContainerFromList(containerStatuses []v1.ContainerStatus, requirement *batch.PodFailurePolicyOnExitCodesRequirement) *v1.ContainerStatus {
	for _, containerStatus := range containerStatuses {
		if requirement.ContainerName == nil || *requirement.ContainerName == containerStatus.Name {
			if containerStatus.State.Terminated.ExitCode != 0 {
				if isOnExitCodesOperatorMatching(containerStatus.State.Terminated.ExitCode, requirement) {
					return &containerStatus
				}
			}
		}
	}
	return nil
}

func isOnExitCodesOperatorMatching(exitCode int32, requirement *batch.PodFailurePolicyOnExitCodesRequirement) bool {
	switch requirement.Operator {
	case batch.PodFailurePolicyOnExitCodesOpIn:
		for _, value := range requirement.Values {
			if value == exitCode {
				return true
			}
		}
		return false
	case batch.PodFailurePolicyOnExitCodesOpNotIn:
		for _, value := range requirement.Values {
			if value == exitCode {
				return false
			}
		}
		return true
	default:
		return false
	}
}
