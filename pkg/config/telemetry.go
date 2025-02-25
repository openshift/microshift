package config

import (
	"fmt"
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
	return nil
}
