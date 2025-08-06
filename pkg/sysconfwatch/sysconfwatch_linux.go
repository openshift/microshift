/*
Copyright © 2022 MicroShift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sysconfwatch

import (
	"context"
	"math"
	"os"
	"syscall"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

const sysConfigCheckInterval = time.Second * 5
const sysConfigAllowedTimeDrift = time.Second * 10

type SysConfWatchController struct {
	NodeIP       string
	NodeIPv6     string
	userNodeIP   string
	userNodeIPv6 string
	timerFd      int
}

func NewSysConfWatchController(cfg *config.Config) *SysConfWatchController {
	// Create a realtime clock timer with asynchronous read support
	fd, err := unix.TimerfdCreate(unix.CLOCK_REALTIME, unix.TFD_CLOEXEC|unix.TFD_NONBLOCK)
	if err != nil {
		klog.Fatalf("failed to create a realtime clock timer %v", err)
	}

	// Set the time interval into distant future
	var ptime = &unix.ItimerSpec{
		Interval: unix.Timespec{Sec: math.MaxInt64, Nsec: 0},
		Value:    unix.Timespec{Sec: 0, Nsec: 0},
	}
	// Start the timer with cancelation if the clock is reset elsewhere
	err = unix.TimerfdSettime(fd, unix.TFD_TIMER_ABSTIME|unix.TFD_TIMER_CANCEL_ON_SET, ptime, nil)
	if err != nil {
		klog.Fatalf("failed to start a realtime clock timer %v", err)
	}
	return &SysConfWatchController{
		NodeIP:       cfg.Node.NodeIP,
		NodeIPv6:     cfg.Node.NodeIPV6,
		userNodeIP:   cfg.UserNodeIP(),
		userNodeIPv6: cfg.UserNodeIPv6(),
		timerFd:      fd,
	}
}

func (s *SysConfWatchController) Name() string { return "sysconfwatch-controller" }
func (s *SysConfWatchController) Dependencies() []string {
	return []string{}
}

func getSysMonTimes() (int64, int64) {
	var stm unix.Timespec
	var mtm unix.Timespec

	// System-wide clock that measures real (i.e. wall-clock) time
	// This clock is affected by discontinuous jumps in the system time (e.g. if the system administrator manually changes the clock)
	// and by the incremental adjustments performed by adjtime and NTP
	if err := unix.ClockGettime(unix.CLOCK_REALTIME, &stm); err != nil {
		klog.Errorf("failed to init clock: %v", err)
	}

	// Clock that cannot be set and represents monotonic time since some unspecified starting point
	// It provides access to a raw hardware-based time that is not subject to NTP adjustments
	// or the incremental adjustments performed by adjtime
	if err := unix.ClockGettime(unix.CLOCK_MONOTONIC_RAW, &mtm); err != nil {
		klog.Errorf("failed to initialize clock: %v", err)
	}

	return stm.Sec, mtm.Sec
}

func sendSigterm() {
	pid := os.Getpid()
	klog.Infof("Sending SIGTERM to self to initiate graceful shutdown. PID: %d", pid)
	p, _ := os.FindProcess(pid)
	err := p.Signal(syscall.SIGTERM)
	if err != nil {
		klog.Errorf("failed to send SIGTERM to self. Forcing shutdown: %v", err)
		os.Exit(1)
	}
}

func (c *SysConfWatchController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	ticker := time.NewTicker(sysConfigCheckInterval)
	defer ticker.Stop()

	klog.Infof("starting sysconfwatch-controller with IP address %q", c.NodeIP)

	var buf []byte = make([]byte, 8)
	// Take a snapshot of the system and monototic clocks as a base reference
	stimeRef, mtimeRef := getSysMonTimes()

	klog.Infof("sysconfwatch-controller is ready")
	close(ready)
	for {
		select {
		case <-ticker.C:
			// Check the IP change
			currentIP, err := util.GetHostIP(c.userNodeIP)
			if err != nil {
				klog.Warningf("cannot find an host IP: %v", err)
				go sendSigterm()
				return nil
			}
			if c.NodeIP != currentIP {
				klog.Warningf("IP address has changed from %q to %q, restarting MicroShift", c.NodeIP, currentIP)
				go sendSigterm()
				return nil
			}
			// Dual stack case
			if c.NodeIPv6 != "" {
				currentIP, err = util.GetHostIPv6(c.userNodeIPv6)
				if err != nil {
					klog.Warningf("cannot find an host IP: %v", err)
					go sendSigterm()
					return nil
				}
				if c.NodeIPv6 != currentIP {
					klog.Warningf("IP address has changed from %q to %q, restarting MicroShift", c.NodeIPv6, currentIP)
					go sendSigterm()
					return nil
				}
			}

			// Check the clock change by initiating an asynchronous read operation on the timer object
			// When the clock is reset, the read operation returns with the ECANCELED error code
			_, err = unix.Read(c.timerFd, buf)
			if err == unix.ECANCELED {
				// Take a snapshot of the current system and monototic clocks
				stimeCur, mtimeCur := getSysMonTimes()

				// Compare the elapsed time for the current and base references
				// Verify that the time drift is in the allowed range
				var stimeDiff = stimeCur - stimeRef
				var mtimeDiff = mtimeCur - mtimeRef
				var smtDiffDrift = stimeDiff - mtimeDiff
				if math.Abs(float64(smtDiffDrift)) < sysConfigAllowedTimeDrift.Seconds() {
					// Allow time adjustments when the drift is the predefined range
					// This comes to prevent restarts when small time adjustments are performed by NTP
					klog.Warningf("realtime clock change detected, time drifted %v seconds within the allowed range", smtDiffDrift)
					// Update the base references to allow cumulative time adjustments to remain in the allowed range
					stimeRef = stimeCur
					mtimeRef = mtimeCur
				} else {
					klog.Warningf("realtime clock change detected, time drifted %v seconds, restarting MicroShift", smtDiffDrift)
					go sendSigterm()
					return nil
				}
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
