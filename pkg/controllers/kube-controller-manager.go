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
	"github.com/sirupsen/logrus"

	"github.com/openshift/microshift/pkg/constant"

	kubecm "k8s.io/kubernetes/cmd/kube-controller-manager/app"
)

func KubeControllerManager(ready chan bool) {
	command := kubecm.NewControllerManagerCommand()
	args := []string{
		"--kubeconfig=" + constant.AdminKubeconfigPath, //KubeControllerManagerKubeconfigPath,
		"--service-account-private-key-file=/etc/kubernetes/ushift-resources/kube-apiserver/secrets/service-account-key/service-account.key",
		"--allocate-node-cidrs=false",
		//"--allocate-node-cidrs=true",
		//"--cluster-cidr=",
		"--authorization-kubeconfig=" + constant.AdminKubeconfigPath,
		"--authentication-kubeconfig=" + constant.AdminKubeconfigPath,
		"--root-ca-file=/etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt",
		"--bind-address=127.0.0.1",
		"--secure-port=10257",
		"--leader-elect=false",
		"--use-service-account-credentials=true",
		"--cluster-signing-cert-file=/etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt",
		"cluster-signing-key-file=/etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.key",
	}
	if err := command.ParseFlags(args); err != nil {
		logrus.Fatalf("failed to parse flags:%v", err)
	}
	go func() {
		command.Run(command, nil)
		logrus.Fatalf("controller-manager exited")
	}()

	ready <- true
}
