//go:build linux
// +build linux

/*
Copyright 2024 The Kubernetes Authors.

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

package watchdog

import (
	"bytes"
	"errors"
	"flag"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"k8s.io/klog/v2"
)

// Mock syncLoopHealthChecker
type mockSyncLoopHealthChecker struct {
	healthCheckErr error
}

func (m *mockSyncLoopHealthChecker) SyncLoopHealthCheck(req *http.Request) error {
	return m.healthCheckErr
}

// Mock WatchdogClient
type mockWatchdogClient struct {
	enabledVal time.Duration
	enabledErr error
	notifyAck  bool
	notifyErr  error
}

func (m *mockWatchdogClient) SdWatchdogEnabled(unsetEnvironment bool) (time.Duration, error) {
	return m.enabledVal, m.enabledErr
}

func (m *mockWatchdogClient) SdNotify(unsetEnvironment bool) (bool, error) {
	return m.notifyAck, m.notifyErr
}

const (
	interval      = 4 * time.Second
	intervalSmall = 1 * time.Second
)

// TestNewHealthChecker tests the NewHealthChecker function.
func TestNewHealthChecker(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		mockEnabled time.Duration
		mockErr     error
		wantErr     bool
	}{
		{"Watchdog enabled", interval, nil, false},
		{"Watchdog not enabled", 0, nil, false},
		{"Watchdog enabled with error", interval, errors.New("mock error"), true},
		{"Watchdog timeout too small", intervalSmall, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockWatchdogClient{
				enabledVal: tt.mockEnabled,
				enabledErr: tt.mockErr,
			}

			_, err := NewHealthChecker(&mockSyncLoopHealthChecker{}, WithWatchdogClient(mockClient))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHealthChecker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHealthCheckerStart tests the Start method of the healthChecker.
func TestHealthCheckerStart(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		enabledVal     time.Duration
		healthCheckErr error
		notifyAck      bool
		notifyErr      error
		expectedLogs   []string
	}{
		{
			name:           "Watchdog enabled and notify succeeds",
			enabledVal:     interval,
			healthCheckErr: nil,
			notifyAck:      true,
			notifyErr:      nil,
			expectedLogs:   []string{"Starting systemd watchdog with interval", "Watchdog plugin notified"},
		},
		{
			name:           "Watchdog enabled and notify fails, notification not supported",
			enabledVal:     interval,
			healthCheckErr: nil,
			notifyAck:      false,
			notifyErr:      nil,
			expectedLogs:   []string{"Starting systemd watchdog with interval", "Failed to notify watchdog", "notification not supported"},
		},
		{
			name:           "Watchdog enabled and notify fails, transmission failed",
			enabledVal:     interval,
			healthCheckErr: nil,
			notifyAck:      false,
			notifyErr:      errors.New("mock notify error"),
			expectedLogs:   []string{"Starting systemd watchdog with interval", "Failed to notify watchdog"},
		},
		{
			name:           "Watchdog enabled and health check fails",
			enabledVal:     interval,
			healthCheckErr: errors.New("mock healthy error"),
			notifyAck:      true,
			notifyErr:      nil,
			expectedLogs:   []string{"Starting systemd watchdog with interval", "Do not notify watchdog this iteration as the kubelet is reportedly not healthy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture logs
			logBuffer := setupLogging(t)

			// Mock SdWatchdogEnabled to return a valid value
			mockClient := &mockWatchdogClient{
				enabledVal: tt.enabledVal,
				notifyAck:  tt.notifyAck,
				notifyErr:  tt.notifyErr,
			}

			// Create a healthChecker
			hc, err := NewHealthChecker(&mockSyncLoopHealthChecker{healthCheckErr: tt.healthCheckErr}, WithWatchdogClient(mockClient))
			if err != nil {
				t.Fatalf("NewHealthChecker() failed: %v", err)
			}

			// Start the health checker
			hc.Start()

			// Wait for a short period to allow the health check to run
			time.Sleep(2 * interval)

			// Check logs to verify the health check ran
			klog.Flush()
			logs := logBuffer.String()
			for _, expectedLog := range tt.expectedLogs {
				if !strings.Contains(logs, expectedLog) {
					t.Errorf("Expected log '%s' not found in logs: %s", expectedLog, logs)
				}
			}
		})
	}
}

// threadSafeBuffer is a thread-safe wrapper around bytes.Buffer.
type threadSafeBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (b *threadSafeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(p)
}

func (b *threadSafeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}

// setupLogging sets up logging to capture output using a thread-safe buffer.
func setupLogging(t *testing.T) *threadSafeBuffer {
	flags := &flag.FlagSet{}
	klog.InitFlags(flags)
	if err := flags.Set("v", "5"); err != nil {
		t.Fatal(err)
	}
	klog.LogToStderr(false)

	logBuffer := &threadSafeBuffer{}

	// Set the output to the thread-safe buffer
	klog.SetOutput(logBuffer)

	return logBuffer
}
