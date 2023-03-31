//go:build !linux
// +build !linux

package sysconfwatch

import (
	"context"
	"time"

	"github.com/openshift/microshift/pkg/config"
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

// NewSysConfWatchController takes a config (which it ignores) in order to match the func signature of it's linux
// variant. see sysconfwatch_linux.go
func NewSysConfWatchController(_ *config.Config) *nonLinuxSysConfWatchController {
	return &nonLinuxSysConfWatchController{}
}
