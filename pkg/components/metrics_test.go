package components

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
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

func TestTriggerMetricsServerServingCertReconciliation(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricsServerServiceName,
			Namespace: metricsNamespace,
			Annotations: map[string]string{
				"service.beta.openshift.io/serving-cert-secret-name": metricsServerTLSResourceName,
			},
		},
	})

	if err := triggerMetricsServerServingCertReconciliation(context.Background(), clientset); err != nil {
		t.Fatalf("triggering service-ca reconciliation: %v", err)
	}

	service, err := clientset.CoreV1().Services(metricsNamespace).Get(context.Background(), metricsServerServiceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting service: %v", err)
	}
	if got := service.Annotations[metricsServerServingCertRecoveryAnnotation]; got == "" {
		t.Errorf("recovery annotation %q was not set", metricsServerServingCertRecoveryAnnotation)
	}
	if got := service.Annotations["service.beta.openshift.io/serving-cert-secret-name"]; got != metricsServerTLSResourceName {
		t.Errorf("serving cert annotation = %q, want %q", got, metricsServerTLSResourceName)
	}
}
