package healthcheck

import (
	"context"
	"time"
)

func MicroShiftHealthcheck(ctx context.Context, timeout time.Duration) error {
	if enabled, err := microshiftServiceShouldBeOk(ctx, timeout); err != nil {
		printPrerunLog()
		return err
	} else if !enabled {
		return nil
	}

	if err := apiServerShouldBeLiveAndReady(ctx, timeout); err != nil {
		return err
	}

	return nil
}
