package assets

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/openshift/library-go/pkg/operator/events"
	"k8s.io/utils/clock"
)

var (
	lock sync.Mutex
)

var assetsEventRecorder events.Recorder = events.NewLoggingEventRecorder("microshift-assets", clock.RealClock{})

type RenderParams map[string]interface{}

// DEPRECATED: RenderFunc is deprecated and should not be used.
// TODO: Remove RenderFunc once all assets are migrated
type RenderFunc func([]byte, RenderParams) ([]byte, error)

// RenderFuncV2 is a better version of RenderFunc, using io.Reader
type RenderFuncV2 func(io.Reader, RenderParams) (io.Reader, error)

// ToRenderFuncV2 converts RenderFunc to RenderFuncV2 by reading the data from io.Reader
// This is a bottleneck for stream processing, but it is necessary for the transition.
func ToRenderFuncV2(f RenderFunc) RenderFuncV2 {
	return func(r io.Reader, params RenderParams) (io.Reader, error) {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		data, err = f(data, params)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(data), nil
	}
}

// DEPRECATED: resourceHandler is deprecated and should not be used.
// TODO: Remove resourceHandler once all assets are migrated
type resourceHandler interface {
	Read([]byte, RenderFunc, RenderParams)
	Handle(ctx context.Context) error
}

// resourceHandlerV2 is a better version of resourceHandler, using
// io.Reader instead of []byte for Read method and returning error instead of panicking.
type resourceHandlerV2 interface {
	Read(io.Reader, RenderFuncV2, RenderParams) error
	Handle(ctx context.Context) error
}
