package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

func get_microshift_version_prom_metric(version_level string) prometheus.GaugeFunc {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "microshift_version",
			Help:        "MicroShift version",
			ConstLabels: prometheus.Labels{"level": version_level},
		},
		func() float64 {
			return get_microshift_version_json_num(version_level)
		},
	)
}

func get_microshift_info_prom_metric() prometheus.GaugeFunc {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "microshift_info",
			Help: "MicroShift info",
			ConstLabels: prometheus.Labels{
				"gitVersion": get_microshift_version_json_str("gitVersion"),
				"gitCommit":  get_microshift_version_json_str("gitCommit"),
				"buildDate":  get_microshift_version_json_str("buildDate"),
			},
		},
		func() float64 {
			return 1
		},
	)
}

func get_microshift_version_json_num(version_level string) float64 {
	cmd := exec.Command("microshift", "version", "-o", "json")
	json_byte, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(json_byte), &result)

	fmt.Println(result)

	number, err := strconv.ParseFloat(fmt.Sprint(result[version_level]), 64)
	if err != nil {
		panic(err)
	}
	return number
}

func get_microshift_version_json_str(version_level string) string {
	cmd := exec.Command("microshift", "version", "-o", "json")
	json_byte, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(json_byte), &result)

	return fmt.Sprint(result[version_level])
}
