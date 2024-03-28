package assets

import (
	"context"
	"sync"

	"github.com/openshift/library-go/pkg/operator/events"
)

var (
	lock sync.Mutex
)

var assetsEventRecorder events.Recorder = events.NewLoggingEventRecorder("microshift-assets")

type RenderParams map[string]interface{}

type RenderFunc func([]byte, RenderParams) ([]byte, error)

type resourceHandler interface {
	Read([]byte, RenderFunc, RenderParams)
	Handle(ctx context.Context) error
}
