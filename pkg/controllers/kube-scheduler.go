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
package controllers

import (
	"context"
	"path/filepath"
	"strconv"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	genericcontrollermanager "k8s.io/controller-manager/app"
	kubescheduler "k8s.io/kubernetes/cmd/kube-scheduler/app"
	schedulerOptions "k8s.io/kubernetes/cmd/kube-scheduler/app/options"
)

const (
	kubeSchedulerStartupTimeout = 60
)

type KubeScheduler struct {
	options    *schedulerOptions.Options
	kubeconfig string
}

func NewKubeScheduler(cfg *config.MicroshiftConfig) *KubeScheduler {
	s := &KubeScheduler{}
	s.configure(cfg)
	return s
}

func (s *KubeScheduler) Name() string           { return "kube-scheduler" }
func (s *KubeScheduler) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *KubeScheduler) configure(cfg *config.MicroshiftConfig) {
	if err := config.KubeSchedulerConfig(cfg); err != nil {
		return
	}

	opts, err := schedulerOptions.NewOptions()
	if err != nil {
		logrus.Fatalf("initialization error command options: %v", err)
	}

	args := []string{
		"--config=" + cfg.DataDir + "/resources/kube-scheduler/config/config.yaml",
		"--logtostderr=" + strconv.FormatBool(cfg.LogDir == "" || cfg.LogAlsotostderr),
		"--alsologtostderr=" + strconv.FormatBool(cfg.LogAlsotostderr),
		"--v=" + strconv.Itoa(cfg.LogVLevel),
		"--vmodule=" + cfg.LogVModule,
	}

	if cfg.LogDir != "" {
		args = append(args, "--log-file="+filepath.Join(cfg.LogDir, "kube-scheduler.log"))
	}

	cmd := &cobra.Command{
		Use:          "kube-scheduler",
		Long:         `kube-scheduler`,
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return nil },
	}

	namedFlagSets := opts.Flags()
	fs := cmd.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}
	if err := cmd.ParseFlags(args); err != nil {
		logrus.Fatalf("%s failed to parse flags: %v", s.Name(), err)
	}

	s.options = opts
	s.kubeconfig = filepath.Join(cfg.DataDir, "resources", "kubeadmin", "kubeconfig")

}

func (s *KubeScheduler) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	// run readiness check
	go func() {
		restConfig, err := clientcmd.BuildConfigFromFlags("", s.kubeconfig)
		if err != nil {
			logrus.Warningf("%s readiness check: %v", s.Name(), err)
			return
		}

		versionedClient, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			logrus.Warningf("%s readiness check: %v", s.Name(), err)
			return
		}

		if genericcontrollermanager.WaitForAPIServer(versionedClient, kubeSchedulerStartupTimeout*time.Second) != nil {
			logrus.Warningf("%s readiness check timed out: %v", s.Name(), err)
			return
		}

		logrus.Infof("%s is ready", s.Name())
		close(ready)
	}()

	cc, sched, err := kubescheduler.Setup(ctx, s.options)
	if err != nil {
		return err
	}

	if err := kubescheduler.Run(ctx, cc, sched); err != nil {
		return err
	}

	return ctx.Err()
}
