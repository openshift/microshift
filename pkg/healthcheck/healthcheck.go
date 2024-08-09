package healthcheck

import (
	"context"
)

func MicroShiftHealthcheck(ctx context.Context) error {
	timeout := getGreenbootTimeoutDuration()

	if enabled, err := microshiftServiceShouldBeOk(ctx, timeout); err != nil {
		return err
	} else if !enabled {
		return nil
	}

	return nil
}
