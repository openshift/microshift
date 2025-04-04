package config

import (
	"fmt"
	"net/url"
)

const (
	StatusEnabled   TelemetryStatusEnum = "Enabled"
	StatusDisabled  TelemetryStatusEnum = "Disabled"
	defaultEndpoint                     = "https://infogw.api.openshift.com"
)

type TelemetryStatusEnum string

type Telemetry struct {
	// Telemetry status, which can be Enabled or Disabled. Defaults to Enabled.
	// +kubebuilder:default=Enabled
	Status TelemetryStatusEnum `json:"status"`

	// Endpoint where to send telemetry data.
	// +kubebuilder:default="https://infogw.api.openshift.com"
	Endpoint string `json:"endpoint"`

	// HTTP proxy to use exclusively for telemetry data. If unset telemetry will
	// default to use the system configured proxy.
	Proxy string `json:"proxy"`
}

func telemetryDefaults() Telemetry {
	return Telemetry{
		Status:   StatusEnabled,
		Endpoint: defaultEndpoint,
	}
}

func (t *Telemetry) validate() error {
	if t.Status != StatusEnabled && t.Status != StatusDisabled {
		return fmt.Errorf("invalid telemetry status: %s", t.Status)
	}
	if t.Proxy != "" {
		if _, err := url.Parse(t.Proxy); err != nil {
			return fmt.Errorf("invalid telemetry proxy URL: %s", t.Proxy)
		}
	}
	return nil
}
