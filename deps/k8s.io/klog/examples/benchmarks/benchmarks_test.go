/*
Copyright 2022 The Kubernetes Authors.

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

package benchmarks

import (
	"flag"
	"fmt"
	"io"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"k8s.io/klog/examples/util/require"
	"k8s.io/klog/v2"
)

const (
	verbosityThreshold = 10
)

func init() {
	// klog gets configured so that it writes to a single output file that
	// will be set during tests with SetOutput.
	klog.InitFlags(nil)
	require.NoError(flag.Set("v", fmt.Sprintf("%d", verbosityThreshold)))
	require.NoError(flag.Set("log_file", "/dev/null"))
	require.NoError(flag.Set("logtostderr", "false"))
	require.NoError(flag.Set("alsologtostderr", "false"))
	require.NoError(flag.Set("stderrthreshold", "10"))
}

type testcase struct {
	name     string
	generate func() interface{}
}

func BenchmarkOutput(b *testing.B) {
	// We'll run each benchmark for different output formatting.
	configs := map[string]struct {
		init, cleanup func()
	}{
		"klog": {
			init: func() { klog.SetOutput(discard{}) },
		},
		"zapr": {
			init:    func() { klog.SetLogger(newZaprLogger()) },
			cleanup: func() { klog.ClearLogger() },
		},
	}

	// Each benchmark tests formatting of one key/value pair, with
	// different values. The order is relevant here.
	var tests []testcase
	for length := 0; length <= 100; length += 10 {
		arg := make([]interface{}, length)
		for i := 0; i < length; i++ {
			arg[i] = KMetadataMock{Name: "a", NS: "a"}
		}
		tests = append(tests, testcase{
			name: fmt.Sprintf("objects/%d", length),
			generate: func() interface{} {
				return klog.KObjSlice(arg)
			},
		})
	}

	// Verbosity checks may influence the result.
	verbosity := map[string]func(value interface{}){
		"no-verbosity-check": func(value interface{}) {
			klog.InfoS("test", "key", value)
		},
		"pass-verbosity-check": func(value interface{}) {
			klog.V(verbosityThreshold).InfoS("test", "key", value)
		},
		"fail-verbosity-check": func(value interface{}) {
			klog.V(verbosityThreshold+1).InfoS("test", "key", value)
		},
		"non-standard-int-key-check": func(value interface{}) {
			klog.InfoS("test", 1, value)
		},
		"non-standard-struct-key-check": func(value interface{}) {
			klog.InfoS("test", struct{ key string }{"test"}, value)
		},
		"non-standard-map-key-check": func(value interface{}) {
			klog.InfoS("test", map[string]bool{"key": true}, value)
		},
		"pass-verbosity-non-standard-int-key-check": func(value interface{}) {
			klog.V(verbosityThreshold).InfoS("test", 1, value)
		},
		"pass-verbosity-non-standard-struct-key-check": func(value interface{}) {
			klog.V(verbosityThreshold).InfoS("test", struct{ key string }{"test"}, value)
		},
		"pass-verbosity-non-standard-map-key-check": func(value interface{}) {
			klog.V(verbosityThreshold).InfoS("test", map[string]bool{"key": true}, value)
		},
		"fail-verbosity-non-standard-int-key-check": func(value interface{}) {
			klog.V(verbosityThreshold+1).InfoS("test", 1, value)
		},
		"fail-verbosity-non-standard-struct-key-check": func(value interface{}) {
			klog.V(verbosityThreshold+1).InfoS("test", struct{ key string }{"test"}, value)
		},
		"fail-verbosity-non-standard-map-key-check": func(value interface{}) {
			klog.V(verbosityThreshold+1).InfoS("test", map[string]bool{"key": true}, value)
		},
	}

	for name, config := range configs {
		b.Run(name, func(b *testing.B) {
			if config.cleanup != nil {
				defer config.cleanup()
			}
			config.init()

			for name, logCall := range verbosity {
				b.Run(name, func(b *testing.B) {
					for _, testcase := range tests {
						b.Run(testcase.name, func(b *testing.B) {
							b.ResetTimer()
							for i := 0; i < b.N; i++ {
								logCall(testcase.generate())
							}
						})
					}
				})
			}
		})
	}
}

func newZaprLogger() logr.Logger {
	encoderConfig := &zapcore.EncoderConfig{
		MessageKey:     "msg",
		CallerKey:      "caller",
		NameKey:        "logger",
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	encoder := zapcore.NewJSONEncoder(*encoderConfig)
	zapV := -zapcore.Level(verbosityThreshold)
	core := zapcore.NewCore(encoder, zapcore.AddSync(discard{}), zapV)
	l := zap.New(core, zap.WithCaller(true))
	logger := zapr.NewLoggerWithOptions(l, zapr.LogInfoLevel("v"), zapr.ErrorKey("err"))
	return logger
}

type KMetadataMock struct {
	Name, NS string
}

func (m KMetadataMock) GetName() string {
	return m.Name
}
func (m KMetadataMock) GetNamespace() string {
	return m.NS
}

type discard struct{}

var _ io.Writer = discard{}

func (discard) Write(p []byte) (int, error) {
	return len(p), nil
}
