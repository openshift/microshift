package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/openshift/microshift/pkg/admin/certificates"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
	"github.com/prometheus/client_golang/prometheus"
)

const certificateRefreshInterval = time.Minute

// CertificateCollector publishes immutable snapshots. Collect never accesses PKI,
// so a slow disk or renewal operation cannot block HTTP requests.
type CertificateCollector struct {
	observe     func(time.Time) certificates.Observation
	mu          sync.RWMutex
	snapshot    []prometheus.Metric
	lastSuccess float64

	notBefore, notAfter, status, pending, success, lastSuccessTime *prometheus.Desc
}

func NewCertificateCollector(observe func(time.Time) certificates.Observation) *CertificateCollector {
	labels := []string{"name", "service", "role", "rotation_policy"}
	c := &CertificateCollector{
		observe:         observe,
		notBefore:       prometheus.NewDesc("microshift_certificate_not_before_timestamp_seconds", "Active on-disk certificate validity start in Unix seconds.", labels, nil),
		notAfter:        prometheus.NewDesc("microshift_certificate_not_after_timestamp_seconds", "Active on-disk certificate expiry in Unix seconds.", labels, nil),
		status:          prometheus.NewDesc("microshift_certificate_status", "Active on-disk certificate status; exactly one state is 1 per certificate.", append(labels, "status"), nil),
		pending:         prometheus.NewDesc("microshift_certificate_renewal_pending", "Whether renewal metadata records a pending generation; not a guarantee of activation eligibility.", nil, nil),
		success:         prometheus.NewDesc("microshift_certificate_inventory_collection_success", "Whether the latest active inventory and pending metadata observation both succeeded.", nil, nil),
		lastSuccessTime: prometheus.NewDesc("microshift_certificate_inventory_last_success_timestamp_seconds", "Unix time of the last complete successful observation, or 0 if none.", nil, nil),
	}
	c.snapshot = []prometheus.Metric{
		prometheus.MustNewConstMetric(c.success, prometheus.GaugeValue, 0),
		prometheus.MustNewConstMetric(c.lastSuccessTime, prometheus.GaugeValue, 0),
	}
	return c
}

func (c *CertificateCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range []*prometheus.Desc{c.notBefore, c.notAfter, c.status, c.pending, c.success, c.lastSuccessTime} {
		ch <- desc
	}
}

func (c *CertificateCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	snapshot := c.snapshot
	c.mu.RUnlock()
	for _, metric := range snapshot {
		ch <- metric
	}
}

func (c *CertificateCollector) Run(ctx context.Context) {
	ticker := time.NewTicker(certificateRefreshInterval)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		c.Refresh(time.Now())
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// Refresh replaces the snapshot, dropping unavailable or obsolete series rather
// than presenting old observations as current health. Only last-success is kept.
func (c *CertificateCollector) Refresh(now time.Time) {
	observation := c.observe(now)
	metrics, inventoryOK := c.inventoryMetrics(observation)
	success := float64(0)
	if observation.PendingError == nil {
		pending := float64(0)
		if observation.RenewalPending {
			pending = 1
		}
		metrics = append(metrics, prometheus.MustNewConstMetric(c.pending, prometheus.GaugeValue, pending))
		if inventoryOK {
			success = 1
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if success == 1 {
		c.lastSuccess = float64(observation.ObservedAt.Unix())
	}
	metrics = append(metrics,
		prometheus.MustNewConstMetric(c.success, prometheus.GaugeValue, success),
		prometheus.MustNewConstMetric(c.lastSuccessTime, prometheus.GaugeValue, c.lastSuccess))
	c.snapshot = metrics
}

func (c *CertificateCollector) inventoryMetrics(observation certificates.Observation) ([]prometheus.Metric, bool) {
	if observation.InventoryError != nil || len(observation.Inventory) == 0 {
		return nil, false
	}
	metrics := make([]prometheus.Metric, 0, len(observation.Inventory)*7+3)
	for _, entry := range observation.Inventory {
		status, err := entry.StatusAt(observation.ObservedAt)
		if err != nil {
			return nil, false
		}
		labels := []string{entry.Name, entry.Service, string(entry.Role), string(entry.RotationPolicy)}
		metrics = append(metrics,
			prometheus.MustNewConstMetric(c.notBefore, prometheus.GaugeValue, float64(entry.Certificate.NotBefore.Unix()), labels...),
			prometheus.MustNewConstMetric(c.notAfter, prometheus.GaugeValue, float64(entry.Certificate.NotAfter.Unix()), labels...))
		for _, state := range []certchains.CertificateStatus{
			certchains.CertificateStatusHealthy, certchains.CertificateStatusExpiresSoon,
			certchains.CertificateStatusExpirationImminent, certchains.CertificateStatusExpired,
			certchains.CertificateStatusNotYetValid,
		} {
			value := float64(0)
			if status == state {
				value = 1
			}
			metrics = append(metrics, prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, value, append(labels, string(state))...))
		}
	}
	return metrics, true
}
