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
package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/constant"
)

// KubeSchedulerConfig creates a config for kube-scheduler in option --config
func KubeSchedulerConfig(path string) error {
	data := []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta1
kind: KubeSchedulerConfiguration
clientConnection:
  kubeconfig:"` + constant.KubeSchedulerKubeconfigPath + `
leaderElection:
  leaderElect: true`)

	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}
