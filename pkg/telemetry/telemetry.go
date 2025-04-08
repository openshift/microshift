/*
Copyright © 2025 MicroShift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package telemetry

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	proto "github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/openshift/microshift/pkg/admin/prerun"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/prometheus/prompb"
	"k8s.io/klog/v2"
)

const (
	authString = `{"authorization_token": "%s", "cluster_id": "%s"}`

	MetricNameCPUCapacity       = "cluster:capacity_cpu_cores:sum"
	MetricNameMemoryCapacity    = "cluster:capacity_memory_bytes:sum"
	MetricNameCPUUsage          = "cluster:cpu_usage_cores:sum"
	MetricNameMemoryUsage       = "cluster:memory_usage_bytes:sum"
	MetricNameResourceUsage     = "cluster:usage:resources:sum"
	MetricNameContainersUsage   = "cluster:usage:containers:sum"
	MetricNameMicroShiftVersion = "microshift_version"

	LabelNameID             = "_id"
	LabelNameArch           = "label_kubernetes_io_arch"
	LabelNameOS             = "label_node_openshift_io_os_id"
	LabelNameInstanceType   = "label_beta_kubernetes_io_instance_type"
	LabelNameResource       = "resource"
	LabelNameVersion        = "version"
	LabelNameDeploymentType = "deployment_type"
	LabelNameOSVersion      = "os_version_id"
)

type Metric struct {
	Name      string
	Labels    []MetricLabel
	Timestamp int64
	Value     float64
}

type MetricLabel struct {
	Name  string
	Value string
}

type TelemetryClient struct {
	endpoint  string
	clusterId string
	// next two attributes are required to compute the cpu usage based
	// on the cpu seconds we get from kubelet.
	previousCPUSeconds   float64
	previousCPUtimestamp int64
	// For proxy configuration.
	transport http.RoundTripper
}

func NewTelemetryClient(endpoint, clusterId, proxy string) *TelemetryClient {
	transport := http.DefaultTransport
	if proxy != "" {
		// Proxy was validated before reaching this point, ignore the error because it cant happen.
		u, _ := url.Parse(proxy)
		transport = &http.Transport{
			Proxy: http.ProxyURL(u),
		}
	}
	return &TelemetryClient{
		endpoint:             endpoint,
		clusterId:            clusterId,
		previousCPUSeconds:   0,
		previousCPUtimestamp: 0,
		transport:            transport,
	}
}

func (t *TelemetryClient) encodeAuth(pullSecret string) string {
	authString := fmt.Sprintf(authString, pullSecret, t.clusterId)
	return base64.StdEncoding.EncodeToString([]byte(authString))
}

func (t *TelemetryClient) Send(ctx context.Context, pullSecret string, metrics []Metric) error {
	wr := convertMetricsToWriteRequest(metrics)
	data, err := proto.Marshal(wr)
	if err != nil {
		return fmt.Errorf("failed to marshal WriteRequest: %v", err)
	}
	compressed := snappy.Encode(nil, data)
	reader := bytes.NewReader(compressed)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint, reader)
	if err != nil {
		return fmt.Errorf("unable to create request: %v", err)
	}

	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.encodeAuth(pullSecret)))

	client := &http.Client{
		Transport: t.transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do the request: %v", err)
	}
	defer func() {
		// Discard the body to close the TCP socket right away instead of waiting for
		// the timeout in TIME_WAIT.
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			klog.Error(err, "error discarding body")
		}
		resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusOK {
		klog.Infof("Metrics sent successfully")
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: %v", err)
	}
	return fmt.Errorf("request unsuccessful. Status code: %v. Body: %v", resp.StatusCode, string(body))
}

func (t *TelemetryClient) Collect(ctx context.Context, cfg *config.Config) ([]Metric, error) {
	kubeletMetrics, err := fetchKubeletMetrics(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch kubelet metrics: %v", err)
	}
	nodeLabels, err := fetchNodeLabels(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch node labels: %v", err)
	}
	kubeResources, err := fetchKubernetesResources(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch kubernetes resources: %v", err)
	}

	metrics := make([]Metric, 0)
	capacityMetrics, err := computeCapacityMetrics(kubeletMetrics, nodeLabels)
	if err != nil {
		return nil, fmt.Errorf("failed to compute capacity metrics: %v", err)
	}
	metrics = append(metrics, capacityMetrics...)

	usageMetrics, currentCPUSeconds, err := computeUsageMetrics(kubeletMetrics, kubeResources, t.previousCPUSeconds, t.previousCPUtimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to compute usage metrics: %v", err)
	}
	t.previousCPUSeconds = currentCPUSeconds
	t.previousCPUtimestamp = time.Now().UnixNano() / time.Millisecond.Nanoseconds()
	metrics = append(metrics, usageMetrics...)

	versionMetric, err := computeMicroShiftVersionMetric()
	if err != nil {
		return nil, fmt.Errorf("failed to compute microshift version metric: %v", err)
	}
	metrics = append(metrics, versionMetric)

	for i := range metrics {
		metrics[i].Labels = append(metrics[i].Labels, MetricLabel{Name: LabelNameID, Value: t.clusterId})
	}

	return metrics, nil
}

func computeCapacityMetrics(kubeletMetrics map[string]*io_prometheus_client.MetricFamily, nodeLabels map[string]string) ([]Metric, error) {
	archValue, ok := nodeLabels["kubernetes.io/arch"]
	if !ok {
		return nil, fmt.Errorf("node label kubernetes.io/arch not found")
	}
	osIdValue, ok := nodeLabels["node.openshift.io/os_id"]
	if !ok {
		return nil, fmt.Errorf("node label node.openshift.io/os_id not found")
	}
	instanceTypeValue, ok := nodeLabels["node.kubernetes.io/instance-type"]
	if !ok {
		return nil, fmt.Errorf("node label node.kubernetes.io/instance-type not found")
	}
	kubeletCPUCapacity, ok := kubeletMetrics["machine_cpu_cores"]
	if !ok {
		return nil, fmt.Errorf("metric machine_cpu_cores not found")
	}
	kubeletMemoryCapacity, ok := kubeletMetrics["machine_memory_bytes"]
	if !ok {
		return nil, fmt.Errorf("metric machine_memory_bytes not found")
	}
	currentTimestamp := time.Now().UnixNano() / time.Millisecond.Nanoseconds()
	return []Metric{
		{
			Name: MetricNameCPUCapacity,
			Labels: []MetricLabel{
				{Name: LabelNameOS, Value: osIdValue},
				{Name: LabelNameInstanceType, Value: instanceTypeValue},
				{Name: LabelNameArch, Value: archValue},
			},
			Timestamp: currentTimestamp,
			Value:     aggregateMetricValues(kubeletCPUCapacity.Metric),
		},
		{
			Name: MetricNameMemoryCapacity,
			Labels: []MetricLabel{
				{Name: LabelNameOS, Value: osIdValue},
				{Name: LabelNameInstanceType, Value: instanceTypeValue},
				{Name: LabelNameArch, Value: archValue},
			},
			Timestamp: currentTimestamp,
			Value:     aggregateMetricValues(kubeletMemoryCapacity.Metric),
		},
	}, nil
}

func computeUsageMetrics(kubeletMetrics map[string]*io_prometheus_client.MetricFamily, kubeResources map[string]int, previousCPUSeconds float64, previousTimestamp int64) ([]Metric, float64, error) {
	currentTimestamp := time.Now().UnixNano() / time.Millisecond.Nanoseconds()

	resourceTypes := []string{"pods", "namespaces", "services", "ingresses.networking.k8s.io", "routes.route.openshift.io", "customresourcedefinitions.apiextensions.k8s.io"}
	resourceMetrics := make([]Metric, len(resourceTypes))
	for i, resource := range resourceTypes {
		resourceMetrics[i] = Metric{
			Name: MetricNameResourceUsage,
			Labels: []MetricLabel{
				{Name: LabelNameResource, Value: resource},
			},
			Timestamp: currentTimestamp,
			Value:     float64(kubeResources[resource]),
		}
	}

	kubeletWorkingBytes, ok := kubeletMetrics["node_memory_working_set_bytes"]
	if !ok {
		return nil, 0, fmt.Errorf("metric machine_cpu_cores not found")
	}
	kubeletCapacityBytes, ok := kubeletMetrics["machine_memory_bytes"]
	if !ok {
		return nil, 0, fmt.Errorf("metric machine_memory_bytes not found")
	}
	resourceMetrics = append(resourceMetrics, Metric{
		Name:      MetricNameMemoryUsage,
		Labels:    []MetricLabel{},
		Timestamp: currentTimestamp,
		Value:     aggregateMetricValues(kubeletWorkingBytes.Metric) / aggregateMetricValues(kubeletCapacityBytes.Metric),
	})

	kubeletCPUSeconds, ok := kubeletMetrics["node_cpu_usage_seconds_total"]
	if !ok {
		return nil, 0, fmt.Errorf("metric node_cpu_usage_seconds_total not found")
	}
	cpuUsage := (aggregateMetricValues(kubeletCPUSeconds.Metric) - previousCPUSeconds) * 1000 / float64(currentTimestamp-previousTimestamp)
	if previousTimestamp == 0 {
		cpuUsage = 0
	}
	resourceMetrics = append(resourceMetrics, Metric{
		Name:      MetricNameCPUUsage,
		Labels:    []MetricLabel{},
		Timestamp: currentTimestamp,
		Value:     cpuUsage,
	})

	kubeletRunningContainers, ok := kubeletMetrics["kubelet_running_containers"]
	if !ok {
		return nil, 0, fmt.Errorf("metric kubelet_running_containers not found")
	}
	runningContainers := filterMetricsByLabel(kubeletRunningContainers.Metric, "status", "running")
	value := 0.0
	for _, metric := range runningContainers {
		value += *metric.Untyped.Value
	}
	resourceMetrics = append(resourceMetrics, Metric{
		Name:      MetricNameContainersUsage,
		Labels:    []MetricLabel{},
		Timestamp: currentTimestamp,
		Value:     value,
	})

	return resourceMetrics, aggregateMetricValues(kubeletCPUSeconds.Metric), nil
}

func computeMicroShiftVersionMetric() (Metric, error) {
	osVersion, err := util.GetOSVersion()
	if err != nil {
		return Metric{}, fmt.Errorf("failed to get OS version: %v", err)
	}
	currentTimestamp := time.Now().UnixNano() / time.Millisecond.Nanoseconds()
	version, err := prerun.GetVersionOfExecutable()
	if err != nil {
		return Metric{}, fmt.Errorf("failed to get version of executable: %v", err)
	}
	return Metric{
		Name: MetricNameMicroShiftVersion,
		Labels: []MetricLabel{
			{Name: LabelNameVersion, Value: version.String()},
			{Name: LabelNameOSVersion, Value: osVersion},
			{Name: LabelNameDeploymentType, Value: getDeploymentType()},
		},
		Timestamp: currentTimestamp,
		Value:     float64(currentTimestamp),
	}, nil
}

func convertMetricsToWriteRequest(metrics []Metric) *prompb.WriteRequest {
	timeSeriesList := make([]prompb.TimeSeries, 0)
	metricMetadataList := make([]prompb.MetricMetadata, 0)
	for _, metric := range metrics {
		labels := []prompb.Label{
			{Name: "__name__", Value: metric.Name},
		}
		for _, label := range metric.Labels {
			labels = append(labels, prompb.Label{
				Name:  label.Name,
				Value: label.Value,
			})
		}
		samples := []prompb.Sample{
			{
				Value:     metric.Value,
				Timestamp: metric.Timestamp,
			},
		}

		timeSeriesList = append(timeSeriesList, prompb.TimeSeries{
			Labels:  labels,
			Samples: samples,
		})

		metricMetadataList = append(metricMetadataList, prompb.MetricMetadata{
			MetricFamilyName: metric.Name,
			Type:             prompb.MetricMetadata_COUNTER,
		})
	}
	return &prompb.WriteRequest{
		Timeseries: timeSeriesList,
		Metadata:   metricMetadataList,
	}
}
