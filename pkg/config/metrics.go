package config

import (
	"fmt"
	"net/netip"
)

const (
	MetricsEnabled  = "Enabled"
	MetricsDisabled = "Disabled"
)

// Metrics configures the MicroShift-owned local metrics endpoint, independently
// of telemetry uploads and the optional OpenTelemetry Collector package.
type Metrics struct {
	// Enable the local metrics endpoint. Changes require a MicroShift restart.
	// +kubebuilder:validation:Enum=Enabled;Disabled
	// +kubebuilder:default=Disabled
	Status string `json:"status"`

	// IP address and port for HTTP /metrics. Only literal loopback addresses are
	// allowed (for example, 127.0.0.1:9090 or [::1]:9090). Local access is unauthenticated.
	// +kubebuilder:default="127.0.0.1:9090"
	ListenAddress string `json:"listenAddress"`
}

func metricsDefaults() Metrics {
	return Metrics{Status: MetricsDisabled, ListenAddress: "127.0.0.1:9090"}
}

func (m Metrics) Validate() error {
	if m.Status != MetricsEnabled && m.Status != MetricsDisabled {
		return fmt.Errorf("metrics.status must be Enabled or Disabled")
	}
	address, err := netip.ParseAddrPort(m.ListenAddress)
	if err != nil {
		return fmt.Errorf("metrics.listenAddress must be a literal loopback IP address and port between 1 and 65535")
	}
	if !address.Addr().IsLoopback() || address.Addr().Zone() != "" {
		return fmt.Errorf("metrics.listenAddress must use a literal loopback IP address")
	}
	if address.Port() == 0 {
		return fmt.Errorf("metrics.listenAddress port must be between 1 and 65535")
	}
	return nil
}
