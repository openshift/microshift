package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// Certificates configures issuance and automatic restart for managed certificates.
type Certificates struct {
	// ForceRestartOnExpirationImminent restarts MicroShift when a managed certificate
	// reaches its critical remaining-validity threshold. When false, only a warning
	// is logged. This setting takes effect at service startup; it does not disable
	// certificate regeneration during a later manual start.
	// +kubebuilder:default=true
	ForceRestartOnExpirationImminent *bool `json:"forceRestartOnExpirationImminent,omitempty"`

	// ServingValidity is the lifetime of newly issued managed serving certificates,
	// expressed as a Go duration. Existing certificates are not rewritten by a
	// configuration change. Expiration is aligned to the next midnight plus this
	// duration and capped at the signing CA's expiration.
	// +kubebuilder:default="8760h"
	// +kubebuilder:example="1008h"
	ServingValidity *metav1.Duration `json:"servingValidity,omitempty"`

	// CAValidity is the lifetime of newly issued managed CA certificates, expressed
	// as a Go duration. It must be greater than ServingValidity; both must be positive.
	// Expiration is aligned to the next midnight plus this duration. Intermediate
	// CAs cannot outlive their signing chain. Existing certificates are unchanged.
	// +kubebuilder:default="87600h"
	// +kubebuilder:example="10008h"
	CAValidity *metav1.Duration `json:"caValidity,omitempty"`
}

// UnmarshalJSON adds field context to duration parsing errors without echoing
// user-supplied values. metav1.Duration otherwise reports only the invalid value.
func (c *Certificates) UnmarshalJSON(data []byte) error {
	type plainCertificates Certificates
	decoded := struct {
		*plainCertificates
		ServingValidity json.RawMessage `json:"servingValidity"`
		CAValidity      json.RawMessage `json:"caValidity"`
	}{plainCertificates: (*plainCertificates)(c)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("invalid certificates configuration: %w", err)
	}
	for _, field := range []struct {
		name   string
		data   json.RawMessage
		target **metav1.Duration
	}{
		{"servingValidity", decoded.ServingValidity, &c.ServingValidity},
		{"caValidity", decoded.CAValidity, &c.CAValidity},
	} {
		if len(field.data) > 0 {
			if err := json.Unmarshal(field.data, field.target); err != nil {
				return fmt.Errorf("certificates.%s must be a Go duration string", field.name)
			}
		}
	}
	return nil
}

func certificateDefaults() Certificates {
	return Certificates{
		ForceRestartOnExpirationImminent: ptr.To(true),
		ServingValidity:                  &metav1.Duration{Duration: cryptomaterial.ShortLivedCertificateValidity},
		CAValidity:                       &metav1.Duration{Duration: cryptomaterial.LongLivedCertificateValidity},
	}
}

func (c Certificates) ForceRestartEnabled() bool {
	return ptr.Deref(c.ForceRestartOnExpirationImminent, true)
}

func (c Certificates) ServingDuration() time.Duration {
	return ptr.Deref(c.ServingValidity, metav1.Duration{Duration: cryptomaterial.ShortLivedCertificateValidity}).Duration
}

func (c Certificates) CADuration() time.Duration {
	return ptr.Deref(c.CAValidity, metav1.Duration{Duration: cryptomaterial.LongLivedCertificateValidity}).Duration
}

func (c Certificates) Validate() error {
	if c.ServingDuration() <= 0 {
		return fmt.Errorf("certificates.servingValidity must be greater than zero")
	}
	if c.CADuration() <= 0 {
		return fmt.Errorf("certificates.caValidity must be greater than zero")
	}
	if c.CADuration() <= c.ServingDuration() {
		return fmt.Errorf("certificates.caValidity must be greater than certificates.servingValidity")
	}
	return nil
}
