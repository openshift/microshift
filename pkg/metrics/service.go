// Package metrics serves MicroShift-owned metrics independently of kube-apiserver.
package metrics

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
)

type Service struct {
	config       config.Metrics
	registry     *prometheus.Registry
	certificates *CertificateCollector
}

func NewService(cfg config.Metrics, certificates *CertificateCollector) *Service {
	registry := prometheus.NewRegistry()
	registry.MustRegister(certificates)
	return &Service{config: cfg, registry: registry, certificates: certificates}
}

// Register adds future MicroShift collectors to this isolated registry. Collect
// implementations must serve cached data, not perform blocking I/O.
func (s *Service) Register(collector prometheus.Collector) error {
	return s.registry.Register(collector)
}

func (s *Service) Name() string           { return "microshift-metrics" }
func (s *Service) Dependencies() []string { return nil }

// Run is deliberately non-fatal: losing observability must not stop MicroShift.
// No other service depends on this readiness signal.
func (s *Service) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	if s.config.Status != "Enabled" {
		close(ready)
		return nil
	}
	if err := s.config.Validate(); err != nil {
		klog.ErrorS(err, "Local metrics endpoint disabled")
		close(ready)
		return nil
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.config.ListenAddress)
	if err != nil {
		klog.ErrorS(err, "Local metrics endpoint unavailable; MicroShift will continue without it")
		close(ready)
		return nil
	}
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{
		ErrorHandling:       promhttp.ContinueOnError,
		MaxRequestsInFlight: 2,
		Timeout:             5 * time.Second,
	}))
	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 * 1024,
	}
	defer func() { _ = server.Close() }()
	collectionCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go s.certificates.Run(collectionCtx)
	serveError := make(chan error, 1)
	go func() { serveError <- server.Serve(listener) }()
	close(ready)
	klog.InfoS("Local metrics endpoint listening", "address", listener.Addr())
	select {
	case <-ctx.Done():
	case err := <-serveError:
		if !errors.Is(err, http.ErrServerClosed) {
			klog.ErrorS(err, "Local metrics endpoint stopped; MicroShift will continue without it")
		}
	}
	return nil
}
