package prometheus

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/blang/semver"
	appsv1 "github.com/openshift/api/apps/v1"
	appsv1listers "github.com/openshift/client-go/apps/listers/apps/v1"
	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/apimachinery/pkg/labels"
	kcorelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"

	util "github.com/openshift/library-go/pkg/apps/appsutil"
)

const (
	completeRolloutCount         = "complete_rollouts_total"
	activeRolloutDurationSeconds = "active_rollouts_duration_seconds"
	lastFailedRolloutTime        = "last_failed_rollout_time"
	strategyCount                = "strategy_total"

	availablePhase = "available"
	failedPhase    = "failed"
	cancelledPhase = "cancelled"
)

var (
	nameToQuery = func(name string) string {
		return strings.Join([]string{"openshift_apps_deploymentconfigs", name}, "_")
	}

	completeRolloutCountDesc = prometheus.NewDesc(
		nameToQuery(completeRolloutCount),
		"Counts total complete rollouts",
		[]string{"phase"}, nil,
	)

	lastFailedRolloutTimeDesc = prometheus.NewDesc(
		nameToQuery(lastFailedRolloutTime),
		"Tracks the time of last failure rollout per deployment config",
		[]string{"namespace", "name", "latest_version"}, nil,
	)

	activeRolloutDurationSecondsDesc = prometheus.NewDesc(
		nameToQuery(activeRolloutDurationSeconds),
		"Tracks the active rollout duration in seconds",
		[]string{"namespace", "name", "phase", "latest_version"}, nil,
	)

	strategyCountDesc = prometheus.NewDesc(
		nameToQuery(strategyCount),
		"Counts strategy usage",
		[]string{"type"}, nil,
	)

	apps = appsCollector{}
)

type appsCollector struct {
	rcLister   kcorelisters.ReplicationControllerLister
	dcLister   appsv1listers.DeploymentConfigLister
	nowFn      func() time.Time
	isCreated  bool
	createOnce sync.Once
	createLock sync.RWMutex
}

func InitializeMetricsCollector(dcLister appsv1listers.DeploymentConfigLister, rcLister kcorelisters.ReplicationControllerLister) {
	apps.dcLister = dcLister
	apps.rcLister = rcLister
	apps.nowFn = time.Now
	if !apps.IsCreated() {
		legacyregistry.MustRegister(&apps)
	}
	klog.V(4).Info("apps metrics registered with prometheus")
}

// Create satisfies the k8s metrics.Registerable interface. It is called when the metric is
// registered with Prometheus via k8s metrics.
func (c *appsCollector) Create(v *semver.Version) bool {
	c.createOnce.Do(func() {
		c.createLock.Lock()
		defer c.createLock.Unlock()
		c.isCreated = true
	})
	return c.IsCreated()
}

// IsCreated indicates if the metrics were created and registered with Prometheus.
func (c *appsCollector) IsCreated() bool {
	return c.isCreated
}

func (c *appsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- completeRolloutCountDesc
	ch <- activeRolloutDurationSecondsDesc
	ch <- lastFailedRolloutTimeDesc
	ch <- strategyCountDesc
}

type failedRollout struct {
	timestamp     float64
	latestVersion int64
}

func (c *appsCollector) ClearState() {
	c.createLock.Lock()
	defer c.createLock.Unlock()
	c.isCreated = false
}

func (c *appsCollector) FQName() string {
	return "openshift_apps_deploymentconfigs"
}

