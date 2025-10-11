package config

import (
	"os"
	"path/filepath"
)

// KubeConfigID identifies the different kubeconfigs managed in the DataDir
type KubeConfigID string

const (
	KubeAdmin               KubeConfigID = "kubeadmin"
	KubeControllerManager   KubeConfigID = "kube-controller-manager"
	KubeScheduler           KubeConfigID = "kube-scheduler"
	Kubelet                 KubeConfigID = "kubelet"
	ClusterPolicyController KubeConfigID = "cluster-policy-controller"
	RouteControllerManager  KubeConfigID = "route-controller-manager"
	ObservabilityClient     KubeConfigID = "observability-client"
)

// KubeConfigPath returns the path to the specified kubeconfig file.
func (cfg *Config) KubeConfigPath(id KubeConfigID) string {
	return filepath.Join(DataDir, "resources", string(id), "kubeconfig")
}

func (cfg *Config) KubeConfigAdminPath(id string) string {
	return filepath.Join(cfg.KubeConfigRootAdminPath(), id, "kubeconfig")
}

func (cfg *Config) KubeConfigRootAdminPath() string {
	return filepath.Join(DataDir, "resources", string(KubeAdmin))
}

func (cfg *Config) BootstrapKubeConfigPath() string {
	return filepath.Join(DataDir, "bootstrap", "kubeconfig")
}

func (cfg *Config) BootstrapKubeConfigExists() bool {
	_, err := os.Stat(cfg.BootstrapKubeConfigPath())
	return err == nil
}
