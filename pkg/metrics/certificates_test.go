package metrics_test

import (
	"crypto/x509"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/admin/certificates"
	"github.com/openshift/microshift/pkg/metrics"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func gather(t *testing.T, registry *prometheus.Registry) map[string]*dto.MetricFamily {
	t.Helper()
	families, err := registry.Gather()
	require.NoError(t, err)
	result := make(map[string]*dto.MetricFamily, len(families))
	for _, family := range families {
		require.Equal(t, dto.MetricType_GAUGE, family.GetType())
		result[family.GetName()] = family
	}
	return result
}

func TestCertificateSnapshots(t *testing.T) {
	start := time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	entry := certchains.CertificateInventoryEntry{
		Name: "test-serving", Service: "test", Role: certchains.CertificateRoleServing,
	}
	observation := certificates.Observation{}
	collector := metrics.NewCertificateCollector(func(now time.Time) certificates.Observation {
		observation.ObservedAt = now
		return observation
	})
	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(collector))
	for _, policy := range []certchains.RotationPolicy{certchains.RotationPolicyStandard, certchains.RotationPolicyExtended} {
		for _, lifetime := range []time.Duration{time.Hour, 365 * 24 * time.Hour} {
			entry.RotationPolicy = policy
			entry.Certificate = x509.Certificate{NotBefore: start, NotAfter: start.Add(lifetime)}
			observation.Inventory = certchains.CertificateInventory{entry}
			warning, critical, err := policy.StatusThresholds()
			require.NoError(t, err)
			for _, tc := range []struct {
				now  time.Time
				want certchains.CertificateStatus
			}{
				{start.Add(-time.Second), certchains.CertificateStatusNotYetValid},
				{start, certchains.CertificateStatusHealthy},
				{start.Add(time.Duration(float64(lifetime)*(1-warning)) + time.Second), certchains.CertificateStatusExpiresSoon},
				{start.Add(time.Duration(float64(lifetime)*(1-critical)) + time.Second), certchains.CertificateStatusExpirationImminent},
				{start.Add(lifetime), certchains.CertificateStatusExpired},
				{start, certchains.CertificateStatusHealthy}, // Clock moved backward; files did not change.
			} {
				collector.Refresh(tc.now)
				families := gather(t, registry)
				require.Len(t, families, 6)
				require.Equal(t, float64(entry.Certificate.NotAfter.Unix()), families["microshift_certificate_not_after_timestamp_seconds"].Metric[0].GetGauge().GetValue())
				require.Equal(t, float64(entry.Certificate.NotBefore.Unix()), families["microshift_certificate_not_before_timestamp_seconds"].Metric[0].GetGauge().GetValue())
				state, err := entry.StatusAt(tc.now)
				require.NoError(t, err)
				require.Equal(t, tc.want, state)
				statuses := families["microshift_certificate_status"].Metric
				require.Len(t, statuses, 5)
				total := float64(0)
				for _, metric := range statuses {
					labels := map[string]string{}
					for _, label := range metric.Label {
						labels[label.GetName()] = label.GetValue()
					}
					require.Len(t, labels, 5)
					require.Equal(t, entry.Name, labels["name"])
					require.Equal(t, entry.Service, labels["service"])
					require.Equal(t, string(entry.Role), labels["role"])
					require.Equal(t, string(policy), labels["rotation_policy"])
					require.Equal(t, labels["status"] == string(state), metric.GetGauge().GetValue() == 1)
					total += metric.GetGauge().GetValue()
				}
				require.Equal(t, float64(1), total)
			}
		}
	}
	observation.RenewalPending = true
	collector.Refresh(start)
	families := gather(t, registry)
	require.Equal(t, float64(1), families["microshift_certificate_renewal_pending"].Metric[0].GetGauge().GetValue())
	require.Equal(t, float64(entry.Certificate.NotAfter.Unix()), families["microshift_certificate_not_after_timestamp_seconds"].Metric[0].GetGauge().GetValue())

	observation.PendingError = errors.New("invalid metadata")
	collector.Refresh(start.Add(time.Minute))
	families = gather(t, registry)
	require.NotContains(t, families, "microshift_certificate_renewal_pending")
	require.Contains(t, families, "microshift_certificate_status")
	require.Zero(t, families["microshift_certificate_inventory_collection_success"].Metric[0].GetGauge().GetValue())
	require.Equal(t, float64(start.Unix()), families["microshift_certificate_inventory_last_success_timestamp_seconds"].Metric[0].GetGauge().GetValue())
	observation.InventoryError = errors.New("busy")
	collector.Refresh(start.Add(2 * time.Minute))
	require.Len(t, gather(t, registry), 2)

	observation.PendingError, observation.InventoryError = nil, nil
	observation.RenewalPending = false
	observation.Inventory[0].Name = "replacement"
	collector.Refresh(start.Add(3 * time.Minute))
	for _, metric := range gather(t, registry)["microshift_certificate_status"].Metric {
		for _, label := range metric.Label {
			if label.GetName() == "name" {
				require.Equal(t, "replacement", label.GetValue())
			}
		}
	}
	// Scrapes and periodic refresh safely share an immutable snapshot.
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 20 {
			collector.Refresh(start)
		}
	})
	for range 20 {
		gather(t, registry)
	}
	wg.Wait()
}
