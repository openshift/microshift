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

	// HostsPath is the path to the hosts file.
	// This file will be mounted as a volume to the coreDNS pod.
	// will be used by coreDNS hosts plugin
	// This is useful for adding custom entries to the hosts file.
	//
	// +kubebuilder:default=/etc/hosts
	// +kubebuilder:example=/etc/hosts
	HostsPath string `json:"hostsPath"`
}
