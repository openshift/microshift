package metrics

import (
	"sync"

	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

var (
	unidleCount = metrics.NewCounter(&metrics.CounterOpts{
		Namespace: "openshift",
		Subsystem: "unidle",
		Name:      "events_total",
		Help:      "Total count of unidling events observed by the unidling controller",
	})
	registerOnce sync.Once
)

func GetEventsTotalCounter() *metrics.Counter {
	registerOnce.Do(func() {
		legacyregistry.MustRegister(unidleCount)
	})
	return unidleCount
}
