// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2026 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package klog

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"k8s.io/klog/v2/internal/buffer"
	"k8s.io/klog/v2/internal/severity"
)

// TestStderrThresholdWithLogToStderr tests the new behavior where stderrthreshold
// can be honored even when logtostderr=true, when legacy_stderr_threshold_behavior=false
func TestStderrThresholdWithLogToStderr(t *testing.T) {
	defer CaptureState().Restore()

	tests := []struct {
		name            string
		logtostderr     bool
		legacyBehavior  bool
		stderrthreshold string
		logLevel        severity.Severity
		expectInStderr  bool
		description     string
	}{
		{
			name:            "legacy behavior - logtostderr=true, all logs to stderr",
			logtostderr:     true,
			legacyBehavior:  true,
			stderrthreshold: "ERROR",
			logLevel:        severity.InfoLog,
			expectInStderr:  true,
			description:     "Legacy: INFO should appear in stderr even with stderrthreshold=ERROR",
		},
		{
			name:            "legacy behavior - logtostderr=true, ERROR threshold ignored",
			logtostderr:     true,
			legacyBehavior:  true,
			stderrthreshold: "ERROR",
			logLevel:        severity.ErrorLog,
			expectInStderr:  true,
			description:     "Legacy: ERROR should appear in stderr",
		},
		{
			name:            "new behavior - logtostderr=true, stderrthreshold honored, INFO filtered",
			logtostderr:     true,
			legacyBehavior:  false,
			stderrthreshold: "ERROR",
			logLevel:        severity.InfoLog,
			expectInStderr:  false,
			description:     "New: INFO should NOT appear in stderr with stderrthreshold=ERROR",
		},
		{
			name:            "new behavior - logtostderr=true, stderrthreshold honored, WARNING filtered",
			logtostderr:     true,
			legacyBehavior:  false,
			stderrthreshold: "ERROR",
			logLevel:        severity.WarningLog,
			expectInStderr:  false,
			description:     "New: WARNING should NOT appear in stderr with stderrthreshold=ERROR",
		},
		{
			name:            "new behavior - logtostderr=true, stderrthreshold honored, ERROR passes",
			logtostderr:     true,
			legacyBehavior:  false,
			stderrthreshold: "ERROR",
			logLevel:        severity.ErrorLog,
			expectInStderr:  true,
			description:     "New: ERROR should appear in stderr with stderrthreshold=ERROR",
		},
		{
			name:            "new behavior - logtostderr=true, stderrthreshold=WARNING, INFO filtered",
			logtostderr:     true,
			legacyBehavior:  false,
			stderrthreshold: "WARNING",
			logLevel:        severity.InfoLog,
			expectInStderr:  false,
			description:     "New: INFO should NOT appear in stderr with stderrthreshold=WARNING",
		},
		{
			name:            "new behavior - logtostderr=true, stderrthreshold=WARNING, WARNING passes",
			logtostderr:     true,
			legacyBehavior:  false,
			stderrthreshold: "WARNING",
			logLevel:        severity.WarningLog,
			expectInStderr:  true,
			description:     "New: WARNING should appear in stderr with stderrthreshold=WARNING",
		},
		{
			name:            "new behavior - logtostderr=true, stderrthreshold=INFO, all pass",
			logtostderr:     true,
			legacyBehavior:  false,
			stderrthreshold: "INFO",
			logLevel:        severity.InfoLog,
			expectInStderr:  true,
			description:     "New: INFO should appear in stderr with stderrthreshold=INFO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state for each test
			state := CaptureState()
			defer state.Restore()

			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Configure flags
			logging.mu.Lock()
			logging.toStderr = tt.logtostderr
			logging.legacyStderrThresholdBehavior = tt.legacyBehavior
			logging.stderrThreshold.Set(tt.stderrthreshold)
			logging.mu.Unlock()

			// Log message based on level
			testMsg := fmt.Sprintf("test message %s", tt.name)
			buf := buffer.GetBuffer()
			buf.WriteString(testMsg)
			buf.WriteString("\n")

			// Call output directly to test the logic
			logging.output(tt.logLevel, nil, buf, 0, "test.go", 123, false)

			// Close writer and read stderr
			w.Close()
			var stderrBuf bytes.Buffer
			io.Copy(&stderrBuf, r)
			os.Stderr = oldStderr

			stderrContent := stderrBuf.String()
			containsMsg := strings.Contains(stderrContent, testMsg)

			if tt.expectInStderr && !containsMsg {
				t.Errorf("%s: expected message in stderr but not found.\nStderr: %q", tt.description, stderrContent)
			}
			if !tt.expectInStderr && containsMsg {
				t.Errorf("%s: did not expect message in stderr but found it.\nStderr: %q", tt.description, stderrContent)
			}
		})
	}
}

