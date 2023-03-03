/*
Copyright © 2021 MicroShift Contributors

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
package node

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"

	utilfeature "k8s.io/apiserver/pkg/util/feature"

	kubelet "k8s.io/kubernetes/cmd/kubelet/app"

	kubeletoptions "k8s.io/kubernetes/cmd/kubelet/app/options"
	kubeletconfig "k8s.io/kubernetes/pkg/kubelet/apis/config"
	"k8s.io/kubernetes/pkg/kubelet/kubeletconfig/configfiles"
	utilfs "k8s.io/kubernetes/pkg/util/filesystem"
)

const (
	// Kubelet component name
	componentKubelet = "kubelet"
)

var microshiftDataDir = config.GetDataDir()

type KubeletServer struct {
	kubeletflags *kubeletoptions.KubeletFlags
	kubeconfig   *kubeletconfig.KubeletConfiguration
}

func NewKubeletServer(cfg *config.MicroshiftConfig) *KubeletServer {
	s := &KubeletServer{}
	s.configure(cfg)
	return s
}

func (s *KubeletServer) Name() string           { return componentKubelet }
func (s *KubeletServer) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *KubeletServer) configure(cfg *config.MicroshiftConfig) {

	if err := s.writeConfig(cfg); err != nil {
		klog.Fatalf("Failed to write kubelet config", err)
	}
	osID, err := loadOSID()
	if err != nil {
		klog.Fatalf("Failed to read OS ID", err)
	}

	kubeletFlags := kubeletoptions.NewKubeletFlags()
	kubeletFlags.BootstrapKubeconfig = cfg.KubeConfigPath(config.Kubelet)
	kubeletFlags.KubeConfig = cfg.KubeConfigPath(config.Kubelet)
	kubeletFlags.RuntimeCgroups = "/system.slice/crio.service"
	kubeletFlags.HostnameOverride = cfg.Node.HostnameOverride
	kubeletFlags.NodeIP = cfg.Node.NodeIP
	kubeletFlags.ContainerRuntime = "remote"
	kubeletFlags.RemoteRuntimeEndpoint = "unix:///var/run/crio/crio.sock"
	kubeletFlags.NodeLabels["node-role.kubernetes.io/control-plane"] = ""
	kubeletFlags.NodeLabels["node-role.kubernetes.io/master"] = ""
	kubeletFlags.NodeLabels["node-role.kubernetes.io/worker"] = ""
	kubeletFlags.NodeLabels["node.openshift.io/os_id"] = osID

	kubeletConfig, err := loadConfigFile(microshiftDataDir + "/resources/kubelet/config/config.yaml")

	if err != nil {
		klog.Fatalf("Failed to load Kubelet Configuration", err)
	}

	s.kubeconfig = kubeletConfig
	s.kubeletflags = kubeletFlags
}

func (s *KubeletServer) writeConfig(cfg *config.MicroshiftConfig) error {
	certsDir := cryptomaterial.CertsDirectory(microshiftDataDir)
	servingCertDir := cryptomaterial.KubeletServingCertDir(certsDir)

	data := []byte(`
kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
authentication:
  x509:
    clientCAFile: ` + cryptomaterial.KubeletClientCAPath(cryptomaterial.CertsDirectory(microshiftDataDir)) + `
  anonymous:
    enabled: false
tlsCertFile: ` + cryptomaterial.ServingCertPath(servingCertDir) + `
tlsPrivateKeyFile: ` + cryptomaterial.ServingKeyPath(servingCertDir) + `
cgroupDriver: "systemd"
failSwapOn: false
volumePluginDir: ` + microshiftDataDir + `/kubelet-plugins/volume/exec
clusterDNS:
  - ` + cfg.Cluster.DNS + `
clusterDomain: cluster.local
containerLogMaxSize: 50Mi
maxPods: 250
kubeAPIQPS: 50
kubeAPIBurst: 100
cgroupsPerQOS: true
enforceNodeAllocatable: []
rotateCertificates: false  #TODO
serializeImagePulls: false
# staticPodPath: /etc/kubernetes/manifests
featureGates:
  APIPriorityAndFairness: true
  PodSecurity: true
  DownwardAPIHugePages: true
  RotateKubeletServerCertificate: false #TODO
serverTLSBootstrap: false #TODO`)

	// Load real resolv.conf in case systemd-resolved is used
	// https://github.com/coredns/coredns/blob/master/plugin/loop/README.md#troubleshooting-loops-in-kubernetes-clusters
	if _, err := os.Stat(config.DefaultSystemdResolvedFile); err == nil {
		data = append(data, fmt.Sprintf("\nresolvConf: %s\n", config.DefaultSystemdResolvedFile)...)
	}

	path := filepath.Join(microshiftDataDir, "resources", "kubelet", "config", "config.yaml")
	os.MkdirAll(filepath.Dir(path), os.FileMode(0700))
	return ioutil.WriteFile(path, data, 0644)
}

func (s *KubeletServer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {

	defer close(stopped)
	// run readiness check
	go func() {
		healthcheckStatus := util.RetryInsecureHttpsGet("http://localhost:10248/healthz")
		if healthcheckStatus != 200 {
			klog.Fatalf("", fmt.Errorf("%s failed to start", s.Name()))
		}
		klog.Infof("%s is ready", s.Name())
		close(ready)
	}()

	// construct a KubeletServer from kubeletFlags and kubeletConfig
	kubeletServer := &kubeletoptions.KubeletServer{
		KubeletFlags:         *s.kubeletflags,
		KubeletConfiguration: *s.kubeconfig,
	}

	kubeletDeps, err := kubelet.UnsecuredDependencies(kubeletServer, utilfeature.DefaultFeatureGate)
	if err != nil {
		klog.Fatalf("Error in fetching depenedencies: %v", err)
	}
	if err := kubelet.Run(ctx, kubeletServer, kubeletDeps, utilfeature.DefaultFeatureGate); err != nil {
		klog.Fatalf("Kubelet failed to start: %v", err)
	}
	return ctx.Err()
}

func loadConfigFile(name string) (*kubeletconfig.KubeletConfiguration, error) {
	const errFmt = "failed to load Kubelet config file %s, error %v"
	// compute absolute path based on current working dir
	kubeletConfigFile, err := filepath.Abs(name)
	if err != nil {
		return nil, fmt.Errorf(errFmt, name, err)
	}
	loader, err := configfiles.NewFsLoader(&utilfs.DefaultFs{}, kubeletConfigFile)
	if err != nil {
		return nil, fmt.Errorf(errFmt, name, err)
	}
	kc, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf(errFmt, name, err)
	}
	return kc, err
}

func loadOSID() (string, error) {
	readFile, err := os.Open("/etc/os-release")
	if err != nil {
		if os.IsNotExist(err) {
			return "unknown", nil
		}
		return "", err
	}
	defer readFile.Close()
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if strings.HasPrefix(line, "ID=") {
			// remove prefix and quotes
			return line[4 : len(line)-1], nil
		}
	}
	return "", fmt.Errorf("OS ID not found")
}
