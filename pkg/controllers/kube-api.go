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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/openshift/microshift/pkg/constant"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	genericcontrollermanager "k8s.io/controller-manager/app"
	kubeapiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
)

func KubeAPIServer(args []string, ready chan bool) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", constant.KubeAPIKubeconfigPath)
	if err != nil {
		return err
	}

	versionedClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	command := kubeapiserver.NewAPIServerCommand()
	go func() {
		logrus.Fatalf("kube-apiserver exited: %v", command.Execute())
	}()
	err = genericcontrollermanager.WaitForAPIServer(versionedClient, 10*time.Second)
	if err != nil {
		logrus.Fatalf("Failed to wait for apiserver being healthy: %v", err)
	}
	ready <- true
	return nil
}
