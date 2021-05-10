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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/openshift/microshift/pkg/controllers"
	kubescheduler "k8s.io/kubernetes/cmd/kube-scheduler/app"
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
	logrus.Infof("starting kube-apiserver")
	controllers.KubeAPIServer(kubeAPIReadyCh)
	<-kubeAPIReadyCh

	logrus.Infof("starting openshift-apiserver")
	ocpAPIReadyCh := make(chan bool, 1)
	controllers.OCPAPIServer(ocpAPIReadyCh)
	<-ocpAPIReadyCh

	kubeCMReadyCh := make(chan bool, 1)
	controllers.KubeControllerManager(kubeCMReadyCh)
	<-kubeCMReadyCh

	/*
		kubeSchedulerReadyCh := make(chan bool, 1)
		kubeScheduler(args, kubeSchedulerReadyCh)
		<-kubeSchedulerReadyCh
	*/
	//TODO: cloud provider

	ocpCMReadyCh := make(chan bool, 1)
	controllers.OCPControllerManager(ocpCMReadyCh)
	<-ocpCMReadyCh
	select {}
}

func kubeScheduler(args []string, ready chan bool) {
	command := kubescheduler.NewSchedulerCommand()
	go func() {
		logrus.Fatalf("kube-scheduler exited: %v", command.Execute())
	}()
	ready <- true
}
