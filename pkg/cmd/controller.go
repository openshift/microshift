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
package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	etcd "go.etcd.io/etcd/embed"

	kubeapiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
	kubecm "k8s.io/kubernetes/cmd/kube-controller-manager/app"
	kubescheduler "k8s.io/kubernetes/cmd/kube-scheduler/app"

	openshift_apiserver "github.com/openshift/openshift-apiserver/pkg/cmd/openshift-apiserver"
	openshift_controller_manager "github.com/openshift/openshift-controller-manager/pkg/cmd/openshift-controller-manager"
)

var ControllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "openshift controller",
	Long:  `openshift controller`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return startController(args)
	},
}

func startController(args []string) error {
	etcd(args)

	kubeAPIServer(args)
	kubeControllerManager(args)
	kubeScheduler(args)
	//TODO: cloud provider

	ocpAPIServer(args)
	ocpControllerManager(args)
	return nil
}

func etcd(args []string) {
	e, err := etcd.StartEtcd(cfg)
	if err != nil {
		logrus.Fatalf("etcd failed to start %v", err)
	}
	defer e.Close()
	select {
	case <-e.Server.ReadyNotify():
		logrus.Info("Server is ready!")
	}
	logrus.Fatalf("etcd exited: %v", e.Err())
}

func kubeAPIServer(args []string) {
	command := kubeapiserver.NewAPIServerCommand()
	go func() {
		logrus.Fatalf("kube-apiserver exited: %v", command.Execute())
	}()
}

func kubeControllerManager(args []string) {
	command := kubecm.NewControllerManagerCommand()
	go func() {
		logrus.Fatalf("controller-manager exited: %v", command.Execute())
	}()
}

func kubeScheduler(args []string) {
	command := kubescheduler.NewSchedulerCommand()
	go func() {
		logrus.Fatalf("kube-scheduler exited: %v", command.Execute())
	}()
}

func newOpenshiftApiServerCommand() *cobra.Command {
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
func ocpAPIServer(args []string) {
	stopCh := genericapiserver.SetupSignalHandler()
	command := newOpenshiftApiServerCommand(stopCh)
	startArgs := append(args, "start")
	command.SetArgs(startArgs)
	go func() {
		logrus.Fatalf("ocp apiserver exited: %v", command.Execute())
	}()
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

func ocpControllerManager(args []string) {
	stopCh := genericapiserver.SetupSignalHandler()
	command := newOpenShiftControllerManagerCommand(stopCh)
	startArgs := append(args, "start")
	command.SetArgs(startArgs)

	go func() {
		logrus.Fatalf("ocp controller-manager exited: %v", command.Execute())
	}()
}
