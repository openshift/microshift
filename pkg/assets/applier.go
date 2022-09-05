package assets

import (
	"sync"

	"github.com/openshift/library-go/pkg/operator/events"
)

var (
	lock sync.Mutex
)

var assetsEventRecorder events.Recorder = events.NewLoggingEventRecorder("microshift-assets")

type RenderParams map[string]interface{}

type RenderFunc func([]byte, RenderParams) ([]byte, error)

type readerApplier interface {
	Reader([]byte, RenderFunc, RenderParams)
	Applier() error
}
