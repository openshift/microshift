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

func KubeProxyConfig(path string) error {
	data := []byte(`
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
clientConnection:
   kubeconfig: ` + constant.AdminKubeconfigPath + `
hostnameOverride: 127.0.0.1
mode:
featureGates:
   AllAlpha: false`)
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}
