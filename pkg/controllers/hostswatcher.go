package controllers

import (
	"github.com/fsnotify/fsnotify"
	"github.com/openshift/microshift/pkg/config"
)

type HostsWatcherManager struct {
	fileWatcher
}

func NewHostsWatcherManager(cfg *config.Config) *HostsWatcherManager {
	return &HostsWatcherManager{fileWatcher{cfg: fileWatcherConfig{
		serviceName:        "hosts-watcher-manager",
		dependencies:       []string{"infrastructure-services-manager"},
		file:               cfg.DNS.Hosts.File,
		kubeconfig:         cfg.KubeConfigPath(config.KubeAdmin),
		enabled:            cfg.DNS.Hosts.Status == config.HostsStatusEnabled,
		configMapNamespace: "openshift-dns",
		configMapName:      "hosts-file",
		configMapDataKey:   "hosts",
		labels: map[string]string{
			"app.kubernetes.io/name":          "microshift-hosts-watcher",
			"app.kubernetes.io/component":     "hosts-file-sync",
			"app.kubernetes.io/managed-by":    "microshift",
			"microshift.io/access-restricted": "coredns-only",
		},
		annotations: map[string]string{
			"microshift.io/hosts-file-path": cfg.DNS.Hosts.File,
		},
		eventMask:        fsnotify.Write | fsnotify.Create,
		reAddOnCreate:    false,
		mergeAnnotations: false,
		deleteOnDisable:  true,
	}}}
}
