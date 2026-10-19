package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestProbeClusterIPFamily(t *testing.T) {
	t.Parallel()

	clientError := errors.New("client creation failed")
	tests := []struct {
		name       string
		clusterIP  string
		newClient  kubernetesClientFactory
		wantFamily string
		wantErr    string
		wantErrIs  error
	}{
		{
			name:       "IPv4 primary",
			clusterIP:  "172.30.0.1",
			wantFamily: "ipv4",
		},
		{
			name:       "IPv6 primary",
			clusterIP:  "fd00::1",
			wantFamily: "ipv6",
		},
		{
			name: "client creation failure",
			newClient: func(*rest.Config) (kubernetes.Interface, error) {
				return nil, clientError
			},
			wantErr:   "failed to create kubernetes client",
			wantErrIs: clientError,
		},
		{
			name:      "service lookup failure",
			newClient: func(*rest.Config) (kubernetes.Interface, error) { return fake.NewSimpleClientset(), nil },
			wantErr:   "failed to get kubernetes.default service",
		},
		{
			name:      "invalid ClusterIP",
			clusterIP: "not-an-ip",
			wantErr:   `kubernetes.default service has invalid ClusterIP "not-an-ip"`,
		},
		{
			name:    "empty ClusterIP",
			wantErr: `kubernetes.default service has invalid ClusterIP ""`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			newClient := test.newClient
			if newClient == nil {
				newClient = func(*rest.Config) (kubernetes.Interface, error) {
					return fake.NewSimpleClientset(&corev1.Service{
						ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: metav1.NamespaceDefault},
						Spec:       corev1.ServiceSpec{ClusterIP: test.clusterIP},
					}), nil
				}
			}

			family, err := probeClusterIPFamily(context.Background(), &rest.Config{}, newClient)
			if family != test.wantFamily {
				t.Fatalf("expected family %q, got %q", test.wantFamily, family)
			}
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("expected error containing %q, got %v", test.wantErr, err)
			}
			if test.wantErrIs != nil && !errors.Is(err, test.wantErrIs) {
				t.Fatalf("expected error to wrap %v, got %v", test.wantErrIs, err)
			}
		})
	}
}

func TestProbeClusterIPFamilySetsServiceLookupDeadline(t *testing.T) {
	t.Parallel()

	requestError := errors.New("request stopped")
	var requestDeadline time.Time
	config := &rest.Config{
		Host: "https://example.invalid",
		Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
			requestDeadline, _ = request.Context().Deadline()
			return nil, requestError
		}),
	}

	startedAt := time.Now()
	_, err := probeClusterIPFamily(context.Background(), config, newKubernetesClientForConfig)
	finishedAt := time.Now()
	if !errors.Is(err, requestError) {
		t.Fatalf("expected request error to be returned, got %v", err)
	}
	if requestDeadline.IsZero() {
		t.Fatal("expected service request context to have a deadline")
	}
	if earliest, latest := startedAt.Add(clusterIPFamilyProbeTimeout), finishedAt.Add(clusterIPFamilyProbeTimeout); requestDeadline.Before(earliest) || requestDeadline.After(latest) {
		t.Fatalf("expected service request deadline to be set with %s timeout, got %s", clusterIPFamilyProbeTimeout, requestDeadline.Sub(startedAt))
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
