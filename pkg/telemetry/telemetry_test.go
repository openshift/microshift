package telemetry

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	proto "github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

func TestTelemetryClient_Send(t *testing.T) {
	clusterId := "fake-cluster-id"
	pullSecret := "fake-pull-secret"
	expectedBearer := fmt.Sprintf("Bearer %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(authString, pullSecret, clusterId))))
	sampleMetrics := []Metric{
		{
			Name:      "fake-metric",
			Labels:    []MetricLabel{{Name: "fake-label", Value: "fake-value"}},
			Timestamp: 1234567890,
			Value:     42,
		},
	}
	expectedWriteRequest := prompb.WriteRequest{
		Metadata: []prompb.MetricMetadata{
			{MetricFamilyName: "fake-metric", Type: prompb.MetricMetadata_COUNTER},
		},
		Timeseries: []prompb.TimeSeries{
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "fake-metric"},
					{Name: "fake-label", Value: "fake-value"},
				},
				Samples: []prompb.Sample{
					{Timestamp: 1234567890, Value: 42},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/x-protobuf" {
			t.Fatalf("Expected Content-Type application/x-protobuf, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Content-Encoding") != "snappy" {
			t.Fatalf("Expected Content-Encoding snappy, got %s", r.Header.Get("Content-Encoding"))
		}
		if r.Header.Get("Authorization") != expectedBearer {
			t.Fatalf("Expected Authorization '%s', got %s", expectedBearer, r.Header.Get("Authorization"))
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		writeBytes, err := snappy.Decode(nil, bodyBytes)
		if err != nil {
			t.Fatalf("Failed to decode snappy body: %v", err)
		}
		var req prompb.WriteRequest
		if err := proto.Unmarshal(writeBytes, &req); err != nil {
			t.Fatalf("failed to unmarshal WriteRequest: %v", err)
		}
		if !reflect.DeepEqual(req, expectedWriteRequest) {
			t.Fatalf("Expected WriteRequest %v, got %v", expectedWriteRequest, req)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewTelemetryClient(server.URL, clusterId, "")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := client.Send(ctx, pullSecret, sampleMetrics)
	if err != nil {
		t.Fatalf("Send method failed: %v", err)
	}
}
