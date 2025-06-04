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

// Package output_test shows how to use k8s.io/klog/v2/test
// and provides unit testing with dependencies that wouldn't
// be acceptable for the main module.
package output_test

import (
	"io"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"k8s.io/klog/v2/test"
	"k8s.io/klog/v2/textlogger"
)

// newLogger is a test.OutputConfig.NewLogger callback which creates a zapr
// logger. The vmodule parameter is ignored because zapr does not support that.
func newLogger(out io.Writer, v int, _ string) logr.Logger {
	return newZaprLogger(out, v)
}

// TestZaprOutput tests the zapr, directly and as backend.
func TestZaprOutput(t *testing.T) {
	test.InitKlog(t)
	t.Run("direct", func(t *testing.T) {
		test.Output(t, test.OutputConfig{NewLogger: newLogger, ExpectedOutputMapping: test.ZaprOutputMappingDirect()})
	})
	t.Run("klog-backend", func(t *testing.T) {
		test.Output(t, test.OutputConfig{NewLogger: newLogger, AsBackend: true, ExpectedOutputMapping: test.ZaprOutputMappingIndirect()})
	})
}

// Benchmark direct zapr output.
func BenchmarkZaprOutput(b *testing.B) {
	test.InitKlog(b)
	test.Benchmark(b, test.OutputConfig{NewLogger: newLogger, ExpectedOutputMapping: test.ZaprOutputMappingDirect()})
}

// TestKlogrStackText tests klogr.klogr -> klog -> text logger.
func TestKlogrStackText(t *testing.T) {
	test.InitKlog(t)
	newLogger := func(out io.Writer, v int, vmodule string) logr.Logger {
		// Backend: text output.
		config := textlogger.NewConfig(
			textlogger.Verbosity(v),
			textlogger.Output(out),
		)
		if err := config.VModule().Set(vmodule); err != nil {
			panic(err)
		}
		klog.SetLogger(textlogger.NewLogger(config))

		// Frontend: klogr.
		return klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog))
	}
	test.Output(t, test.OutputConfig{NewLogger: newLogger, SupportsVModule: true})
}

// TestKlogrStackKlogr tests klogr.klogr -> klog -> zapr.
//
// This exposes whether verbosity is passed through correctly
// (https://github.com/kubernetes/klog/issues/294) because klogr logging
// records that.
func TestKlogrStackZapr(t *testing.T) {
	test.InitKlog(t)
	mapping := test.ZaprOutputMappingIndirect()

	// klogr doesn't warn about invalid KVs and just inserts
	// "(MISSING)".
	for key, value := range map[string]string{
		`I output.go:<LINE>] "odd arguments" akey="avalue" akey2="(MISSING)"
`: `{"caller":"test/output.go:<LINE>","msg":"odd arguments","v":0,"akey":"avalue","akey2":"(MISSING)"}
`,

		`I output.go:<LINE>] "both odd" basekey1="basevar1" basekey2="(MISSING)" akey="avalue" akey2="(MISSING)"
`: `{"caller":"test/output.go:<LINE>","msg":"both odd","v":0,"basekey1":"basevar1","basekey2":"(MISSING)","akey":"avalue","akey2":"(MISSING)"}
`,
		`I output.go:<LINE>] "integer keys" %!s(int=1)="value" %!s(int=2)="value2" akey="avalue" akey2="(MISSING)"
`: `{"caller":"test/output.go:<LINE>","msg":"non-string key argument passed to logging, ignoring all later arguments","invalid key":1}
{"caller":"test/output.go:<LINE>","msg":"integer keys","v":0}
`,
		`I output.go:<LINE>] "struct keys" {name}="value" test="other value" key="val"
`: `{"caller":"test/output.go:<LINE>","msg":"non-string key argument passed to logging, ignoring all later arguments","invalid key":{}}
{"caller":"test/output.go:<LINE>","msg":"struct keys","v":0}
`,
		`I output.go:<LINE>] "map keys" map[test:%!s(bool=true)]="test"
`: `{"caller":"test/output.go:<LINE>","msg":"non-string key argument passed to logging, ignoring all later arguments","invalid key":{"test":true}}
{"caller":"test/output.go:<LINE>","msg":"map keys","v":0}
`,
	} {
		mapping[key] = value
	}

	newLogger := func(out io.Writer, v int, _ string) logr.Logger {
		// Backend: zapr as configured in k8s.io/component-base/logs/json.
		klog.SetLogger(newZaprLogger(out, v))

		// Frontend: klogr.
		return klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog))
	}
	test.Output(t, test.OutputConfig{NewLogger: newLogger, ExpectedOutputMapping: mapping})
}

