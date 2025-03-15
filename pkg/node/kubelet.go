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
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"

	embedded "github.com/openshift/microshift/assets"
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

type KubeletServer struct {
	kubeletflags *kubeletoptions.KubeletFlags
	kubeconfig   *kubeletconfig.KubeletConfiguration
}

func NewKubeletServer(cfg *config.Config) *KubeletServer {
	s := &KubeletServer{}
	s.configure(cfg)
	return s
}

func (s *KubeletServer) Name() string           { return componentKubelet }
func (s *KubeletServer) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *KubeletServer) configure(cfg *config.Config) {
	if err := s.writeConfig(cfg); err != nil {
		klog.Fatalf("Failed to write kubelet config %v", err)
	}
	osID, err := loadOSID()
	if err != nil {
		klog.Fatalf("Failed to read OS ID %v", err)
	}

	nodeIP := cfg.Node.NodeIP
	if len(cfg.Node.NodeIPV6) != 0 {
		nodeIP = fmt.Sprintf("%s,%s", nodeIP, cfg.Node.NodeIPV6)
	}
	kubeletFlags := kubeletoptions.NewKubeletFlags()
	kubeletFlags.BootstrapKubeconfig = cfg.KubeConfigPath(config.Kubelet)
	kubeletFlags.KubeConfig = cfg.KubeConfigPath(config.Kubelet)
	kubeletFlags.RuntimeCgroups = "/system.slice/crio.service"
	kubeletFlags.HostnameOverride = cfg.Node.HostnameOverride
	kubeletFlags.NodeIP = nodeIP
	kubeletFlags.NodeLabels["node-role.kubernetes.io/control-plane"] = ""
	kubeletFlags.NodeLabels["node-role.kubernetes.io/master"] = ""
	kubeletFlags.NodeLabels["node-role.kubernetes.io/worker"] = ""
	kubeletFlags.NodeLabels["node.openshift.io/os_id"] = osID
	kubeletFlags.NodeLabels["node.kubernetes.io/instance-type"] = "rhde"

	kubeletConfig, err := loadConfigFile(filepath.Join(config.DataDir, "/resources/kubelet/config/config.yaml"))

	if err != nil {
		klog.Fatalf("Failed to load Kubelet Configuration %v", err)
	}

	s.kubeconfig = kubeletConfig
	s.kubeletflags = kubeletFlags
}

func (s *KubeletServer) writeConfig(cfg *config.Config) error {
	data, err := s.generateConfig(cfg)
	if err != nil {
		return err
	}

	path := filepath.Join(config.DataDir, "resources", "kubelet", "config", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), os.FileMode(0700)); err != nil {
		return fmt.Errorf("failed to create dir %q: %w", path, err)
	}

	return os.WriteFile(path, data, 0400)
}

func (s *KubeletServer) generateConfig(cfg *config.Config) ([]byte, error) {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	servingCertDir := cryptomaterial.KubeletServingCertDir(certsDir)

	tplData, err := embedded.Asset("core/kubelet.yaml")
	if err != nil {
		return nil, fmt.Errorf("loading kubelet.yaml asset failed: %w", err)
	}

	tpl, err := template.New("").Option("missingkey=error").Parse(string(tplData))
	if err != nil {
		return nil, fmt.Errorf("creating a template for kubelet config failed: %w", err)
	}

	resolvConf := ""
	// Load real resolv.conf in case systemd-resolved is used
	// https://github.com/coredns/coredns/blob/master/plugin/loop/README.md#troubleshooting-loops-in-kubernetes-clusters
	if _, err := os.Stat(config.DefaultSystemdResolvedFile); err == nil {
		resolvConf = config.DefaultSystemdResolvedFile
	}

	userProvidedConfig := ""
	if cfg.Kubelet != nil {
		b, err := yaml.Marshal(cfg.Kubelet)
		if err != nil {
			return nil, fmt.Errorf("failed to re-marshal user provided kubelet config: %w", err)
		}
		userProvidedConfig = string(b)
	}

	tplParams := map[string]string{
		"clientCAFile":       cryptomaterial.KubeletClientCAPath(cryptomaterial.CertsDirectory(config.DataDir)),
		"tlsCertFile":        cryptomaterial.ServingCertPath(servingCertDir),
		"tlsPrivateKeyFile":  cryptomaterial.ServingKeyPath(servingCertDir),
		"volumePluginDir":    config.DataDir + "/kubelet-plugins/volume/exec",
		"clusterDNSIP":       cfg.Network.DNS,
		"resolvConf":         resolvConf,
		"tlsCipherSuites":    strings.Join(cfg.ApiServer.TLS.CipherSuites, ","),
		"tlsMinVersion":      cfg.ApiServer.TLS.MinVersion,
		"userProvidedConfig": userProvidedConfig,
	}

	var data bytes.Buffer
	if err := tpl.Execute(&data, tplParams); err != nil {
		return nil, fmt.Errorf("templating kubelet config failed: %w", err)
	}

	return data.Bytes(), nil
}

func (s *KubeletServer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	// construct a KubeletServer from kubeletFlags and kubeletConfig
	kubeletServer := &kubeletoptions.KubeletServer{
		KubeletFlags:         *s.kubeletflags,
		KubeletConfiguration: *s.kubeconfig,
	}

	kubeletDeps, err := kubelet.UnsecuredDependencies(kubeletServer, utilfeature.DefaultFeatureGate)
	if err != nil {
		return fmt.Errorf("error fetching dependencies: %w", err)
	}

	errc := make(chan error)

	// Run healthcheck probe and kubelet in parallel.
	// No matter which ends first - if it ends with an error,
	// it'll cause ServiceManager to trigger graceful shutdown.

	// run readiness check
	go func() {
		// This endpoint does not use TLS, but reusing the same function without verification.
		healthcheckStatus := util.RetryInsecureGet(ctx, "http://localhost:10248/healthz")
		if healthcheckStatus != 200 {
			e := fmt.Errorf("%s failed to start", s.Name())
			klog.Error(e)
			errc <- e
			return
		}
		klog.Infof("%s is ready", s.Name())
		close(ready)
	}()

	panicChannel := make(chan any, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicChannel <- r
			}
		}()
		errc <- kubelet.Run(ctx, kubeletServer, kubeletDeps, utilfeature.DefaultFeatureGate)
	}()

	select {
	case err := <-errc:
		return err
	case perr := <-panicChannel:
		panic(perr)
	}
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
