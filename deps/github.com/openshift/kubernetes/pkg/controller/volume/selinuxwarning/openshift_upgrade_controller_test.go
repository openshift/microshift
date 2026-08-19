package selinuxwarning

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2/ktesting"
	volumecache "k8s.io/kubernetes/pkg/controller/volume/selinuxwarning/cache"
)

var _ volumecache.ConflictCounter = &fakeVolumeCache{}

func (f *fakeVolumeCache) GetConflictCount() int {
	count := 0
	for _, conflicts := range f.conflictsToSend {
		count += len(conflicts)
	}
	return count
}

func TestReportSELinuxConflicts(t *testing.T) {
	cmTrue := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: configMapNamespace,
		},
		Data: map[string]string{
			"conflictsPresent": "True",
		},
	}
	cmFalse := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: configMapNamespace,
		},
		Data: map[string]string{
			"conflictsPresent": "False",
		},
	}
	tests := []struct {
		name            string
		conflicts       map[cache.ObjectName][]volumecache.Conflict
		initialConflict metav1.ConditionStatus
		// If set, the ConfigMap already exists before the test runs.
		existingConfigMap *v1.ConfigMap

		expectConfigMapData map[string]string
		// If true, no ConfigMap write is expected (status didn't change).
		expectNoWrite bool
	}{
		{
			name:            "no conflicts, create the config map",
			initialConflict: metav1.ConditionUnknown,
			conflicts:       nil,
			expectConfigMapData: map[string]string{
				"conflictsPresent": "False",
			},
		},
		{
			name:            "conflicts present, create the config map",
			initialConflict: metav1.ConditionUnknown,
			conflicts: map[cache.ObjectName][]volumecache.Conflict{
				{Namespace: "ns1", Name: "pod1"}: {
					{
						PropertyName:       "SELinuxLabel",
						EventReason:        "SELinuxLabelConflict",
						Pod:                cache.ObjectName{Namespace: "ns1", Name: "pod1"},
						PropertyValue:      ":::s0:c1,c2",
						OtherPod:           cache.ObjectName{Namespace: "ns1", Name: "pod2"},
						OtherPropertyValue: ":::s0:c98,c99",
					},
				},
			},
			expectConfigMapData: map[string]string{
				"conflictsPresent": "True",
			},
		},
		{
			name:              "no conflicts, status was already False",
			conflicts:         nil,
			initialConflict:   metav1.ConditionFalse,
			existingConfigMap: cmFalse,
			expectNoWrite:     true,
		},
		{
			name: "conflicts present, status was already True",
			conflicts: map[cache.ObjectName][]volumecache.Conflict{
				{Namespace: "ns1", Name: "pod1"}: {
					{
						PropertyName: "SELinuxLabel",
						EventReason:  "SELinuxLabelConflict",
					},
				},
			},
			initialConflict:   metav1.ConditionTrue,
			existingConfigMap: cmTrue,
			expectNoWrite:     true,
		},
		{
			name:              "no conflicts, status changes from True to False",
			conflicts:         nil,
			initialConflict:   metav1.ConditionTrue,
			existingConfigMap: cmTrue,
			expectConfigMapData: map[string]string{
				"conflictsPresent": "False",
			},
		},
		{
			name: "conflicts appear, status changes from False to True",
			conflicts: map[cache.ObjectName][]volumecache.Conflict{
				{Namespace: "ns1", Name: "pod1"}: {
					{
						PropertyName: "SELinuxLabel",
						EventReason:  "SELinuxLabelConflict",
					},
				},
			},
			initialConflict:   metav1.ConditionFalse,
			existingConfigMap: cmFalse,
			expectConfigMapData: map[string]string{
				"conflictsPresent": "True",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ctx := ktesting.NewTestContext(t)

			var fakeClient *fake.Clientset
			if tt.existingConfigMap != nil {
				fakeClient = fake.NewClientset(tt.existingConfigMap)
			} else {
				fakeClient = fake.NewClientset()
			}

			labelCache := &fakeVolumeCache{
				conflictsToSend: tt.conflicts,
			}

			c := &SELinuxConflictsReporterController{
				kubeClient:        fakeClient,
				conflictCounter:   labelCache,
				previousConflicts: tt.initialConflict,
			}

			c.reportSELinuxConflicts(ctx)

			if tt.expectNoWrite {
				cm, err := fakeClient.CoreV1().ConfigMaps(configMapNamespace).Get(ctx, configMapName, metav1.GetOptions{})
				if tt.existingConfigMap != nil {
					// The ConfigMap should still exist unchanged.
					if err != nil {
						t.Fatalf("expected ConfigMap to exist, got error: %v", err)
					}
					if cm.Data["conflictsPresent"] != tt.existingConfigMap.Data["conflictsPresent"] {
						t.Errorf("ConfigMap data changed unexpectedly: got %v, want %v", cm.Data, tt.existingConfigMap.Data)
					}
				} else {
					if err == nil || !apierrors.IsNotFound(err) {
						t.Fatalf("expected ConfigMap to not exist, got error: %v", err)
					}
				}
				return
			}

			cm, err := fakeClient.CoreV1().ConfigMaps(configMapNamespace).Get(ctx, configMapName, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("failed to get ConfigMap: %v", err)
			}
			for key, expectedValue := range tt.expectConfigMapData {
				if cm.Data[key] != expectedValue {
					t.Errorf("ConfigMap data[%q] = %q, want %q", key, cm.Data[key], expectedValue)
				}
			}
		})
	}
}

