package components

import (
	"context"
	"errors"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestMetricsServerServingCertReady(t *testing.T) {
	tests := []struct {
		name   string
		secret *corev1.Secret
		ready  bool
	}{
		{
			name:  "missing secret",
			ready: false,
		},
		{
			name:   "empty secret data",
			secret: &corev1.Secret{},
			ready:  false,
		},
		{
			name: "missing private key",
			secret: &corev1.Secret{Data: map[string][]byte{
				corev1.TLSCertKey: []byte("certificate"),
			}},
			ready: false,
		},
		{
			name: "serving certificate and private key",
			secret: &corev1.Secret{Data: map[string][]byte{
				corev1.TLSCertKey:       []byte("certificate"),
				corev1.TLSPrivateKeyKey: []byte("private-key"),
			}},
			ready: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := metricsServerServingCertReady(tt.secret); got != tt.ready {
				t.Errorf("metricsServerServingCertReady() = %t, want %t", got, tt.ready)
			}
		})
	}
}

func TestIsTransientKubernetesAPIError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		transient bool
	}{
		{
			name:      "service unavailable",
			err:       apierrors.NewServiceUnavailable("service unavailable"),
			transient: true,
		},
		{
			name:      "connection refused",
			err:       syscall.ECONNREFUSED,
			transient: true,
		},
		{
			name:      "unexpected server response",
			err:       apierrors.NewGenericServerResponse(502, "get", schema.GroupResource{Resource: "services"}, metricsServerServiceName, "bad gateway", 0, true),
			transient: true,
		},
		{
			name:      "unauthorized",
			err:       apierrors.NewUnauthorized("unauthorized"),
			transient: false,
		},
		{
			name:      "forbidden",
			err:       apierrors.NewForbidden(schema.GroupResource{Resource: "services"}, metricsServerServiceName, errors.New("forbidden")),
			transient: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTransientKubernetesAPIError(tt.err); got != tt.transient {
				t.Errorf("isTransientKubernetesAPIError() = %t, want %t", got, tt.transient)
			}
		})
	}
}

func TestWaitForServiceCAControllerRetriesTransientReadError(t *testing.T) {
	clientset := fake.NewSimpleClientset(newServiceCADeployment())
	var gets atomic.Int32
	clientset.PrependReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if gets.Add(1) == 1 {
			return true, nil, apierrors.NewServiceUnavailable("service unavailable")
		}
		return false, nil, nil
	})

	if err := waitForServiceCAController(context.Background(), clientset, testMetricsServerServingCertRetryBackoff()); err != nil {
		t.Fatalf("waiting for service-ca controller: %v", err)
	}
	if got := gets.Load(); got != 2 {
		t.Errorf("deployment gets = %d, want 2 after one transient error", got)
	}
}

func TestWaitForServiceCAControllerReturnsUnauthorizedError(t *testing.T) {
	clientset := fake.NewSimpleClientset(newServiceCADeployment())
	var gets atomic.Int32
	clientset.PrependReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		gets.Add(1)
		return true, nil, apierrors.NewUnauthorized("unauthorized")
	})

	err := waitForServiceCAController(context.Background(), clientset, testMetricsServerServingCertRetryBackoff())
	if !apierrors.IsUnauthorized(err) {
		t.Fatalf("waiting for service-ca controller = %v, want unauthorized error", err)
	}
	if got := gets.Load(); got != 1 {
		t.Errorf("deployment gets = %d, want 1 for terminal error", got)
	}
}

func TestWaitForMetricsServerServiceRetriesTransientReadError(t *testing.T) {
	clientset := fake.NewSimpleClientset(newMetricsServerService())
	var gets atomic.Int32
	clientset.PrependReactor("get", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if gets.Add(1) == 1 {
			return true, nil, apierrors.NewServiceUnavailable("service unavailable")
		}
		return false, nil, nil
	})

	if err := waitForMetricsServerService(context.Background(), clientset, testMetricsServerServingCertRetryBackoff()); err != nil {
		t.Fatalf("waiting for metrics-server Service: %v", err)
	}
	if got := gets.Load(); got != 2 {
		t.Errorf("service gets = %d, want 2 after one transient error", got)
	}
}

