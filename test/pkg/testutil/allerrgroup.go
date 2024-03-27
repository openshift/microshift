package testutil

import (
	"errors"
	"sync"

	"k8s.io/klog/v2"
)

// Based on sync.WaitGroup and errgroup.Group but with capturing multiple errors, not first one.

type AllErrGroup struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []error

	debug int
}

func NewAllErrGroup() *AllErrGroup {
	return &AllErrGroup{}
}

func (g *AllErrGroup) Go(f func() error) {
	g.wg.Add(1)
	g.debug += 1
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.mu.Lock()
			defer g.mu.Unlock()
			g.errs = append(g.errs, err)
		}
	}()
}

func (g *AllErrGroup) Wait() error {
	klog.InfoS("Waiting for AllErrGroup", "goroutines", g.debug)
	g.wg.Wait()
	err := errors.Join(g.errs...)
	return err
}
