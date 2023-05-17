package buffer

import (
	"context"
	"runtime/pprof"
	"unsafe"
)

//go:linkname runtime_getProfLabel runtime/pprof.runtime_getProfLabel
func runtime_getProfLabel() unsafe.Pointer

func getMicroshiftLoggerComponent() string {
	labels := (*map[string]string)(runtime_getProfLabel())
	if labels == nil {
		return "???"
	}

	c, ok := (*labels)["microshift_logger_component"]
	if !ok {
		return "???"
	}

	return c
}

func WithMicroshiftLoggerComponent(c string, f func()) {
	pprof.Do(context.Background(), pprof.Labels("microshift_logger_component", c), func(context.Context) {
		f()
	})
}