// TestKlogrInternalStackText tests klog.klogr (the simplified version used for contextual logging) -> klog -> text logger.
func TestKlogrInternalStackText(t *testing.T) {
	test.InitKlog(t)
	newLogger := func(out io.Writer, v int, vmodule string) logr.Logger {
		// Backend: text output.
		config := textlogger.NewConfig(
			textlogger.Verbosity(v),
			textlogger.Output(out),
		)
		if err := config.VModule().Set(vmodule); err != nil {
			panic(err)
		}
		klog.SetLogger(textlogger.NewLogger(config))

		// Frontend: internal klogr.
		return klog.NewKlogr()
	}
	test.Output(t, test.OutputConfig{NewLogger: newLogger, SupportsVModule: true})
}

// TestKlogrInternalStackKlogr tests klog.klogr (the simplified version used for contextual logging) -> klog -> zapr.
//
// This exposes whether verbosity is passed through correctly
// (https://github.com/kubernetes/klog/issues/294) because klogr logging
// records that.
func TestKlogrInternalStackZapr(t *testing.T) {
	test.InitKlog(t)
	mapping := test.ZaprOutputMappingIndirect()

	// klogr doesn't warn about invalid KVs and just inserts
	// "(MISSING)".
	for key, value := range map[string]string{
		`I output.go:<LINE>] "odd arguments" akey="avalue" akey2="(MISSING)"
`: `{"caller":"test/output.go:<LINE>","msg":"odd arguments","v":0,"akey":"avalue","akey2":"(MISSING)"}
`,

		`I output.go:<LINE>] "both odd" basekey1="basevar1" basekey2="(MISSING)" akey="avalue" akey2="(MISSING)"
`: `{"caller":"test/output.go:<LINE>","msg":"both odd","v":0,"basekey1":"basevar1","basekey2":"(MISSING)","akey":"avalue","akey2":"(MISSING)"}
`,
		`I output.go:<LINE>] "integer keys" %!s(int=1)="value" %!s(int=2)="value2" akey="avalue" akey2="(MISSING)"
`: `{"caller":"test/output.go:<LINE>","msg":"non-string key argument passed to logging, ignoring all later arguments","invalid key":1}
{"caller":"test/output.go:<LINE>","msg":"integer keys","v":0}
`,
		`I output.go:<LINE>] "struct keys" {name}="value" test="other value" key="val"
`: `{"caller":"test/output.go:<LINE>","msg":"non-string key argument passed to logging, ignoring all later arguments","invalid key":{}}
{"caller":"test/output.go:<LINE>","msg":"struct keys","v":0}
`,
		`I output.go:<LINE>] "map keys" map[test:%!s(bool=true)]="test"
`: `{"caller":"test/output.go:<LINE>","msg":"non-string key argument passed to logging, ignoring all later arguments","invalid key":{"test":true}}
{"caller":"test/output.go:<LINE>","msg":"map keys","v":0}
`,
	} {
		mapping[key] = value
	}

	newLogger := func(out io.Writer, v int, _ string) logr.Logger {
		// Backend: zapr as configured in k8s.io/component-base/logs/json.
		klog.SetLogger(newZaprLogger(out, v))

		// Frontend: internal klogr.
		return klog.NewKlogr()
	}
	test.Output(t, test.OutputConfig{NewLogger: newLogger, ExpectedOutputMapping: mapping})
}

func newZaprLogger(out io.Writer, v int) logr.Logger {
	encoderConfig := &zapcore.EncoderConfig{
		MessageKey:     "msg",
		CallerKey:      "caller",
		NameKey:        "logger",
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	encoder := zapcore.NewJSONEncoder(*encoderConfig)
	zapV := -zapcore.Level(v)
	core := zapcore.NewCore(encoder, zapcore.AddSync(out), zapV)
	l := zap.New(core, zap.WithCaller(true))
	logger := zapr.NewLoggerWithOptions(l, zapr.LogInfoLevel("v"), zapr.ErrorKey("err"))
	return logger
}
