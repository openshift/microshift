package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCertificateRotationRestartPolicy(t *testing.T) {
	for _, forceRestart := range []bool{true, false} {
		t.Run(map[bool]string{true: "enabled", false: "disabled"}[forceRestart], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			watchCertificateRotation(ctx, cancel, time.Now().Add(-time.Second), forceRestart)
			if forceRestart {
				require.ErrorIs(t, ctx.Err(), context.Canceled)
			} else {
				require.NoError(t, ctx.Err(), "critical expiry must not stop a warn-only service")
			}
		})
	}
}

func TestCertificateRotationWatcherStopsWithService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	watchCertificateRotation(ctx, func() { t.Fatal("service cancellation must not trigger certificate rotation") }, time.Now().Add(time.Hour), true)
}
