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
	"github.com/openshift/microshift/pkg/components"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/controllers"
	"github.com/sirupsen/logrus"
)

func startControllerOnly(cfg *config.MicroshiftConfig) error {
	etcdReadyCh := make(chan bool, 1)
	if err := controllers.StartEtcd(cfg, etcdReadyCh); err != nil {
		return err
	}
	<-etcdReadyCh

	logrus.Infof("starting kube-apiserver")
	controllers.KubeAPIServer(cfg)

	logrus.Infof("starting kube-controller-manager")
	controllers.KubeControllerManager(cfg)

	logrus.Infof("starting kube-scheduler")
	controllers.KubeScheduler(cfg)

	if err := controllers.PrepareOCP(cfg); err != nil {
		return err
	}

	logrus.Infof("starting openshift-apiserver")
	controllers.OCPAPIServer(cfg)

	//TODO: cloud provider
	controllers.OCPControllerManager(cfg)

	if err := controllers.StartOCPAPIComponents(cfg); err != nil {
		return err
	}

	if err := components.StartComponents(cfg); err != nil {
		return err
	}
	return nil
}
