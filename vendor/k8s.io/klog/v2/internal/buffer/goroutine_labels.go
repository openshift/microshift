package buffer

import (
	"context"
	"runtime/pprof"
	"unsafe"
)

//go:linkname runtime_getProfLabel runtime/pprof.runtime_getProfLabel
func runtime_getProfLabel() unsafe.Pointer

// Definitions of 'label' and 'LabelSet' from /usr/local/go1.24.4/src/runtime/pprof/label.go

type label struct {
	key   string
	value string
}
type LabelSet struct {
	list []label
}

func getMicroshiftLoggerComponent() string {
	labels := (*LabelSet)(runtime_getProfLabel())
	if labels == nil {
		return "???"
	}

	for _, label := range labels.list {
		if label.key == "microshift_logger_component" {
			return label.value
		}
	}

	return "???"
}

func WithMicroshiftLoggerComponent(c string, f func()) {
	pprof.Do(context.Background(), pprof.Labels("microshift_logger_component", c), func(context.Context) {
		f()
	})
}
