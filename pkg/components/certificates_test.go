package components

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestCertificateGeneration(t *testing.T) {
	clientCert := filepath.Join(t.TempDir(), "client.crt")
	writeCertificate := func(data string) {
		t.Helper()
		if err := os.WriteFile(clientCert, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	generation := func(bundle string, paths ...string) string {
		t.Helper()
		result, err := certificateGeneration([]byte(bundle), paths...)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	writeCertificate("original public client certificate")
	ovn := generation("original API CA", clientCert)
	inCluster := generation("original API CA")
	if ovn != generation("original API CA", clientCert) || inCluster != generation("original API CA") {
		t.Fatal("unchanged certificates must produce the same generation on every startup")
	}
	writeCertificate("renewed public client certificate")
	if ovn == generation("original API CA", clientCert) {
		t.Fatal("leaf renewal must restart host-mounted kubeconfig consumers")
	}
	if inCluster != generation("original API CA") {
		t.Fatal("leaf renewal must not restart service-account-only consumers")
	}
	if inCluster == generation("renewed API CA") || ovn == generation("renewed API CA", clientCert) {
		t.Fatal("CA renewal must change both generations")
	}
	if _, err := certificateGeneration(nil, clientCert+"-missing"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing certificate must fail, got %v", err)
	}
}

func TestWaitForComponentCABundle(t *testing.T) {
	for _, initial := range []string{"missing", "old", "current"} {
		t.Run(initial, func(t *testing.T) {
			client := fake.NewClientset()
			calls := 0
			client.PrependReactor("get", "configmaps", func(action clienttesting.Action) (bool, runtime.Object, error) {
				calls++
				if action.GetNamespace() != "openshift-dns" || action.(clienttesting.GetAction).GetName() != "kube-root-ca.crt" {
					t.Fatalf("unexpected resource: %v", action)
				}
				if calls == 1 && initial == "missing" {
					return true, nil, apierrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, "kube-root-ca.crt")
				}
				bundle := "current"
				if calls == 1 && initial == "old" {
					bundle = "old"
				}
				return true, &corev1.ConfigMap{Data: map[string]string{"ca.crt": bundle}}, nil
			})
			if err := waitForComponentCABundle(context.Background(), client, "openshift-dns", "current"); err != nil {
				t.Fatal(err)
			}
			if initial != "current" && calls < 2 {
				t.Fatal("must wait for the current CA before replacing pods")
			}
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := waitForComponentCABundle(ctx, fake.NewClientset(), "openshift-dns", "current")
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "openshift-dns") {
		t.Fatalf("expected bounded wait with namespace diagnostic, got %v", err)
	}
}

func TestDaemonSetRolledOut(t *testing.T) {
	ready := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Generation: 2},
		Status: appsv1.DaemonSetStatus{
			ObservedGeneration: 2, DesiredNumberScheduled: 1, UpdatedNumberScheduled: 1, NumberAvailable: 1,
		},
	}
	if !daemonSetRolledOut(ready) {
		t.Fatal("completed rollout not recognized")
	}
	for _, tc := range []struct {
		name   string
		change func(*appsv1.DaemonSet)
	}{
		{"old ready pod", func(ds *appsv1.DaemonSet) { ds.Status.UpdatedNumberScheduled = 0 }},
		{"controller has not observed update", func(ds *appsv1.DaemonSet) { ds.Status.ObservedGeneration = 1 }},
		{"new pod unavailable", func(ds *appsv1.DaemonSet) { ds.Status.NumberAvailable = 0 }},
		{"no scheduled nodes", func(ds *appsv1.DaemonSet) { ds.Status = appsv1.DaemonSetStatus{ObservedGeneration: 2} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ds := ready.DeepCopy()
			tc.change(ds)
			if daemonSetRolledOut(ds) {
				t.Fatal("incomplete rollout reported ready")
			}
		})
	}
}

func TestWaitForOVNCertificateRollout(t *testing.T) {
	client := fake.NewClientset()
	calls := 0
	client.PrependReactor("get", "daemonsets", func(action clienttesting.Action) (bool, runtime.Object, error) {
		calls++
		if action.GetNamespace() != "openshift-ovn-kubernetes" {
			t.Fatalf("unexpected namespace: %s", action.GetNamespace())
		}
		ds := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Generation: 2},
			Status: appsv1.DaemonSetStatus{
				ObservedGeneration: 2, DesiredNumberScheduled: 1, UpdatedNumberScheduled: 1, NumberAvailable: 1,
			},
		}
		if calls == 1 {
			// Simulate a restart during rollout: the old pod still reports Ready.
			ds.Status.UpdatedNumberScheduled = 0
		}
		return true, ds, nil
	})
	if err := waitForOVNCertificateRollout(context.Background(), client); err != nil {
		t.Fatal(err)
	}
	if calls < 3 {
		t.Fatal("must wait for both current DaemonSets, not stale readiness")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitForOVNCertificateRollout(ctx, fake.NewClientset()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