func TestWaitForMetricsServerServiceReturnsForbiddenError(t *testing.T) {
	clientset := fake.NewSimpleClientset(newMetricsServerService())
	var gets atomic.Int32
	clientset.PrependReactor("get", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		gets.Add(1)
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "services"}, metricsServerServiceName, errors.New("forbidden"))
	})

	err := waitForMetricsServerService(context.Background(), clientset, testMetricsServerServingCertRetryBackoff())
	if !apierrors.IsForbidden(err) {
		t.Fatalf("waiting for metrics-server Service = %v, want forbidden error", err)
	}
	if got := gets.Load(); got != 1 {
		t.Errorf("service gets = %d, want 1 for terminal error", got)
	}
}

func TestWaitForMetricsServerServingCertRecoversMissingSecret(t *testing.T) {
	service := newMetricsServerService()
	service.UID = "metrics-server-service-uid"
	service.Annotations[servingCertGenerationErrorAnnotation] = "service-ca was unavailable"
	service.Annotations[servingCertGenerationErrorNumAnnotation] = "10"
	service.Annotations[alphaServingCertGenerationErrorAnnotation] = "service-ca was unavailable"
	service.Annotations[alphaServingCertGenerationErrorNumAnnotation] = "10"
	clientset := fake.NewSimpleClientset(newServiceCADeployment(), service)

	var serviceUpdates atomic.Int32
	clientset.PrependReactor("update", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		serviceUpdates.Add(1)
		updatedService := action.(k8stesting.UpdateAction).GetObject().(*corev1.Service)
		if err := clientset.Tracker().Update(corev1.SchemeGroupVersion.WithResource("services"), updatedService, metricsNamespace); err != nil {
			return true, nil, err
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      metricsServerTLSResourceName,
				Namespace: metricsNamespace,
				Annotations: map[string]string{
					"service.beta.openshift.io/service-name":                     metricsServerServiceName,
					"service.beta.openshift.io/service-serving-cert-secret-name": metricsServerTLSResourceName,
				},
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				corev1.TLSCertKey:       []byte("service-ca-certificate"),
				corev1.TLSPrivateKeyKey: []byte("service-ca-private-key"),
			},
		}
		if err := clientset.Tracker().Create(corev1.SchemeGroupVersion.WithResource("secrets"), secret, metricsNamespace); err != nil {
			return true, nil, err
		}
		return true, updatedService, nil
	})

	if err := waitForMetricsServerServingCertWithOptions(context.Background(), clientset, testMetricsServerServingCertWaitOptions()); err != nil {
		t.Fatalf("waiting for recovered serving cert: %v", err)
	}
	if got := serviceUpdates.Load(); got != 1 {
		t.Errorf("service updates = %d, want 1", got)
	}

	secret, err := clientset.CoreV1().Secrets(metricsNamespace).Get(context.Background(), metricsServerTLSResourceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting recreated serving cert: %v", err)
	}
	if !metricsServerServingCertReady(secret) {
		t.Error("recreated serving cert is not ready")
	}

	updatedService, err := clientset.CoreV1().Services(metricsNamespace).Get(context.Background(), metricsServerServiceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting updated Service: %v", err)
	}
	for _, annotation := range []string{
		servingCertGenerationErrorAnnotation,
		servingCertGenerationErrorNumAnnotation,
		alphaServingCertGenerationErrorAnnotation,
		alphaServingCertGenerationErrorNumAnnotation,
	} {
		if got := updatedService.Annotations[annotation]; got != "" {
			t.Errorf("Service annotation %q = %q, want cleared", annotation, got)
		}
	}
	if got := updatedService.Annotations["service.beta.openshift.io/serving-cert-secret-name"]; got != metricsServerTLSResourceName {
		t.Errorf("serving cert annotation = %q, want %q", got, metricsServerTLSResourceName)
	}
}

func TestWaitForMetricsServerServingCertMissingService(t *testing.T) {
	clientset := fake.NewSimpleClientset(newServiceCADeployment())

	err := waitForMetricsServerServingCertWithOptions(context.Background(), clientset, testMetricsServerServingCertWaitOptions())
	if !errors.Is(err, wait.ErrWaitTimeout) {
		t.Fatalf("waiting for missing Service = %v, want wait timeout", err)
	}
	if got := countServiceUpdates(clientset.Actions()); got != 0 {
		t.Errorf("service updates = %d, want 0", got)
	}
}

