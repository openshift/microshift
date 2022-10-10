package api

type Config struct {
	Components Components `json:components`
}

type Components struct {
	Network Network `json:network`
	DNS     DNS     `json:dns`
}

type Network struct {
	// IP address pool to use for pod IPs.
	// This field is immutable after installation.
	ClusterNetwork []ClusterNetworkEntry `json:"clusterNetwork"`

	// IP address pool for services.
	// Currently, we only support a single entry here.
	// This field is immutable after installation.
	ServiceNetwork []string `json:"serviceNetwork"`

	// The port range allowed for Services of type NodePort.
	// If not specified, the default of 30000-32767 will be used.
	// Such Services without a NodePort specified will have one
	// automatically allocated from this range.
	// This parameter can be updated after the cluster is
	// installed.
	// +kubebuilder:validation:Pattern=`^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])-([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`
	ServiceNodePortRange string `json:"serviceNodePortRange,omitempty"`
}

type ClusterNetworkEntry struct {
	// The complete block for pod IPs.
	CIDR string `json:"cidr"`
}

type DNS struct {
	// baseDomain is the base domain of the cluster. All managed DNS records will
	// be sub-domains of this base.
	//
	// For example, given the base domain `openshift.example.com`, an API server
	// DNS record may be created for `cluster-api.openshift.example.com`.
	//
	// Once set, this field cannot be changed.
	BaseDomain string `json:"baseDomain"`
}
