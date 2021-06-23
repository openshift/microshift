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
)

func KubeletConfig(cfg *MicroshiftConfig) error {
	data := []byte(`
kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
authentication:
  x509:
    clientCAFile: ` + cfg.DataDir + `/certs/ca-bundle/ca-bundle.crt
  anonymous:
    enabled: false
tlsCertFile: ` + cfg.DataDir + `/resources/kubelet/secrets/kubelet-client/tls.crt
tlsPrivateKeyFile: ` + cfg.DataDir + `/resources/kubelet/secrets/kubelet-client/tls.key
cgroupDriver: "systemd"
cgroupRoot: /
failSwapOn: false
clusterDNS:
  - ` + cfg.Cluster.DNS + `
clusterDomain: ` + cfg.Cluster.Domain + `
containerLogMaxSize: 50Mi
maxPods: 250
kubeAPIQPS: 50
kubeAPIBurst: 100
cgroupsPerQOS: false
enforceNodeAllocatable: []
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
	os.MkdirAll(filepath.Dir(cfg.DataDir+"/resources/kubelet/config/config.yaml"), os.FileMode(0755))
	return ioutil.WriteFile(cfg.DataDir+"/resources/kubelet/config/config.yaml", data, 0644)
}

func OpenShiftSDNConfig(cfg *MicroshiftConfig) error {
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
clusterCIDR: - ` + cfg.Cluster.ClusterCIDR + `
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
	os.MkdirAll(filepath.Dir(cfg.DataDir+"/resources/openshift-sdn/config/config.yaml"), os.FileMode(0755))
	return ioutil.WriteFile(cfg.DataDir+"/resources/openshift-sdn/config/config.yaml", data, 0644)
}

func KubeProxyConfig(cfg *MicroshiftConfig) error {
	data := []byte(`
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
clientConnection:
  kubeconfig: ` + cfg.DataDir + `/resources/kube-proxy/kubeconfig
hostnameOverride: ` + cfg.HostName + `
clusterCIDR: ` + cfg.Cluster.ClusterCIDR + `
mode: "iptables"
iptables:
  masqueradeAll: true
conntrack:
  maxPerCore: 0
featureGates:
   AllAlpha: false`)
	os.MkdirAll(filepath.Dir(cfg.DataDir+"/resources/kube-proxy/config/config.yaml"), os.FileMode(0755))
	return ioutil.WriteFile(cfg.DataDir+"/resources/kube-proxy/config/config.yaml", data, 0644)
}
