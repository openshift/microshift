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

	"github.com/openshift/microshift/pkg/config"

	kubecm "k8s.io/kubernetes/cmd/kube-controller-manager/app"
)

func KubeControllerManager(cfg *config.MicroshiftConfig) {
	command := kubecm.NewControllerManagerCommand()
	args := []string{
		"--kubeconfig=" + cfg.DataDir + "/resources/kubeadmin/kubeconfig",
		"--service-account-private-key-file=" + cfg.DataDir + "/resources/kube-apiserver/secrets/service-account-key/service-account.key",
		"--allocate-node-cidrs=true",
		"--cluster-cidr=" + cfg.Cluster.ClusterCIDR,
		"--authorization-kubeconfig=" + cfg.DataDir + "/resources/kubeadmin/kubeconfig",
		"--authentication-kubeconfig=" + cfg.DataDir + "/resources/kubeadmin/kubeconfig",
		"--root-ca-file=" + cfg.DataDir + "/certs/ca-bundle/ca-bundle.crt",
		"--bind-address=127.0.0.1",
		"--secure-port=10257",
		"--leader-elect=false",
		"--use-service-account-credentials=true",
		"--cluster-signing-cert-file=" + cfg.DataDir + "/certs/ca-bundle/ca-bundle.crt",
		"--cluster-signing-key-file=" + cfg.DataDir + "/certs/ca-bundle/ca-bundle.key",
	}
	if err := command.ParseFlags(args); err != nil {
		logrus.Fatalf("failed to parse flags: %v", err)
	}
	go func() {
		command.Run(command, nil)
		logrus.Fatalf("controller-manager exited")
	}()

}