func TestTriggerMetricsServerServingCertReconciliationRetriesConflict(t *testing.T) {
	clientset := fake.NewSimpleClientset(newMetricsServerService())
	var updates atomic.Int32
	clientset.PrependReactor("update", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if updates.Add(1) == 1 {
			return true, nil, apierrors.NewConflict(schema.GroupResource{Resource: "services"}, metricsServerServiceName, errors.New("conflict"))
		}
		return false, nil, nil
	})

	if err := triggerMetricsServerServingCertReconciliation(context.Background(), clientset, testMetricsServerServingCertRetryBackoff()); err != nil {
		t.Fatalf("triggering service-ca reconciliation: %v", err)
	}
	if got := updates.Load(); got != 2 {
		t.Errorf("service updates = %d, want 2 after one conflict", got)
	}
}

func TestTriggerMetricsServerServingCertReconciliationRetriesTransientReadAndUpdateErrors(t *testing.T) {
	clientset := fake.NewSimpleClientset(newMetricsServerService())
	var gets atomic.Int32
	var updates atomic.Int32
	clientset.PrependReactor("get", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if gets.Add(1) == 1 {
			return true, nil, apierrors.NewServiceUnavailable("service unavailable")
		}
		return false, nil, nil
	})
	clientset.PrependReactor("update", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if updates.Add(1) == 1 {
			return true, nil, apierrors.NewInternalError(errors.New("internal error"))
		}
		return false, nil, nil
	})

	if err := triggerMetricsServerServingCertReconciliation(context.Background(), clientset, testMetricsServerServingCertRetryBackoff()); err != nil {
		t.Fatalf("triggering service-ca reconciliation: %v", err)
	}
	if got := gets.Load(); got != 3 {
		t.Errorf("service gets = %d, want 3 after one transient get and update error", got)
	}
	if got := updates.Load(); got != 2 {
		t.Errorf("service updates = %d, want 2 after one transient update error", got)
	}
}

func TestTriggerMetricsServerServingCertReconciliationReturnsForbiddenUpdateError(t *testing.T) {
	clientset := fake.NewSimpleClientset(newMetricsServerService())
	var updates atomic.Int32
	clientset.PrependReactor("update", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		updates.Add(1)
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "services"}, metricsServerServiceName, errors.New("forbidden"))
	})

	err := triggerMetricsServerServingCertReconciliation(context.Background(), clientset, testMetricsServerServingCertRetryBackoff())
	if !apierrors.IsForbidden(err) {
		t.Fatalf("triggering service-ca reconciliation = %v, want forbidden error", err)
	}
	if got := updates.Load(); got != 1 {
		t.Errorf("service updates = %d, want 1 for terminal error", got)
	}
}

func testMetricsServerServingCertWaitOptions() metricsServerServingCertWaitOptions {
	return metricsServerServingCertWaitOptions{
		timeout:                time.Second,
		pollInterval:           time.Millisecond,
		controllerRetryBackoff: wait.Backoff{Steps: 1},
		serviceRetryBackoff:    wait.Backoff{Steps: 1},
	}
}

func testMetricsServerServingCertRetryBackoff() wait.Backoff {
	return wait.Backoff{Steps: 3, Duration: time.Millisecond, Factor: 1}
}

func newServiceCADeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceCADeploymentName,
			Namespace: serviceCANamespace,
		},
		Status: appsv1.DeploymentStatus{Conditions: []appsv1.DeploymentCondition{{
			Type:   appsv1.DeploymentAvailable,
			Status: corev1.ConditionTrue,
		}}},
	}
}

func newMetricsServerService() *corev1.Service {
	return &corev1.Service{ObjectMeta: metav1.ObjectMeta{
		Name:      metricsServerServiceName,
		Namespace: metricsNamespace,
		Annotations: map[string]string{
			"service.beta.openshift.io/serving-cert-secret-name": metricsServerTLSResourceName,
		},
	}}
}

func countServiceUpdates(actions []k8stesting.Action) int {
	updates := 0
	for _, action := range actions {
		if action.GetVerb() == "update" && action.GetResource().Resource == "services" {
			updates++
		}
	}
	return updates
}
