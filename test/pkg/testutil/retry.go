package testutil

import (
	"errors"
	"time"

	"k8s.io/klog/v2"
)

const (
	interval = time.Second * 10
)

func Retry(attempts int, f func() error) error {
	errs := []error{}
	for i := 0; i < attempts; i++ {
		err := f()
		if err == nil {
			return nil
		}

		klog.ErrorS(err, "Retrying soon", "interval", interval, "attempt", i+1, "attempts", attempts)
		errs = append(errs, err)
		time.Sleep(interval)
	}
	return errors.Join(errs...)
}
