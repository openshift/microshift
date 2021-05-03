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
	kubeapiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
)

func KubeAPIServer(args []string, ready chan bool) {
	command := kubeapiserver.NewAPIServerCommand()
	go func() {
		logrus.Fatalf("kube-apiserver exited: %v", command.Execute())
	}()
	ready <- true
}
