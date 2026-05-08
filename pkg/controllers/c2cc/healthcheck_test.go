package c2cc

import (
	"context"
	"net"
	"testing"
	"time"

	microshiftv1alpha1 "github.com/openshift/microshift/pkg/apis/microshift/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	fakeclientset "github.com/openshift/microshift/pkg/generated/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestCrNameForRemote(t *testing.T) {
	tests := []struct {
		name     string
		nextHop  string
		expected string
	}{
		{"IPv4", "10.100.0.2", "c2cc-10-100-0-2"},
		{"IPv6", "fd00::2", "c2cc-fd00--2"},
		{"IPv6 full", "2001:db8::1", "c2cc-2001-db8--1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.nextHop)
			require.NotNil(t, ip)
			assert.Equal(t, tt.expected, crNameForRemote(ip))
		})
	}
}

func TestBuildDesiredCRs(t *testing.T) {
	cfg := &config.Config{
		C2CC: config.C2CC{
			ResolvedProbeInterval: 15 * time.Second,
			Resolved: []config.ResolvedRemoteCluster{
				{
					NextHop: net.ParseIP("10.100.0.2"),
					ProbeIP: "10.46.0.11",
				},
				{
					NextHop: net.ParseIP("10.100.0.3"),
					ProbeIP: "10.47.0.11",
				},
			},
		},
	}

	mgr := newHealthcheckCRManager(nil, cfg)
	desired := mgr.buildDesiredCRs()

	assert.Len(t, desired, 2)

	cr1, ok := desired["c2cc-10-100-0-2"]
	require.True(t, ok, "expected CR for 10.100.0.2")
	assert.Equal(t, "10.46.0.11:8080", cr1.Spec.ProbeTarget)
	assert.Equal(t, 15*time.Second, cr1.Spec.ProbeInterval.Duration)
	assert.Equal(t, managerName, cr1.Labels[managedByLabel])

	cr2, ok := desired["c2cc-10-100-0-3"]
	require.True(t, ok, "expected CR for 10.100.0.3")
	assert.Equal(t, "10.47.0.11:8080", cr2.Spec.ProbeTarget)
	assert.Equal(t, 15*time.Second, cr2.Spec.ProbeInterval.Duration)
}

func newFakeClientset(objects ...runtime.Object) *fakeclientset.Clientset {
	return fakeclientset.NewSimpleClientset(objects...)
}

func TestReconcileCreatesNewCRs(t *testing.T) {
	cs := newFakeClientset()
	cfg := &config.Config{
		C2CC: config.C2CC{
			ResolvedProbeInterval: 10 * time.Second,
			Resolved: []config.ResolvedRemoteCluster{
				{
					NextHop: net.ParseIP("10.100.0.2"),
					ProbeIP: "10.46.0.11",
				},
			},
		},
	}

	mgr := newHealthcheckCRManager(cs.MicroshiftV1alpha1(), cfg)
	err := mgr.reconcile(context.Background())
	require.NoError(t, err)

	var creates int
	for _, a := range cs.Actions() {
		if a.GetVerb() == "create" {
			creates++
			cr := a.(ktesting.CreateAction).GetObject().(*microshiftv1alpha1.RemoteCluster)
			assert.Equal(t, "c2cc-10-100-0-2", cr.Name)
			assert.Equal(t, "10.46.0.11:8080", cr.Spec.ProbeTarget)
		}
	}
	assert.Equal(t, 1, creates)
}

func TestReconcileDeletesStaleCRs(t *testing.T) {
	staleCR := &microshiftv1alpha1.RemoteCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "c2cc-10-100-0-99",
			Labels: map[string]string{managedByLabel: managerName},
		},
		Spec: microshiftv1alpha1.RemoteClusterSpec{
			ProbeTarget:   "10.99.0.11:8080",
			ProbeInterval: metav1.Duration{Duration: 10 * time.Second},
		},
	}
	cs := newFakeClientset(staleCR)

	cfg := &config.Config{
		C2CC: config.C2CC{
			ResolvedProbeInterval: 10 * time.Second,
			Resolved:              []config.ResolvedRemoteCluster{},
		},
	}

	mgr := newHealthcheckCRManager(cs.MicroshiftV1alpha1(), cfg)
	err := mgr.reconcile(context.Background())
	require.NoError(t, err)

	var deletes int
	for _, a := range cs.Actions() {
		if a.GetVerb() == "delete" {
			deletes++
		}
	}
	assert.Equal(t, 1, deletes)
}

func TestReconcileUpdatesCR(t *testing.T) {
	existingCR := &microshiftv1alpha1.RemoteCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "c2cc-10-100-0-2",
			Labels: map[string]string{managedByLabel: managerName},
		},
		Spec: microshiftv1alpha1.RemoteClusterSpec{
			ProbeTarget:   "10.46.0.11:8080",
			ProbeInterval: metav1.Duration{Duration: 30 * time.Second},
		},
	}
	cs := newFakeClientset(existingCR)

	cfg := &config.Config{
		C2CC: config.C2CC{
			ResolvedProbeInterval: 15 * time.Second,
			Resolved: []config.ResolvedRemoteCluster{
				{
					NextHop: net.ParseIP("10.100.0.2"),
					ProbeIP: "10.46.0.11",
				},
			},
		},
	}

	mgr := newHealthcheckCRManager(cs.MicroshiftV1alpha1(), cfg)
	err := mgr.reconcile(context.Background())
	require.NoError(t, err)

	var updates int
	for _, a := range cs.Actions() {
		if a.GetVerb() == "update" {
			updates++
			cr := a.(ktesting.UpdateAction).GetObject().(*microshiftv1alpha1.RemoteCluster)
			assert.Equal(t, 15*time.Second, cr.Spec.ProbeInterval.Duration)
		}
	}
	assert.Equal(t, 1, updates)
}
