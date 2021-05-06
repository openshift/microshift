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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/openshift/microshift/pkg/controllers"

	genericapiserver "k8s.io/apiserver/pkg/server"

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
	etcdReadyCh := make(chan bool, 1)
	if err := controllers.StartEtcd(etcdReadyCh); err != nil {
		return err
	}
	<-etcdReadyCh
	kubeAPIReadyCh := make(chan bool, 1)
	controllers.KubeAPIServer(args, kubeAPIReadyCh)
	<-kubeAPIReadyCh
	kubeCMReadyCh := make(chan bool, 1)
	kubeControllerManager(args, kubeCMReadyCh)
	<-kubeCMReadyCh
	kubeSchedulerReadyCh := make(chan bool, 1)
	kubeScheduler(args, kubeSchedulerReadyCh)
	<-kubeSchedulerReadyCh
	//TODO: cloud provider

	ocpAPIReadyCh := make(chan bool, 1)
	ocpAPIServer(args, ocpAPIReadyCh)
	<-ocpAPIReadyCh
	ocpCMReadyCh := make(chan bool, 1)
	ocpControllerManager(args, ocpCMReadyCh)
	<-ocpCMReadyCh
	select {}
}

func kubeControllerManager(args []string, ready chan bool) {
	command := kubecm.NewControllerManagerCommand()
	go func() {
		logrus.Fatalf("controller-manager exited: %v", command.Execute())
	}()
	ready <- true
}

func kubeScheduler(args []string, ready chan bool) {
	command := kubescheduler.NewSchedulerCommand()
	go func() {
		logrus.Fatalf("kube-scheduler exited: %v", command.Execute())
	}()
	ready <- true
}

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
func ocpAPIServer(args []string, ready chan bool) {
	stopCh := genericapiserver.SetupSignalHandler(false)
	command := newOpenshiftApiServerCommand(stopCh)
	startArgs := append(args, "start")
	command.SetArgs(startArgs)
	go func() {
		logrus.Fatalf("ocp apiserver exited: %v", command.Execute())
	}()
	ready <- true
}

func newOpenShiftControllerManagerCommand(stopCh <-chan struct{}) *cobra.Command {
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

func ocpControllerManager(args []string, ready chan bool) {
	stopCh := genericapiserver.SetupSignalHandler(false)
	command := newOpenShiftControllerManagerCommand(stopCh)
	startArgs := append(args, "start")
	command.SetArgs(startArgs)

	go func() {
		logrus.Fatalf("ocp controller-manager exited: %v", command.Execute())
	}()
	ready <- true
}
