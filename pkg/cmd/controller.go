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

	"github.com/openshift/microshift/pkg/components"
	"github.com/openshift/microshift/pkg/controllers"
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
	if err := startControllerOnly(); err != nil {
		return err
	}
	select {}
}

func startControllerOnly() error {
	etcdReadyCh := make(chan bool, 1)
	if err := controllers.StartEtcd(etcdReadyCh); err != nil {
		return err
	}
	<-etcdReadyCh

	logrus.Infof("starting kube-apiserver")
	controllers.KubeAPIServer()

	logrus.Infof("starting kube-controller-manager")
	controllers.KubeControllerManager()

	logrus.Infof("starting kube-scheduler")
	controllers.KubeScheduler()

	if err := controllers.PrepareOCP(); err != nil {
		return err
	}

	logrus.Infof("starting openshift-apiserver")
	controllers.OCPAPIServer()

	//TODO: cloud provider
	controllers.OCPControllerManager()

	if err := controllers.StartOCPAPIComponents(); err != nil {
		return err
	}

	if err := components.StartComponents(); err != nil {
		return err
	}
	return nil
}
