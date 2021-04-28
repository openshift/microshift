package importer

import (
	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

var (
	v1ImageImportsCounter = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Subsystem: "apiserver",
			Name:      "v1_image_imports_total",
			Help:      "Counter of images imported using v1 protocol",
		},
		[]string{"repository"},
	)
)

func init() {
	legacyregistry.MustRegister(v1ImageImportsCounter)
}
