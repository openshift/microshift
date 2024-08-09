package healthcheck

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

func microshiftServiceShouldBeOk(ctx context.Context, timeout time.Duration) (bool, error) {
	systemd, err := NewSystemd(ctx)
	if err != nil {
		return false, err
	}
	defer systemd.Close()

	if isEnabled, err := systemd.IsServiceEnabled(ctx, "microshift.service"); err != nil {
		return false, err
	} else if !isEnabled {
		klog.Info("microshift.service is not enabled")
		return false, nil
	}
	klog.Info("microshift.service is enabled")

	klog.Info("Waiting for microshift.service to be ready")
	if err := wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, func(ctx context.Context) (done bool, err error) {
		return systemd.IsServiceActiveAndNotFailed(ctx, "microshift.service")
	}); err != nil {
		return false, err
	}

	klog.Info("microshift.service is ready")
	return true, nil
}

func NewSystemd(ctx context.Context) (*Systemd, error) {
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		klog.ErrorS(err, "Failed to create connection to systemd")
		return nil, err
	}
	return &Systemd{connection: conn}, nil
}

type Systemd struct {
	connection *dbus.Conn
}

func (s *Systemd) Close() {
	s.connection.Close()
}

func (s *Systemd) IsServiceEnabled(ctx context.Context, service string) (bool, error) {
	if s.connection == nil {
		return false, fmt.Errorf("struct not initialized")
	}

	props, err := s.connection.GetAllPropertiesContext(ctx, service)
	if err != nil {
		klog.ErrorS(err, "Failed to get properties of service", "service", service)
		return false, err
	}

	state, ok := props["UnitFileState"]
	if !ok {
		return false, fmt.Errorf("could not find 'UnitFileState' in service properties")
	}

	return state == "enabled", nil
}

func (s *Systemd) IsServiceActiveAndNotFailed(ctx context.Context, service string) (bool, error) {
	if s.connection == nil {
		return false, fmt.Errorf("struct not initialized")
	}

	props, err := s.connection.GetAllPropertiesContext(ctx, service)
	if err != nil {
		klog.ErrorS(err, "Failed to get properties of service", "service", service)
		return false, err
	}

	activeState, ok := props["ActiveState"]
	if !ok {
		return false, fmt.Errorf("could not find 'ActiveState' in service properties")
	}

	if activeState == "failed" {
		return false, fmt.Errorf("service %s has failed", service)
	}

	if activeState == "inactive" {
		return false, fmt.Errorf("service %s is inactive", service)
	}

	// https://github.com/systemd/systemd/blob/0dd6fe931d08f17e4ee2c6410c993b7f2ffc1dd3/src/systemctl/systemctl-is-active.c#L55-L64
	return activeState == "active" || activeState == "reloading", nil
}
