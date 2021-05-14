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

	//"text/template"

	"github.com/openshift/microshift/pkg/constant"
)

func KubeletConfig(path string) error {
	data := []byte(`
kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
authentication:
  x509:
    clientCAFile: /etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt
  anonymous:
    enabled: false
cgroupDriver: "systemd"
cgroupRoot: /
failSwapOn: false
clusterDNS:
  - ` + constant.ClusterDNS + `
clusterDomain: ` + constant.DomainName + `
containerLogMaxSize: 50Mi
maxPods: 250
kubeAPIQPS: 50
kubeAPIBurst: 100
rotateCertificates: false  #TODO
serializeImagePulls: false
# staticPodPath: /etc/kubernetes/manifests
systemCgroups: /system.slice
featureGates:
  APIPriorityAndFairness: true
  LegacyNodeRoleBehavior: false
  # Will be removed in future openshift/api update https://github.com/openshift/api/commit/c8c8f6d0f4a8ac4ff4ad7d1a84b27e1aa7ebf9b4
  RemoveSelfLink: false
  NodeDisruptionExclusion: true
  RotateKubeletServerCertificate: false #TODO
  SCTPSupport: true
  ServiceNodeExclusion: true
  SupportPodPidsLimit: true
serverTLSBootstrap: false #TODO`)
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}

func OpenShiftSDNConfig(path string) error {
	data := []byte(`
apiVersion: kubeproxy.config.k8s.io/v1alpha1
bindAddress: 0.0.0.0
bindAddressHardFail: false
clientConnection:
  acceptContentTypes: ""
  burst: 0
  contentType: ""
  kubeconfig: ""
  qps: 0
clusterCIDR: - ` + constant.ClusterCIDR + `
configSyncPeriod: 0s
conntrack:
  maxPerCore: null
  min: null
  tcpCloseWaitTimeout: null
  tcpEstablishedTimeout: null
detectLocalMode: ""
enableProfiling: false
featureGates:
  EndpointSlice: false
  EndpointSliceProxying: false`)
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}
