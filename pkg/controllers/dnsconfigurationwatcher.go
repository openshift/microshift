package controllers

import (
	"github.com/fsnotify/fsnotify"
	"github.com/openshift/microshift/pkg/config"
)

type DNSConfigurationWatcherManager struct {
	fileWatcher
}

func NewDNSConfigurationWatcherManager(cfg *config.Config) *DNSConfigurationWatcherManager {
	return &DNSConfigurationWatcherManager{fileWatcher{cfg: fileWatcherConfig{
		serviceName:        "dns-configuration-watcher-manager",
		dependencies:       []string{"infrastructure-services-manager"},
		file:               cfg.DNS.ConfigFile,
		kubeconfig:         cfg.KubeConfigPath(config.KubeAdmin),
		enabled:            cfg.DNS.ConfigFile != "",
		configMapNamespace: "openshift-dns",
		configMapName:      "dns-default",
		configMapDataKey:   "Corefile",
		labels: map[string]string{
			"dns.operator.openshift.io/owning-dns": "default",
		},
		annotations: map[string]string{
			"microshift.io/dns-config-file": cfg.DNS.ConfigFile,
		},
		eventMask:        fsnotify.Write | fsnotify.Create | fsnotify.Rename | fsnotify.Remove,
		reAddOnCreate:    true,
		mergeAnnotations: true,
		deleteOnDisable:  false,
	}}}
}
