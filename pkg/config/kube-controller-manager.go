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
	"os"
	"path/filepath"
	"text/template"
)

// KubeControllerManagerConfig creates a config for kube-controller-manager in option --openshift-config
func KubeControllerManagerConfig(path string) error {
	configTemplate := template.Must(template.New("config").Parse(`
apiVersion: kubecontrolplane.config.openshift.io/v1
kind: KubeControllerManagerConfig  
extendedArguments:
  enable-dynamic-provisioning:
  - "true"
  allocate-node-cidrs:
  - "false"
  configure-cloud-routes:
  - "false"
  use-service-account-credentials:
  - "true"
  flex-volume-plugin-dir:
  - "/etc/kubernetes/kubelet-plugins/volume/exec" # created by machine-config-operator, owned by storage team/hekumar@redhat.com
  pv-recycler-pod-template-filepath-nfs: # owned by storage team/fbertina@redhat.com
  - "/etc/kubernetes/ushift-resources/configmaps/recycler-config/recycler-pod.yaml"
  pv-recycler-pod-template-filepath-hostpath: # owned by storage team/fbertina@redhat.com
  - "/etc/kubernetes/ushift-resources/configmaps/recycler-config/recycler-pod.yaml"
  leader-elect:
  - "true"
  leader-elect-retry-period:
  - "3s"
  leader-elect-resource-lock:
  - "configmaps"
  controllers:
  - "*"
  - "-ttl" # TODO: this is excluded in kube-core, but not in #21092
  - "-bootstrapsigner"
  - "-tokencleaner"
  experimental-cluster-signing-duration:
  - "720h"
  secure-port:
  - "10257"
  port:
  - "0"
  cert-dir:
  - "/var/run/kubernetes"
  kube-api-qps:
  - "150" # this is a historical values
  kube-api-burst:
  - "300" # this is a historical values
  `))
	data := struct {
		ClientCACert, KubeConfig, ServingCert, ServingKey, ServingClientCert,
		IngressDomain, EtcdUrl, EtcdCert, EtcdKey, EtcdCA string
	}{
		/*
			ClientCACert:      ,
			KubeConfig:        ,
			ServingCert:       ,
			ServingKey:        ,
			ServingClientCert: ,
			IngressDomain:     ,
			EtcdUrl:           ,
			EtcdCA:            ,
			EtcdCert:          ,
			EtcdKey:           ,
		*/
	}
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	return configTemplate.Execute(output, &data)
}
