package cmd

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestWaitForClusterNodes(t *testing.T) {
	ready := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "primary"},
		Status: corev1.NodeStatus{
			Addresses:  []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "192.0.2.1"}},
			Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
		},
	}
	notReady := *ready.DeepCopy()
	notReady.Status.Conditions[0].Status = corev1.ConditionFalse
	unknown := *ready.DeepCopy()
	unknown.Status.Conditions = nil
	noInternalIP := *ready.DeepCopy()
	noInternalIP.Status.Addresses[0].Type = corev1.NodeExternalIP

	for _, tc := range []struct {
		name    string
		initial []corev1.Node
		retry   bool
	}{
		{name: "ready immediately", initial: []corev1.Node{ready}},
		{name: "filters unusable nodes", initial: []corev1.Node{notReady, unknown, noInternalIP, ready}},
		{name: "waits for a node", retry: true},
		{name: "waits for readiness", initial: []corev1.Node{notReady}, retry: true},
		{name: "waits for an internal IP", initial: []corev1.Node{noInternalIP}, retry: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewClientset()
			calls := 0
			client.PrependReactor("list", "nodes", func(clienttesting.Action) (bool, runtime.Object, error) {
				calls++
				nodes := tc.initial
				if tc.retry && calls > 1 {
					nodes = []corev1.Node{ready}
				}
				return true, &corev1.NodeList{Items: nodes}, nil
			})
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			members, err := waitForClusterNodes(ctx, client)
			require.NoError(t, err)
			require.Equal(t, []string{"primary=https://192.0.2.1:2380"}, members)
			if tc.retry {
				require.Equal(t, 2, calls)
			} else {
				require.Equal(t, 1, calls)
			}
		})
	}
}

func TestWaitForClusterNodesDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()

	members, err := waitForClusterNodes(ctx, fake.NewClientset())
	require.Nil(t, members)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.ErrorContains(t, err, "waiting for a ready cluster node with an internal IP")
}

func TestWaitForClusterNodesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	client := fake.NewClientset()
	client.PrependReactor("list", "nodes", func(clienttesting.Action) (bool, runtime.Object, error) {
		cancel()
		return true, &corev1.NodeList{}, nil
	})

	members, err := waitForClusterNodes(ctx, client)
	require.Nil(t, members)
	require.ErrorIs(t, err, context.Canceled)
}

func TestWaitForClusterNodesAPIError(t *testing.T) {
	resource := schema.GroupResource{Resource: "nodes"}
	for _, tc := range []struct {
		name  string
		err   error
		retry bool
	}{
		{name: "internal error", err: apierrors.NewInternalError(errors.New("temporary failure")), retry: true},
		{name: "server timeout", err: apierrors.NewServerTimeout(resource, "list", 1), retry: true},
		{name: "service unavailable", err: apierrors.NewServiceUnavailable("starting"), retry: true},
		{name: "timeout", err: apierrors.NewTimeoutError("timed out", 1), retry: true},
		{name: "too many requests", err: apierrors.NewTooManyRequests("busy", 1), retry: true},
		{name: "bad gateway", err: apierrors.NewGenericServerResponse(http.StatusBadGateway, "list", resource, "", "bad gateway", 1, true), retry: true},
		{name: "network timeout", err: &net.DNSError{Err: "timed out", IsTimeout: true}, retry: true},
		{name: "connection refused", err: syscall.ECONNREFUSED, retry: true},
		{name: "connection reset", err: syscall.ECONNRESET, retry: true},
		{name: "EOF", err: io.EOF, retry: true},
		{name: "unexpected EOF", err: io.ErrUnexpectedEOF, retry: true},
		{name: "HTTP2 connection lost", err: errors.New("http2: client connection lost"), retry: true},
		{name: "forbidden", err: apierrors.NewForbidden(resource, "", nil)},
		{name: "unauthorized", err: apierrors.NewUnauthorized("invalid credentials")},
		{name: "bad request", err: apierrors.NewBadRequest("invalid request")},
		{name: "plain forbidden", err: apierrors.NewGenericServerResponse(http.StatusForbidden, "list", resource, "", "forbidden", 0, true)},
		{name: "plain unauthorized", err: apierrors.NewGenericServerResponse(http.StatusUnauthorized, "list", resource, "", "unauthorized", 0, true)},
		{name: "plain bad request", err: apierrors.NewGenericServerResponse(http.StatusBadRequest, "list", resource, "", "bad request", 0, true)},
		{name: "plain not found", err: apierrors.NewGenericServerResponse(http.StatusNotFound, "list", resource, "", "not found", 0, true)},
		{name: "unknown error", err: errors.New("unexpected failure")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewClientset()
			calls := 0
			client.PrependReactor("list", "nodes", func(clienttesting.Action) (bool, runtime.Object, error) {
				calls++
				if calls == 1 {
					return true, nil, tc.err
				}
				return true, &corev1.NodeList{Items: []corev1.Node{{
					ObjectMeta: metav1.ObjectMeta{Name: "primary"},
					Status: corev1.NodeStatus{
						Addresses:  []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "192.0.2.1"}},
						Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
					},
				}}}, nil
			})
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			members, err := waitForClusterNodes(ctx, client)
			if tc.retry {
				require.NoError(t, err)
				require.Equal(t, []string{"primary=https://192.0.2.1:2380"}, members)
				require.Equal(t, 2, calls)
			} else {
				require.Nil(t, members)
				require.ErrorIs(t, err, tc.err)
				require.Equal(t, 1, calls, "permanent errors must not be retried")
			}
		})
	}
}

func TestWaitForClusterNodesTransientErrorDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	client := fake.NewClientset()
	client.PrependReactor("list", "nodes", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewServiceUnavailable("still starting")
	})

	members, err := waitForClusterNodes(ctx, client)
	require.Nil(t, members)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