func (c *appsCollector) CollectDeploymentStats(ch chan<- prometheus.Metric) {
	rcList, err := c.rcLister.List(labels.Everything())
	if err != nil {
		klog.Warningf("Collecting deployment metrics for deplyomentconfigs.apps failed: %v", err)
		return
	}

	var available, failed, cancelled float64
	latestFailedRollouts := map[string]failedRollout{}
	latestCompletedRolloutVersion := map[string]int64{}
	for _, d := range rcList {
		dcName := util.DeploymentConfigNameFor(d)
		if len(dcName) == 0 {
			continue
		}
		latestVersion := util.DeploymentVersionFor(d)
		key := d.Namespace + "/" + dcName

		if util.IsTerminatedDeployment(d) {
			if util.IsDeploymentCancelled(d) {
				cancelled++
				continue
			}
			if util.IsFailedDeployment(d) {
				failed++
				// Track the latest failed rollout per deployment config
				// continue only when this is the latest version (add if below)
				if r, exists := latestFailedRollouts[key]; exists && latestVersion <= r.latestVersion {
					continue
				}
				latestFailedRollouts[key] = failedRollout{
					timestamp:     float64(d.CreationTimestamp.Unix()),
					latestVersion: latestVersion,
				}
				continue
			}
			if util.IsCompleteDeployment(d) {
				// Track the latest completed rollout per deployment config so we can prune
				// the failed ones that are older.
				v, exists := latestCompletedRolloutVersion[key]
				if !exists || latestVersion > v {
					latestCompletedRolloutVersion[key] = latestVersion
				}

				available++
				continue
			}
		}

		// TODO: Figure out under what circumstances the phase is not set.
		phase := strings.ToLower(string(util.DeploymentStatusFor(d)))
		if len(phase) == 0 {
			phase = "unknown"
		}

		// Record duration in seconds for active rollouts
		// TODO: possible time skew?
		durationSeconds := c.nowFn().Unix() - d.CreationTimestamp.Unix()
		ch <- prometheus.MustNewConstMetric(
			activeRolloutDurationSecondsDesc,
			prometheus.CounterValue,
			float64(durationSeconds),
			[]string{
				d.Namespace,
				dcName,
				phase,
				fmt.Sprintf("%d", latestVersion),
			}...)
	}

	// Record latest failed rollouts
	for key, r := range latestFailedRollouts {
		// If a completed rollout is found AFTER we recorded a failed rollout,
		// do not record the lastFailedRollout as the latest rollout is not
		// failed.
		v, exists := latestCompletedRolloutVersion[key]
		if exists && v >= r.latestVersion {
			continue
		}

		parts := strings.Split(key, "/")
		ch <- prometheus.MustNewConstMetric(
			lastFailedRolloutTimeDesc,
			prometheus.GaugeValue,
			r.timestamp,
			[]string{
				parts[0],
				parts[1],
				fmt.Sprintf("%d", r.latestVersion),
			}...)
	}

	ch <- prometheus.MustNewConstMetric(completeRolloutCountDesc, prometheus.GaugeValue, available, []string{availablePhase}...)
	ch <- prometheus.MustNewConstMetric(completeRolloutCountDesc, prometheus.GaugeValue, failed, []string{failedPhase}...)
	ch <- prometheus.MustNewConstMetric(completeRolloutCountDesc, prometheus.GaugeValue, cancelled, []string{cancelledPhase}...)
}

func (c *appsCollector) CollectDCStats(ch chan<- prometheus.Metric) {
	dcList, err := c.dcLister.List(labels.Everything())
	if err != nil {
		klog.Warningf("Collecting deployment metrics for deplyomentconfigs.apps failed: %v", err)
		return
	}

	// Init the map so we always send even 0 metrics
	strategies := map[string]float64{
		string(appsv1.DeploymentStrategyTypeRecreate): 0,
		string(appsv1.DeploymentStrategyTypeRolling):  0,
		string(appsv1.DeploymentStrategyTypeCustom):   0,
	}
	for _, dc := range dcList {
		strategies[string(dc.Spec.Strategy.Type)] += 1
	}

	strategiesKeys := make([]string, 0, len(strategies))
	for k := range strategies {
		strategiesKeys = append(strategiesKeys, k)
	}
	sort.Strings(strategiesKeys)
	for _, s := range strategiesKeys {
		v := strategies[s]
		ch <- prometheus.MustNewConstMetric(strategyCountDesc, prometheus.GaugeValue, v, []string{strings.ToLower(s)}...)
	}
}

// Collect implements the prometheus.Collector interface.
func (c *appsCollector) Collect(ch chan<- prometheus.Metric) {
	c.CollectDCStats(ch)
	c.CollectDeploymentStats(ch)
}
