//go:build !linux
// +build !linux

package sysconfwatch

import (
	"context"
	"github.com/openshift/microshift/pkg/config"
	"time"
)

type nonLinuxSysConfWatchController struct{}

func (n *nonLinuxSysConfWatchController) Name() string {
	return "non-linux-sysconfwatch-controller"
}

func (n *nonLinuxSysConfWatchController) Dependencies() []string {
	return []string{}
}

func (n *nonLinuxSysConfWatchController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// checking
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func NewSysConfWatchController(cfg *config.MicroshiftConfig) *nonLinuxSysConfWatchController {
	return &nonLinuxSysConfWatchController{}
}
