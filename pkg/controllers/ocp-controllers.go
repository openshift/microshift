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
		"--requestheader-client-ca-file=" + cfg.DataDir + "/certs/ca-bundle/ca-bundle.crt",
		"--requestheader-allowed-names=kube-apiserver-proxy,system:kube-apiserver-proxy,system:openshift-aggregator",
		"--requestheader-username-headers=X-Remote-User",
		"--requestheader-group-headers=X-Remote-Group",
		"--requestheader-extra-headers-prefix=X-Remote-Extra-",
		"--client-ca-file=" + cfg.DataDir + "/certs/ca-bundle/ca-bundle.crt",
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
	if err := createAPIHeadlessSvc(cfg); err != nil {
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

func OCPControllerManager(cfg *config.MicroshiftConfig) {
	command := newOpenShiftControllerManagerCommand()
	args := []string{
		"--config=" + cfg.DataDir + "/resources/openshift-controller-manager/config/config.yaml",
	}
	startArgs := append(args, "start")
	command.SetArgs(startArgs)

	go func() {
		logrus.Fatalf("ocp controller-manager exited: %v", command.Execute())
	}()
}