func TestApplySELinuxConflictsConfigMap(t *testing.T) {
	cmTrue := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: configMapNamespace,
		},
		Data: map[string]string{
			"conflictsPresent": "True",
		},
	}
	cmFalse := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: configMapNamespace,
		},
		Data: map[string]string{
			"conflictsPresent": "False",
		},
	}
	tests := []struct {
		name              string
		existingConfigMap *v1.ConfigMap
		conflictsPresent  metav1.ConditionStatus
		expectData        map[string]string
	}{
		{
			name:             "creates ConfigMap when it does not exist",
			conflictsPresent: metav1.ConditionTrue,
			expectData: map[string]string{
				"conflictsPresent": "True",
			},
		},
		{
			name:              "patches ConfigMap when it already exists",
			existingConfigMap: cmFalse,
			conflictsPresent:  metav1.ConditionTrue,
			expectData: map[string]string{
				"conflictsPresent": string(metav1.ConditionTrue),
			},
		},
		{
			name:              "patches ConfigMap from True to False",
			existingConfigMap: cmTrue,
			conflictsPresent:  metav1.ConditionFalse,
			expectData: map[string]string{
				"conflictsPresent": "False",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			var fakeClient *fake.Clientset
			if tt.existingConfigMap != nil {
				fakeClient = fake.NewClientset(tt.existingConfigMap)
			} else {
				fakeClient = fake.NewClientset()
			}

			c := &SELinuxConflictsReporterController{
				kubeClient: fakeClient,
				// the rest of the struct is not used in this test
			}

			err := c.applySELinuxConflictsConfigMap(ctx, tt.conflictsPresent)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			cm, err := fakeClient.CoreV1().ConfigMaps(configMapNamespace).Get(ctx, configMapName, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("failed to get ConfigMap: %v", err)
			}
			for key, expectedValue := range tt.expectData {
				if cm.Data[key] != expectedValue {
					t.Errorf("ConfigMap data[%q] = %q, want %q", key, cm.Data[key], expectedValue)
				}
			}
		})
	}
}

func TestGetConflicts(t *testing.T) {
	tests := []struct {
		name      string
		conflicts map[cache.ObjectName][]volumecache.Conflict
		expected  metav1.ConditionStatus
	}{
		{
			name:      "no conflicts returns False",
			conflicts: nil,
			expected:  metav1.ConditionFalse,
		},
		{
			name:      "empty conflicts returns False",
			conflicts: map[cache.ObjectName][]volumecache.Conflict{},
			expected:  metav1.ConditionFalse,
		},
		{
			name: "one conflict returns True",
			conflicts: map[cache.ObjectName][]volumecache.Conflict{
				{Namespace: "ns1", Name: "pod1"}: {
					{
						PropertyName:       "SELinuxLabel",
						EventReason:        "SELinuxLabelConflict",
						Pod:                cache.ObjectName{Namespace: "ns1", Name: "pod1"},
						PropertyValue:      ":::s0:c1,c2",
						OtherPod:           cache.ObjectName{Namespace: "ns1", Name: "pod2"},
						OtherPropertyValue: ":::s0:c98,c99",
					},
				},
			},
			expected: metav1.ConditionTrue,
		},
		{
			name: "multiple conflicts returns True",
			conflicts: map[cache.ObjectName][]volumecache.Conflict{
				{Namespace: "ns1", Name: "pod1"}: {
					{PropertyName: "SELinuxLabel"},
				},
				{Namespace: "ns1", Name: "pod2"}: {
					{PropertyName: "SELinuxLabel"},
					{PropertyName: "SELinuxChangePolicy"},
				},
			},
			expected: metav1.ConditionTrue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ctx := ktesting.NewTestContext(t)
			logger := ktesting.NewLogger(t, ktesting.NewConfig())
			_ = ctx

			labelCache := &fakeVolumeCache{
				conflictsToSend: tt.conflicts,
			}

			c := &SELinuxConflictsReporterController{
				conflictCounter: labelCache,
			}

			got := c.getConflicts(logger)
			if got != tt.expected {
				t.Errorf("getConflicts() = %v, want %v", got, tt.expected)
			}
		})
	}
}
