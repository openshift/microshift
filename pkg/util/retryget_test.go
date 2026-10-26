package util

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newHealthServer serves every request with the given status, after an
// optional delay, on a listener the caller provides. The server owns the
// listener from here on and closes it in t.Cleanup.
func newHealthServer(t *testing.T, ln net.Listener, status int, delay time.Duration) *httptest.Server {
	t.Helper()
	srv := &httptest.Server{
		Listener: ln,
		Config: &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(delay)
			w.WriteHeader(status)
		})},
	}
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

func TestRetryGet_ReturnsStatusWhenListening(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := newHealthServer(t, ln, http.StatusServiceUnavailable, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	status := RetryGet(ctx, srv.URL+"/healthz", "")
	if status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", status, http.StatusServiceUnavailable)
	}
}

func TestRetryGet_RecoversWhenListenerAppearsLater(t *testing.T) {
	t.Parallel()
	// A readiness probe usually starts before the service binds its
	// listener, so its first attempts are refused. Reserve a port, release
	// it, and bring a server up on it after a delay. The port could in
	// theory be taken by another process in between; the unusual status
	// code makes sure the answer came from this server.
	const uniqueStatus = 299
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	const delay = 1500 * time.Millisecond
	start := time.Now()
	serverErr := make(chan error, 1)
	go func() {
		time.Sleep(delay)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			serverErr <- fmt.Errorf("listen on %s: %w", addr, err)
			return
		}
		newHealthServer(t, ln, uniqueStatus, 0)
		serverErr <- nil
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	status := RetryGet(ctx, fmt.Sprintf("http://%s/healthz", addr), "")
	elapsed := time.Since(start)
	if err := <-serverErr; err != nil {
		t.Fatalf("delayed server: %v", err)
	}
	if status != uniqueStatus {
		t.Fatalf("status = %d, want %d", status, uniqueStatus)
	}
	if elapsed < delay {
		t.Fatalf("answered after %s, before the listener existed at %s", elapsed, delay)
	}
}

func TestRetryGet_SlowResponseIsNotBoundedByDialTimeout(t *testing.T) {
	t.Parallel()
	// Only name resolution and the TCP connect are bounded. A listener that
	// takes longer than retryGetDialTimeout to answer must still be accepted.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	delay := retryGetDialTimeout + time.Second
	srv := newHealthServer(t, ln, http.StatusOK, delay)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	start := time.Now()
	status := RetryGet(ctx, srv.URL+"/healthz", "")
	elapsed := time.Since(start)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if elapsed < delay {
		t.Fatalf("took %s, want at least %s (the slow answer must be waited for)", elapsed, delay)
	}
}