// TestAlsologtostderrthreshold tests the new alsologtostderrthreshold flag
func TestAlsologtostderrthreshold(t *testing.T) {
	defer CaptureState().Restore()

	tests := []struct {
		name                     string
		alsologtostderr          bool
		alsologtostderrthreshold string
		logLevel                 severity.Severity
		expectInStderr           bool
		description              string
	}{
		{
			name:                     "alsologtostderr=true, threshold=ERROR, INFO filtered",
			alsologtostderr:          true,
			alsologtostderrthreshold: "ERROR",
			logLevel:                 severity.InfoLog,
			expectInStderr:           false,
			description:              "INFO should NOT appear in stderr with alsologtostderrthreshold=ERROR",
		},
		{
			name:                     "alsologtostderr=true, threshold=ERROR, WARNING filtered",
			alsologtostderr:          true,
			alsologtostderrthreshold: "ERROR",
			logLevel:                 severity.WarningLog,
			expectInStderr:           false,
			description:              "WARNING should NOT appear in stderr with alsologtostderrthreshold=ERROR",
		},
		{
			name:                     "alsologtostderr=true, threshold=ERROR, ERROR passes",
			alsologtostderr:          true,
			alsologtostderrthreshold: "ERROR",
			logLevel:                 severity.ErrorLog,
			expectInStderr:           true,
			description:              "ERROR should appear in stderr with alsologtostderrthreshold=ERROR",
		},
		{
			name:                     "alsologtostderr=true, threshold=WARNING, INFO filtered",
			alsologtostderr:          true,
			alsologtostderrthreshold: "WARNING",
			logLevel:                 severity.InfoLog,
			expectInStderr:           false,
			description:              "INFO should NOT appear in stderr with alsologtostderrthreshold=WARNING",
		},
		{
			name:                     "alsologtostderr=true, threshold=WARNING, WARNING passes",
			alsologtostderr:          true,
			alsologtostderrthreshold: "WARNING",
			logLevel:                 severity.WarningLog,
			expectInStderr:           true,
			description:              "WARNING should appear in stderr with alsologtostderrthreshold=WARNING",
		},
		{
			name:                     "alsologtostderr=true, default threshold (INFO), all pass",
			alsologtostderr:          true,
			alsologtostderrthreshold: "INFO",
			logLevel:                 severity.InfoLog,
			expectInStderr:           true,
			description:              "INFO should appear in stderr with alsologtostderrthreshold=INFO (default)",
		},
		{
			name:                     "alsologtostderr=false, threshold=ERROR, no stderr",
			alsologtostderr:          false,
			alsologtostderrthreshold: "ERROR",
			logLevel:                 severity.ErrorLog,
			expectInStderr:           false,
			description:              "ERROR should NOT appear in stderr when alsologtostderr=false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state for each test
			state := CaptureState()
			defer state.Restore()

			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Configure flags - disable logtostderr and enable file logging
			logging.mu.Lock()
			logging.toStderr = false
			logging.alsoToStderr = tt.alsologtostderr
			logging.alsologtostderrthreshold.Set(tt.alsologtostderrthreshold)
			logging.stderrThreshold.Set("FATAL") // Set high to avoid interference
			// Use buffer writers instead of files
			logging.file = [severity.NumSeverity]io.Writer{
				new(flushBuffer), new(flushBuffer), new(flushBuffer), new(flushBuffer),
			}
			logging.mu.Unlock()

			// Log message based on level
			testMsg := fmt.Sprintf("test message %s", tt.name)
			buf := buffer.GetBuffer()
			buf.WriteString(testMsg)
			buf.WriteString("\n")

			// Call output directly to test the logic
			logging.output(tt.logLevel, nil, buf, 0, "test.go", 123, false)

			// Close writer and read stderr
			w.Close()
			var stderrBuf bytes.Buffer
			io.Copy(&stderrBuf, r)
			os.Stderr = oldStderr

			stderrContent := stderrBuf.String()
			containsMsg := strings.Contains(stderrContent, testMsg)

			if tt.expectInStderr && !containsMsg {
				t.Errorf("%s: expected message in stderr but not found.\nStderr: %q", tt.description, stderrContent)
			}
			if !tt.expectInStderr && containsMsg {
				t.Errorf("%s: did not expect message in stderr but found it.\nStderr: %q", tt.description, stderrContent)
			}
		})
	}
}

// TestFlagParsing tests that the new flags can be parsed correctly
func TestNewFlagParsing(t *testing.T) {
	defer CaptureState().Restore()

	// Create a new flag set for testing
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	InitFlags(fs)

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name: "valid legacy_stderr_threshold_behavior=true",
			args: []string{"-legacy_stderr_threshold_behavior=true"},
		},
		{
			name: "valid legacy_stderr_threshold_behavior=false",
			args: []string{"-legacy_stderr_threshold_behavior=false"},
		},
		{
			name: "valid alsologtostderrthreshold=ERROR",
			args: []string{"-alsologtostderrthreshold=ERROR"},
		},
		{
			name: "valid alsologtostderrthreshold=WARNING",
			args: []string{"-alsologtostderrthreshold=WARNING"},
		},
		{
			name: "valid alsologtostderrthreshold=INFO",
			args: []string{"-alsologtostderrthreshold=INFO"},
		},
		{
			name:        "invalid alsologtostderrthreshold",
			args:        []string{"-alsologtostderrthreshold=INVALID"},
			expectError: true,
		},
		{
			name: "combined flags",
			args: []string{
				"-logtostderr=true",
				"-legacy_stderr_threshold_behavior=false",
				"-stderrthreshold=ERROR",
				"-alsologtostderrthreshold=WARNING",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			fs = flag.NewFlagSet("test", flag.ContinueOnError)
			InitFlags(fs)

			err := fs.Parse(tt.args)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
