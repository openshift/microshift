/*
Copyright © 2021 Microshift Contributors

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
package node

import (
	"context"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"

	kubeproxy "k8s.io/kubernetes/cmd/kube-proxy/app"
)

const (
	// proxy component name
	componentKubeProxy = "kube-proxy"
)

type ProxyOptions struct {
	options *kubeproxy.Options
}

func NewKubeProxyServer(cfg *config.MicroshiftConfig) *ProxyOptions {
	s := &ProxyOptions{}
	s.configure(cfg)
	return s
}

func (s *ProxyOptions) Name() string           { return componentKubeProxy }
func (s *ProxyOptions) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *ProxyOptions) configure(cfg *config.MicroshiftConfig) error {
	if err := config.KubeProxyConfig(cfg); err != nil {
		logrus.Infof("Failed to create a new kube-proxy configuration: %v", err)
		return err
	}
	args := []string{
		"--logtostderr=" + strconv.FormatBool(cfg.LogDir == "" || cfg.LogAlsotostderr),
		"--alsologtostderr=" + strconv.FormatBool(cfg.LogAlsotostderr),
		"--v=" + strconv.Itoa(cfg.LogVLevel),
		"--vmodule=" + cfg.LogVModule,
	}
	cmd := &cobra.Command{
		Use:          componentKubeProxy,
		Long:         componentKubeProxy,
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return nil },
	}
	if cfg.LogDir != "" {
		args = append(args, "--log-file="+filepath.Join(cfg.LogDir, "kube-proxy.log"))
	}

	opts := kubeproxy.NewOptions()
	opts.ConfigFile = cfg.DataDir + "/resources/kube-proxy/config/config.yaml"
	opts.Complete()
	s.options = opts
	cmd.SetArgs(args)

	if err := cmd.ParseFlags(args); err != nil {
		logrus.Fatalf("failed to parse flags:%v", err)
	}
	logrus.Infof("starting %s, args: %v", s.Name(), args)
	return nil
}

func (s *ProxyOptions) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {

	defer close(stopped)
	// run readiness check
	go func() {
		healthcheckStatus := util.RetryInsecureHttpsGet("http://127.0.0.1:10256/healthz")
		if healthcheckStatus != 200 {
			logrus.Fatalf("%s failed to start", s.Name())
		}
		logrus.Infof("%s is ready", s.Name())
		close(ready)
	}()
	if err := s.options.Run(); err != nil {
		logrus.Fatalf("%s failed to start %v", s.Name(), err)
	}

	return ctx.Err()
}
