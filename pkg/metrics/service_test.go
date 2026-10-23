package metrics_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/admin/certificates"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestLocalMetricsService(t *testing.T) {
	for _, address := range []string{"127.0.0.1:0", "[::1]:0"} {
		t.Run(address, func(t *testing.T) {
			listener, err := net.Listen("tcp", address)
			require.NoError(t, err)
			address = listener.Addr().String()
			require.NoError(t, listener.Close())
			blocked := make(chan struct{})
			defer close(blocked)
			collector := metrics.NewCertificateCollector(func(time.Time) certificates.Observation {
				<-blocked // Neither slow collection nor API readiness gates serving.
				return certificates.Observation{}
			})
			service := metrics.NewService(config.Metrics{Status: "Enabled", ListenAddress: address}, collector)
			other := prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_other_collector", Help: "Independent collector."})
			other.Set(42)
			require.NoError(t, service.Register(other))
			require.Empty(t, service.Dependencies())
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			ready, stopped := make(chan struct{}), make(chan struct{})
			result := make(chan error, 1)
			go func() { result <- service.Run(ctx, ready, stopped) }()
			select {
			case <-ready:
			case <-time.After(5 * time.Second):
				t.Fatal("metrics listener did not start")
			}
			client := &http.Client{Timeout: 2 * time.Second, Transport: &http.Transport{Proxy: nil}}
			defer client.CloseIdleConnections()
			response, err := client.Get("http://" + address + "/metrics")
			require.NoError(t, err)
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
			require.Contains(t, string(body), "# TYPE microshift_certificate_inventory_collection_success gauge")
			require.Contains(t, string(body), "microshift_certificate_inventory_collection_success 0")
			require.Contains(t, string(body), "test_other_collector 42")
			require.NotContains(t, string(body), "go_goroutines") // No global registry.
			cancel()
			select {
			case err := <-result:
				require.NoError(t, err)
			case <-time.After(5 * time.Second):
				t.Fatal("metrics listener did not stop")
			}
		})
	}
}

func TestMetricsDisabledOrBusyPortIsNonFatal(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	for _, status := range []string{"Enabled", "Disabled"} {
		collector := metrics.NewCertificateCollector(func(time.Time) certificates.Observation {
			t.Error("collection should not start without a listener")
			return certificates.Observation{}
		})
		service := metrics.NewService(config.Metrics{Status: status, ListenAddress: listener.Addr().String()}, collector)
		ready, stopped := make(chan struct{}), make(chan struct{})
		require.NoError(t, service.Run(t.Context(), ready, stopped))
		require.Eventually(t, func() bool { return isClosed(ready) && isClosed(stopped) }, time.Second, time.Millisecond)
	}
}

func isClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
