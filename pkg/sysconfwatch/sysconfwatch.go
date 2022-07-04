/*
Copyright © 2022 Microshift Contributors

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
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

const sysConfigCheckInterval = time.Second * 5

type SysConfWatchController struct {
	NodeIP  string
	timerFd int
}

func NewSysConfWatchController(cfg *config.MicroshiftConfig) *SysConfWatchController {
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
		NodeIP:  cfg.NodeIP,
		timerFd: fd,
	}
}

func (s *SysConfWatchController) Name() string { return "sysconfwatch-controller" }
func (s *SysConfWatchController) Dependencies() []string {
	return []string{}
}

func (c *SysConfWatchController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	ticker := time.NewTicker(sysConfigCheckInterval)
	defer ticker.Stop()

	klog.Infof("starting sysconfwatch-controller with IP address %q", c.NodeIP)

	var buf []byte = make([]byte, 8)
	for {
		select {
		case <-ticker.C:
			// Check the IP change
			currentIP, _ := util.GetHostIP()
			if c.NodeIP != currentIP {
				klog.Warningf("IP address has changed from %q to %q, restarting MicroShift", c.NodeIP, currentIP)
				os.Exit(0)
				return nil
			}

			// Check the clock change
			// Initiate an asynchronous read operation on the timer object
			// When the clock is reset, the read operation returns with the ECANCELED error code
			_, err := unix.Read(c.timerFd, buf)
			if err == unix.ECANCELED {
				klog.Warningf("realtime clock change detected, restarting MicroShift")
				os.Exit(0)
				return nil
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
