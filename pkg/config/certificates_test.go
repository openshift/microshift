package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestCertificateConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name          string
		dropins       []string
		serving, ca   time.Duration
		forceRestart  bool
		errorContains string
	}{
		{name: "defaults", serving: 8760 * time.Hour, ca: 87600 * time.Hour, forceRestart: true},
		{name: "custom", dropins: []string{"certificates: {servingValidity: 1008h, caValidity: 17520h, forceRestartOnExpirationImminent: false}"}, serving: 1008 * time.Hour, ca: 17520 * time.Hour},
		{name: "partial", dropins: []string{"certificates: {servingValidity: 1008h}"}, serving: 1008 * time.Hour, ca: 87600 * time.Hour, forceRestart: true},
		{name: "sub-hour", dropins: []string{"certificates: {servingValidity: 90m}"}, serving: 90 * time.Minute, ca: 87600 * time.Hour, forceRestart: true},
		{name: "dropin merge", dropins: []string{"certificates: {servingValidity: 1008h, caValidity: 17520h}", "certificates: {forceRestartOnExpirationImminent: false}"}, serving: 1008 * time.Hour, ca: 17520 * time.Hour},
		{name: "dropin re-enable", dropins: []string{"certificates: {forceRestartOnExpirationImminent: false}", "certificates: {forceRestartOnExpirationImminent: true}"}, serving: 8760 * time.Hour, ca: 87600 * time.Hour, forceRestart: true},
		{name: "dropin reset", dropins: []string{"certificates: {servingValidity: 1008h}", "certificates: {servingValidity: null}"}, serving: 8760 * time.Hour, ca: 87600 * time.Hour, forceRestart: true},
		{name: "zero serving", dropins: []string{"certificates: {servingValidity: 0s}"}, errorContains: "certificates.servingValidity"},
		{name: "negative serving", dropins: []string{"certificates: {servingValidity: -1h}"}, errorContains: "certificates.servingValidity"},
		{name: "zero CA", dropins: []string{"certificates: {caValidity: 0s}"}, errorContains: "certificates.caValidity"},
		{name: "negative CA", dropins: []string{"certificates: {caValidity: -1h}"}, errorContains: "certificates.caValidity"},
		{name: "equal", dropins: []string{"certificates: {caValidity: 8760h}"}, errorContains: "must be greater than certificates.servingValidity"},
		{name: "CA shorter", dropins: []string{"certificates: {caValidity: 1008h}"}, errorContains: "must be greater than certificates.servingValidity"},
		{name: "empty", dropins: []string{`certificates: {servingValidity: ""}`}, errorContains: "certificates.servingValidity must be a Go duration string"},
		{name: "malformed", dropins: []string{"certificates: {servingValidity: six-weeks}"}, errorContains: "certificates.servingValidity must be a Go duration string"},
		{name: "unsupported unit", dropins: []string{"certificates: {caValidity: 3650d}"}, errorContains: "certificates.caValidity must be a Go duration string"},
		{name: "overflow", dropins: []string{"certificates: {caValidity: 999999999999999h}"}, errorContains: "certificates.caValidity must be a Go duration string"},
		{name: "non-string", dropins: []string{"certificates: {servingValidity: 1008}"}, errorContains: "certificates.servingValidity must be a Go duration string"},
		{name: "explicit zero overrides default", dropins: []string{"certificates: {servingValidity: 1008h}", "certificates: {servingValidity: 0s}"}, errorContains: "certificates.servingValidity"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dropins := make([][]byte, 0, len(tc.dropins))
			for _, text := range tc.dropins {
				dropins = append(dropins, []byte(text))
			}
			cfg, err := getActiveConfigFromYAMLDropins(dropins)
			if tc.errorContains != "" {
				require.ErrorContains(t, err, tc.errorContains)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.serving, cfg.Certificates.ServingDuration())
			require.Equal(t, tc.ca, cfg.Certificates.CADuration())
			require.Equal(t, tc.forceRestart, cfg.Certificates.ForceRestartEnabled())
			encoded, err := yaml.Marshal(map[string]Certificates{"certificates": cfg.Certificates})
			require.NoError(t, err)
			roundtrip, err := getActiveConfigFromYAMLDropins([][]byte{encoded})
			require.NoError(t, err)
			require.Equal(t, cfg.Certificates, roundtrip.Certificates)
		})
	}
}
