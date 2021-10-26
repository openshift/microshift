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
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	openshift_apiserver "github.com/openshift/openshift-apiserver/pkg/cmd/openshift-apiserver"
	openshift_controller_manager "github.com/openshift/openshift-controller-manager/pkg/cmd/openshift-controller-manager"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
)

func newOpenshiftApiServerCommand(stopCh <-chan struct{}) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openshift-apiserver",
		Short: "Command for the OpenShift API Server",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	start := openshift_apiserver.NewOpenShiftAPIServerCommand("start", os.Stdout, os.Stderr, stopCh)
	cmd.AddCommand(start)

	return cmd
}

func OCPAPIServer(cfg *config.MicroshiftConfig) error {
	stopCh := make(chan struct{})
	command := newOpenshiftApiServerCommand(stopCh)
	args := []string{
		"start",
		"--config=" + cfg.DataDir + "/resources/openshift-apiserver/config/config.yaml",
		"--authorization-kubeconfig=" + cfg.DataDir + "/resources/kubeadmin/kubeconfig",
		"--authentication-kubeconfig=" + cfg.DataDir + "/resources/kubeadmin/kubeconfig",
		"--logtostderr=" + strconv.FormatBool(cfg.LogDir == "" || cfg.LogAlsotostderr),
		"--alsologtostderr=" + strconv.FormatBool(cfg.LogAlsotostderr),
		"--v=" + strconv.Itoa(cfg.LogVLevel),
		"--vmodule=" + cfg.LogVModule,
	}
	if cfg.LogDir != "" {
		args = append(args, "--log-dir="+cfg.LogDir)
	}

	command.SetArgs(args)
	logrus.Infof("starting openshift-apiserver, args: %v", args)
	go func() {
		logrus.Fatalf("ocp apiserver exited: %v", command.Execute())
	}()

	// ocp api service registration
	if err := createAPIHeadlessSvc(cfg, "openshift-apiserver", 8444); err != nil {
		logrus.Warningf("failed to apply headless svc %v", err)
		return err
	}
	if err := createAPIHeadlessSvc(cfg, "openshift-oauth-apiserver", 8443); err != nil {
		logrus.Warningf("failed to apply headless svc %v", err)
		return err
	}
	if err := createAPIRegistration(cfg); err != nil {
		logrus.Warningf("failed to register api %v", err)
		return err
	}

	// probe ocp api services
	restConfig, err := clientcmd.BuildConfigFromFlags("", cfg.DataDir+"/resources/kubeadmin/kubeconfig")
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	err = waitForOCPAPIServer(client, 10*time.Second)
	if err != nil {
		logrus.Warningf("Failed to wait for ocp apiserver: %v", err)
		return nil
	}

	logrus.Info("ocp apiserver is ready")

	return nil
}

func waitForOCPAPIServer(client kubernetes.Interface, timeout time.Duration) error {
	var lastErr error

	err := wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		for _, apiSvc := range []string{
			"apps.openshift.io",
			"authorization.openshift.io",
			"build.openshift.io",
			"image.openshift.io",
			"project.openshift.io",
			"quota.openshift.io",
			"route.openshift.io",
			"security.openshift.io",
			"template.openshift.io", //TODO missing templateinstances
		} {
			status := 0
			url := "/apis/" + apiSvc
			result := client.Discovery().RESTClient().Get().AbsPath(url).Do(context.TODO()).StatusCode(&status)
			if result.Error() != nil {
				lastErr = fmt.Errorf("failed to get apiserver %s status: %v", url, result.Error())
				return false, nil
			}
			if status != http.StatusOK {
				content, _ := result.Raw()
				lastErr = fmt.Errorf("APIServer isn't available: %v", string(content))
				logrus.Warningf("APIServer isn't available yet: %v. Waiting a little while.", string(content))
				return false, nil
			}
		}
		return true, nil
	})

	if err != nil {
		return fmt.Errorf("%v: %v", err, lastErr)
	}

	return nil
}

func newOpenShiftControllerManagerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openshift-controller-manager",
		Short: "Command for the OpenShift Controllers",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	start := openshift_controller_manager.NewOpenShiftControllerManagerCommand("start", os.Stdout, os.Stderr)
	cmd.AddCommand(start)
	return cmd
}

type OCPControllerManager struct {
	ConfigFilePath string
	Output         io.Writer
}

const (
	// OCPControllerManager component name
	componentOCM = "ocp-controller-manager"
)

func NewOpenShiftControllerManager(cfg *config.MicroshiftConfig) *OCPControllerManager {
	s := &OCPControllerManager{}
	s.configure(cfg)
	return s
}

func (s *OCPControllerManager) Name() string           { return componentOCM }
func (s *OCPControllerManager) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *OCPControllerManager) configure(cfg *config.MicroshiftConfig) error {
	var configFilePath = cfg.DataDir + "/resources/openshift-controller-manager/config/config.yaml"

	if err := config.OpenShiftControllerManagerConfig(cfg); err != nil {
		logrus.Infof("Failed to create a new ocp-controller-manager configuration: %v", err)
		return err
	}
	args := []string{
		"--config=" + configFilePath,
	}

	options := openshift_controller_manager.OpenShiftControllerManager{Output: os.Stdout}
	options.ConfigFilePath = configFilePath

	cmd := &cobra.Command{
		Use:          componentOCM,
		Long:         componentOCM,
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return nil },
	}

	flags := cmd.Flags()
	cmd.SetArgs(args)
	flags.StringVar(&options.ConfigFilePath, "config", options.ConfigFilePath, "Location of the master configuration file to run from.")
	cmd.MarkFlagFilename("config", "yaml", "yml")
	cmd.MarkFlagRequired("config")

	s.ConfigFilePath = options.ConfigFilePath
	s.Output = options.Output

	return nil
}

func (s *OCPControllerManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	// run readiness check
	go func() {
		healthcheckStatus := util.RetryTCPConnection("127.0.0.1", "8445")
		if !healthcheckStatus {
			logrus.Fatalf("%s failed to start", s.Name())
		}
		logrus.Infof("%s is ready", s.Name())
		close(ready)
	}()
	options := openshift_controller_manager.OpenShiftControllerManager{Output: os.Stdout}
	options.ConfigFilePath = s.ConfigFilePath
	if err := options.StartControllerManager(); err != nil {
		logrus.Fatalf("Failed to start ocp-controller-manager %v", err)
	}
	return ctx.Err()
}
