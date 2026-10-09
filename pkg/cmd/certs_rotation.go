package cmd

import (
	"context"
	"time"

	"k8s.io/klog/v2"
)

// The policy is resolved at startup, like the rest of MicroShift's configuration.
// Online renewal does not change this deadline or the active certificate set.
func watchCertificateRotation(ctx context.Context, stop context.CancelFunc, rotationDate time.Time, forceRestart bool) {
	deadline, cancel := context.WithDeadline(ctx, rotationDate)
	defer cancel()
	<-deadline.Done()
	if ctx.Err() != nil {
		return
	}
	if forceRestart {
		klog.Info("Stopping services for certificate rotation")
		stop()
		return
	}
	klog.Warning("Certificates have reached ExpirationImminent; automatic certificate restart is disabled. Use microshift certs status to inspect active certificates, prepare renewal, and restart MicroShift before they expire")
}
