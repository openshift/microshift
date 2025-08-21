package healthcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func Test_analyzeEventsLookingForUnpulledOrFailedImages(t *testing.T) {
	testCases := []struct {
		name              string
		existingPodsNames sets.Set[string]
		pullingEvents     *corev1.EventList
		pulledEvents      *corev1.EventList
		failedEvents      *corev1.EventList
		expectedUnpulled  []unpulledImage
		expectedFailed    []failedImage
	}{
		{
			name:              "no events",
			existingPodsNames: sets.New[string](),
			pullingEvents:     &corev1.EventList{},
			pulledEvents:      &corev1.EventList{},
			failedEvents:      &corev1.EventList{},
			expectedUnpulled:  []unpulledImage{},
			expectedFailed:    []failedImage{},
		},
		{
			name:              "image still being pulled",
			existingPodsNames: sets.New("test-pod"),
			pullingEvents: &corev1.EventList{
				Items: []corev1.Event{
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "test-pod",
							Namespace: "test-ns",
						},
						Message: `Pulling image "nginx:latest"`,
					},
				},
			},
			pulledEvents: &corev1.EventList{},
			failedEvents: &corev1.EventList{},
			expectedUnpulled: []unpulledImage{
				{Namespace: "test-ns", PodName: "test-pod", Image: "nginx:latest"},
			},
			expectedFailed: []failedImage{},
		},
		{
			name:              "image successfully pulled",
			existingPodsNames: sets.New("test-pod"),
			pullingEvents: &corev1.EventList{
				Items: []corev1.Event{
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "test-pod",
							Namespace: "test-ns",
						},
						Message: `Pulling image "nginx:latest"`,
					},
				},
			},
			pulledEvents: &corev1.EventList{
				Items: []corev1.Event{
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "test-pod",
							Namespace: "test-ns",
						},
						Message: `Successfully pulled image "nginx:latest"`,
					},
				},
			},
			failedEvents:     &corev1.EventList{},
			expectedUnpulled: []unpulledImage{},
			expectedFailed:   []failedImage{},
		},
		{
			name:              "image failed to pull",
			existingPodsNames: sets.New("test-pod"),
			pullingEvents: &corev1.EventList{
				Items: []corev1.Event{
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "test-pod",
							Namespace: "test-ns",
						},
						Message: `Pulling image "nginx:latest"`,
					},
				},
			},
			pulledEvents: &corev1.EventList{},
			failedEvents: &corev1.EventList{
				Items: []corev1.Event{
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "test-pod",
							Namespace: "test-ns",
						},
						Message: `Failed to pull image "nginx:latest": error message`,
					},
				},
			},
			expectedUnpulled: []unpulledImage{},
			expectedFailed: []failedImage{
				{
					unpulledImage: unpulledImage{Namespace: "test-ns", PodName: "test-pod", Image: "nginx:latest"},
					Message:       `Failed to pull image "nginx:latest": error message`,
				},
			},
		},
		{
			name:              "skip events for non-existing pods",
			existingPodsNames: sets.New("existing-pod"),
			pullingEvents: &corev1.EventList{
				Items: []corev1.Event{
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "deleted-pod",
							Namespace: "test-ns",
						},
						Message: `Pulling image "nginx:latest"`,
					},
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "existing-pod",
							Namespace: "test-ns",
						},
						Message: `Pulling image "redis:latest"`,
					},
				},
			},
			pulledEvents: &corev1.EventList{},
			failedEvents: &corev1.EventList{},
			expectedUnpulled: []unpulledImage{
				{Namespace: "test-ns", PodName: "existing-pod", Image: "redis:latest"},
			},
			expectedFailed: []failedImage{},
		},
		{
			name:              "multiple images with mixed states",
			existingPodsNames: sets.New("pod1", "pod2", "pod3"),
			pullingEvents: &corev1.EventList{
				Items: []corev1.Event{
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "pod1",
							Namespace: "ns1",
						},
						Message: `Pulling image "nginx:latest"`,
					},
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "pod2",
							Namespace: "ns2",
						},
						Message: `Pulling image "redis:latest"`,
					},
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "pod3",
							Namespace: "ns3",
						},
						Message: `Pulling image "postgres:13"`,
					},
				},
			},
			pulledEvents: &corev1.EventList{
				Items: []corev1.Event{
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "pod1",
							Namespace: "ns1",
						},
						Message: `Successfully pulled image "nginx:latest"`,
					},
				},
			},
			failedEvents: &corev1.EventList{
				Items: []corev1.Event{
					{
						InvolvedObject: corev1.ObjectReference{
							Name:      "pod2",
							Namespace: "ns2",
						},
						Message: `Failed to pull image "redis:latest": connection timeout`,
					},
				},
			},
			expectedUnpulled: []unpulledImage{
				{Namespace: "ns3", PodName: "pod3", Image: "postgres:13"},
			},
			expectedFailed: []failedImage{
				{
					unpulledImage: unpulledImage{Namespace: "ns2", PodName: "pod2", Image: "redis:latest"},
					Message:       `Failed to pull image "redis:latest": connection timeout`,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unpulled, failed := analyzeEventsLookingForUnpulledOrFailedImages(
				tc.existingPodsNames,
				tc.pullingEvents,
				tc.pulledEvents,
				tc.failedEvents,
			)

			assert.ElementsMatch(t, tc.expectedUnpulled, unpulled)
			assert.ElementsMatch(t, tc.expectedFailed, failed)
		})
	}
}
