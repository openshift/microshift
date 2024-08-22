package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// create prometheus registry
	registry := prometheus.NewRegistry()

	// add runtime metrics
	registry.MustRegister(
		get_microshift_version_prom_metric("minor"),
		get_microshift_version_prom_metric("major"),
		get_microshift_version_prom_metric("patch"),
		get_microshift_info_prom_metric(),
	)

	// expose /metrics endpoint
	http.Handle(
		"/metrics", promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{}),
	)
	log.Fatalln(http.ListenAndServe(":9090", nil))
}
