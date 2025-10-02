package config

type DNS struct {
	// baseDomain is the base domain of the cluster. All managed DNS records will
	// be sub-domains of this base.
	//
	// For example, given the base domain `example.com`, router exposed
	// domains will be formed as `*.apps.example.com` by default,
	// and API service will have a DNS entry for `api.example.com`,
	// as well as "api-int.example.com" for internal k8s API access.
	//
	// Once set, this field cannot be changed.
	// +kubebuilder:default=example.com
	// +kubebuilder:example=microshift.example.com
	BaseDomain string `json:"baseDomain"`

	// HostsWatcher configures the hosts file watcher service.
	// When configured, this service monitors the hosts file for changes
	// and creates/updates ConfigMaps in specified namespaces.
	// If not specified, the hosts watcher service is disabled.
	HostsWatcher *HostsWatcher `json:"hostsWatcher,omitempty"`
}

type HostsWatcher struct {
	// HostsPath is the path to the hosts file to monitor.
	// If not specified, defaults to "/etc/hosts".
	// +kubebuilder:default="/etc/hosts"
	// +kubebuilder:example="/etc/hosts"
	HostsPath string `json:"hostsPath"`

	// TargetNamespaces specifies the list of namespaces where ConfigMaps
	// should be created when the hosts file changes.
	// If empty, ConfigMaps will be created in the "default" namespace.
	// +kubebuilder:default={"default"}
	// +kubebuilder:example={"default","kube-system","openshift-config"}
	TargetNamespaces []string `json:"targetNamespaces"`

	// ConfigMapName specifies the name of the ConfigMap to create/update.
	// +kubebuilder:default="hosts-file"
	// +kubebuilder:example="hosts-file"
	ConfigMapName string `json:"configMapName"`
}
