package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetricsConfiguration(t *testing.T) {
	cfg, err := getActiveConfigFromYAMLDropins(nil)
	require.NoError(t, err)
	require.Equal(t, metricsDefaults(), cfg.Metrics)
	cfg, err = getActiveConfigFromYAMLDropins([][]byte{
		[]byte("metrics: {status: Enabled}"),
		[]byte("metrics: {listenAddress: '[::1]:9091'}"),
	})
	require.NoError(t, err)
	require.Equal(t, Metrics{Status: "Enabled", ListenAddress: "[::1]:9091"}, cfg.Metrics)
	for _, address := range []string{"0.0.0.0:9090", "[::]:9090", "192.0.2.10:9090", "localhost:9090", ":9090", "127.0.0.1:0", "127.0.0.1:65536", "[::1%lo]:9090"} {
		t.Run(address, func(t *testing.T) {
			require.Error(t, (Metrics{Status: "Enabled", ListenAddress: address}).Validate())
		})
	}
	require.Error(t, (Metrics{Status: "unknown", ListenAddress: "127.0.0.1:9090"}).Validate())
}
