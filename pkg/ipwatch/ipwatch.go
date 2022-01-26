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

package ipwatch

import (
	"context"
	"os"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

const ipCheckInterval = time.Second * 5

type IPWatchController struct {
	NodeIP      string
	stopCh      chan struct{}
	stopEtcd    *chan<- struct{}
	etcdStopped *<-chan struct{}
}

func NewIPWatchController(cfg *config.MicroshiftConfig, stopEtcd *chan<- struct{}, etcdStopped *<-chan struct{}) *IPWatchController {
	return &IPWatchController{
		NodeIP:      cfg.NodeIP,
		stopEtcd:    stopEtcd,
		etcdStopped: etcdStopped,
	}
}

func (s *IPWatchController) Name() string { return "ipwatch-controller" }
func (s *IPWatchController) Dependencies() []string {
	return []string{}
}

func (c *IPWatchController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	ticker := time.NewTicker(ipCheckInterval)
	defer ticker.Stop()

	klog.Infof("Starting ipwatch-controller with IP address %q", c.NodeIP)

	for {
		select {
		case <-ticker.C:
			currentIP, _ := util.GetHostIP()
			if c.NodeIP != currentIP {
				klog.Warningf("IP address has changed from %q to %q, restarting MicroShift", c.NodeIP, currentIP)
				c.stopEtcdAndExit()
				return nil
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *IPWatchController) stopEtcdAndExit() {
	klog.Infof("stopEtcdAndExit c.stopCh = %+v c.etcdStopped = %+v", c.stopEtcd, c.etcdStopped)
	if c.stopEtcd == nil || c.etcdStopped == nil {
		os.Exit(0)
	}

	klog.Info("Stopping etcd service")

	close(*c.stopEtcd)

	select {
	case <-*c.etcdStopped:
		klog.Info("Etcd stopped gracefully")
		os.Exit(0)
	case <-time.Tick(30 * time.Second):
		klog.Error("Etcd stop timed out, stoping forcedfully")
		os.Exit(1)
	}
}
