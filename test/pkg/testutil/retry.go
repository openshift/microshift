package testutil

import (
	"context"
	"errors"
	"time"

	"k8s.io/klog/v2"
)

func Retry(ctx context.Context, attempts int, interval time.Duration, f func() error) error {
	errs := []error{}
	for i := 0; i < attempts; i++ {
		err := f()
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			errs = append(errs, ctx.Err())
			break
		}

		klog.ErrorS(err, "Retrying soon", "interval", interval, "attempt", i+1, "attempts", attempts)
		errs = append(errs, err)
		time.Sleep(interval)
	}
	return errors.Join(errs...)
}
