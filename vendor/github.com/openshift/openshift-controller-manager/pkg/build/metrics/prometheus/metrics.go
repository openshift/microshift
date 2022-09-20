package prometheus

import (
	"sync"

	semver "github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kselector "k8s.io/apimachinery/pkg/labels"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"

	buildv1 "github.com/openshift/api/build/v1"
	buildlister "github.com/openshift/client-go/build/listers/build/v1"
)

const (
	separator        = "_"
	buildSubsystem   = "openshift_build"
	buildCount       = "total"
	buildCountQuery  = buildSubsystem + separator + buildCount
	activeBuild      = "active_time_seconds"
	activeBuildQuery = buildSubsystem + separator + activeBuild
)

var (
	buildCountDesc = prometheus.NewDesc(
		buildCountQuery,
		"Counts builds by phase, reason, and strategy",
		[]string{"phase", "reason", "strategy"},
		nil,
	)
	activeBuildDesc = prometheus.NewDesc(
		activeBuildQuery,
		"Shows the last transition time in unix epoch for running builds by namespace, name, phase, reason, and strategy",
		[]string{"namespace", "name", "phase", "reason", "strategy"},
		nil,
	)

	bc             = buildCollector{}
	cancelledPhase = string(buildv1.BuildPhaseCancelled)
	completePhase  = string(buildv1.BuildPhaseComplete)
	failedPhase    = string(buildv1.BuildPhaseFailed)
	errorPhase     = string(buildv1.BuildPhaseError)
	newPhase       = string(buildv1.BuildPhaseNew)
	pendingPhase   = string(buildv1.BuildPhasePending)
	runningPhase   = string(buildv1.BuildPhaseRunning)
)

type buildCollector struct {
	lister     buildlister.BuildLister
	isCreated  bool
	createOnce sync.Once
	createLock sync.RWMutex
}

// IntializeMetricsCollector calls into prometheus to register the buildCollector struct as a
// Collector in prometheus for the terminal and active build metrics.
func IntializeMetricsCollector(buildLister buildlister.BuildLister) {
	if !bc.IsCreated() {
		bc.lister = buildLister
		legacyregistry.MustRegister(&bc)
	}
	klog.V(4).Info("build metrics registered with prometheus")
}

// Create satisfies the k8s metrics.Registerable interface. It is called when the metric is
// registered with Prometheus via k8s metrics.
func (bc *buildCollector) Create(v *semver.Version) bool {
	bc.createOnce.Do(func() {
		bc.createLock.Lock()
		defer bc.createLock.Unlock()
		bc.isCreated = true
	})
	return bc.IsCreated()
}

// IsCreated indicates if the build metrics were created and registered with Prometheus.
func (bc *buildCollector) IsCreated() bool {
	return bc.isCreated
}

// Describe implements the prometheus.Collector interface.
func (bc *buildCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- buildCountDesc
	ch <- activeBuildDesc
}

type collectKey struct {
	phase    string
	reason   string
	strategy string
}

// Collect implements the prometheus.Collector interface.
func (bc *buildCollector) Collect(ch chan<- prometheus.Metric) {
	result, err := bc.lister.List(kselector.Everything())

	if err != nil {
		klog.V(4).Infof("Collect err %v", err)
		return
	}

	// collectBuild will return counts for the build's phase/reason tuple,
	// and counts for these tuples be added to the total amount posted to prometheus
	counts := map[collectKey]int{}
	for _, b := range result {
		k := bc.collectBuild(ch, b)
		counts[k] = counts[k] + 1
	}

	for key, count := range counts {
		addCountGauge(ch, buildCountDesc, key.phase, key.reason, key.strategy, float64(count))
	}
}

func (bc *buildCollector) ClearState() {
	bc.createLock.Lock()
	defer bc.createLock.Unlock()
	bc.isCreated = false
}

func (bc *buildCollector) FQName() string {
	return buildSubsystem
}

func addCountGauge(ch chan<- prometheus.Metric, desc *prometheus.Desc, phase, reason, strategy string, v float64) {
	lv := []string{phase, reason, strategy}
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, lv...)
}

func addTimeGauge(ch chan<- prometheus.Metric, b *buildv1.Build, time *metav1.Time, desc *prometheus.Desc, phase string, reason string, strategy string) {
	if time != nil {
		lv := []string{b.ObjectMeta.Namespace, b.ObjectMeta.Name, phase, reason, strategy}
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(time.Unix()), lv...)
	}
}

func (bc *buildCollector) collectBuild(ch chan<- prometheus.Metric, b *buildv1.Build) (key collectKey) {

	r := string(b.Status.Reason)
	s := strategyType(b.Spec.Strategy)
	key = collectKey{reason: r, strategy: s}
	switch b.Status.Phase {
	// remember, new and pending builds don't have a start time
	case buildv1.BuildPhaseNew:
		key.phase = newPhase
		addTimeGauge(ch, b, &b.CreationTimestamp, activeBuildDesc, newPhase, r, s)
	case buildv1.BuildPhasePending:
		key.phase = pendingPhase
		addTimeGauge(ch, b, &b.CreationTimestamp, activeBuildDesc, pendingPhase, r, s)
	case buildv1.BuildPhaseRunning:
		key.phase = runningPhase
		addTimeGauge(ch, b, b.Status.StartTimestamp, activeBuildDesc, runningPhase, r, s)
	case buildv1.BuildPhaseFailed:
		key.phase = failedPhase
	case buildv1.BuildPhaseError:
		key.phase = errorPhase
	case buildv1.BuildPhaseCancelled:
		key.phase = cancelledPhase
	case buildv1.BuildPhaseComplete:
		key.phase = completePhase
	}
	return key
}

func strategyType(strategy buildv1.BuildStrategy) string {
	switch {
	case strategy.DockerStrategy != nil:
		return "Docker"
	case strategy.CustomStrategy != nil:
		return "Custom"
	case strategy.SourceStrategy != nil:
		return "Source"
	case strategy.JenkinsPipelineStrategy != nil:
		return "JenkinsPipeline"
	}
	return ""
}
