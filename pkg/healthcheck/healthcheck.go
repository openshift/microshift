package healthcheck

import (
	"context"
	"time"
)

func MicroShiftHealthcheck(ctx context.Context, timeout time.Duration) error {
	if enabled, err := microshiftServiceShouldBeOk(ctx, timeout); err != nil {
		return err
	} else if !enabled {
		return nil
	}

	return nil
}
