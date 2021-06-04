package assets

import (
	"sync"
)

var (
	lock sync.Mutex
)

type RenderParams map[string]string

type RenderFunc func([]byte, RenderParams) ([]byte, error)

type readerApplier interface {
	Reader([]byte, RenderFunc, RenderParams)
	Applier() error
}
