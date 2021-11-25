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
package node

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	cliflag "k8s.io/component-base/cli/flag"

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

func NewKubeletServer(cfg *config.MicroshiftConfig) *KubeletServer {
	s := &KubeletServer{}
	s.configure(cfg)
	return s
}

func (s *KubeletServer) Name() string           { return componentKubelet }
func (s *KubeletServer) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *KubeletServer) configure(cfg *config.MicroshiftConfig) error {

	if err := s.writeConfig(cfg); err != nil {
		logrus.Fatalf("Failed to write kubelet config: %v", err)
	}

	// Prepare commandline args
	args := []string{
		"--bootstrap-kubeconfig=" + cfg.DataDir + "/resources/kubelet/kubeconfig",
		"--kubeconfig=" + cfg.DataDir + "/resources/kubelet/kubeconfig",
		"--container-runtime=remote",
		"--container-runtime-endpoint=unix:///var/run/crio/crio.sock",
		"--runtime-cgroups=/system.slice/crio.service",
		"--node-ip=" + cfg.NodeIP,
		"--logtostderr=" + strconv.FormatBool(cfg.LogDir == "" || cfg.LogAlsotostderr),
		"--alsologtostderr=" + strconv.FormatBool(cfg.LogAlsotostderr),
		"--v=" + strconv.Itoa(cfg.LogVLevel),
		"--vmodule=" + cfg.LogVModule,
	}
	if cfg.LogDir != "" {
		args = append(args, "--log-file="+filepath.Join(cfg.LogDir, "kubelet.log"))
	}
	cleanFlagSet := pflag.NewFlagSet(componentKubelet, pflag.ContinueOnError)
	cleanFlagSet.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	kubeletFlags := kubeletoptions.NewKubeletFlags()
	kubeletConfig, err := loadConfigFile(cfg.DataDir + "/resources/kubelet/config/config.yaml")

	if err != nil {
		logrus.Fatalf("Failed to load Kubelet Configuration %v", err)
	}

	cmd := &cobra.Command{
		Use:          componentKubelet,
		Long:         componentKubelet,
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return nil },
	}

	// keep cleanFlagSet separate, so Cobra doesn't pollute it with the global flags
	kubeletFlags.AddFlags(cleanFlagSet)
	kubeletoptions.AddKubeletConfigFlags(cleanFlagSet, kubeletConfig)
	kubeletoptions.AddGlobalFlags(cleanFlagSet)
	cmd.Flags().AddFlagSet(cleanFlagSet)

	if err := cmd.ParseFlags(args); err != nil {
		logrus.Fatalf("%s failed to parse flags: %v", s.Name(), err)
	}
	s.kubeconfig = kubeletConfig
	s.kubeletflags = kubeletFlags

	logrus.Infof("Starting kubelet %s, args: %v", cfg.NodeIP, args)
	return nil
}

func (s *KubeletServer) writeConfig(cfg *config.MicroshiftConfig) error {
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
volumePluginDir: ` + cfg.DataDir + `/kubelet-plugins/volume/exec
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

	// Load real resolv.conf in case systemd-resolved is used
	// https://github.com/coredns/coredns/blob/master/plugin/loop/README.md#troubleshooting-loops-in-kubernetes-clusters
	if _, err := os.Stat("/run/systemd/resolve/resolv.conf"); !errors.Is(err, os.ErrNotExist) {
		data = append(data, "\nresolvConf: /run/systemd/resolve/resolv.conf"...)
	}

	path := filepath.Join(cfg.DataDir, "resources", "kubelet", "config", "config.yaml")
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}

func (s *KubeletServer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {

	defer close(stopped)
	// run readiness check
	go func() {
		healthcheckStatus := util.RetryInsecureHttpsGet("http://127.0.0.1:10248/healthz")
		if healthcheckStatus != 200 {
			logrus.Fatalf("%s failed to start", s.Name())
		}
		logrus.Infof("%s is ready", s.Name())
		close(ready)
	}()

	// construct a KubeletServer from kubeletFlags and kubeletConfig
	kubeletServer := &kubeletoptions.KubeletServer{
		KubeletFlags:         *s.kubeletflags,
		KubeletConfiguration: *s.kubeconfig,
	}

	kubeletDeps, err := kubelet.UnsecuredDependencies(kubeletServer, utilfeature.DefaultFeatureGate)
	if err != nil {
		logrus.Fatalf("Error in fetching depenedencies %v", err)
	}
	if err := kubelet.Run(ctx, kubeletServer, kubeletDeps, utilfeature.DefaultFeatureGate); err != nil {
		logrus.Fatalf("Kubelet failed to start  %v", err)
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
	loader, err := configfiles.NewFsLoader(utilfs.DefaultFs{}, kubeletConfigFile)
	if err != nil {
		return nil, fmt.Errorf(errFmt, name, err)
	}
	kc, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf(errFmt, name, err)
	}
	return kc, err
}
