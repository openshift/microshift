package healthcheck

import (
	"errors"
	"sync"

	"k8s.io/klog/v2"
)

// AllErrGroup is a helper to wait for all goroutines and get all errors that occurred.
// It's based on sync.WaitGroup (which doesn't capture any errors) and errgroup.Group (which only captures the first error).
type AllErrGroup struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []error

	amount int
}

func (g *AllErrGroup) Go(f func() error) {
	g.wg.Add(1)
	g.amount += 1
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
	klog.V(2).Infof("Waiting for %d goroutines", g.amount)
	g.wg.Wait()
	return errors.Join(g.errs...)
}
