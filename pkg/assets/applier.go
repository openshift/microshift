package assets

import (
	"sync"
)

var (
	lock sync.Mutex
)

type RenderFunc func([]byte) ([]byte, error)

type readerApplier interface {
	Reader([]byte, RenderFunc)
	Applier() error
}
