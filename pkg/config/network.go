package config

import (
	"fmt"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
	"k8s.io/apimachinery/pkg/util/sets"
)

// CNIPlugin is an enum value that determines whether MicroShift deploys OVNK.
// +kubebuilder:validation:Enum:="";none;ovnk
type CNIPlugin string

// MultusStatusEnum is an enum value that determines whether MicroShift deploys Multus CNI.
// +kubebuilder:validation:Enum:=Enabled;Disabled
type MultusStatusEnum string

const (
	// CniPluginUnset exists to support backwards compatibility with existing MicroShift clusters. When .network.cniPlugin is
	// "", MicroShift will default to deploying OVNK. This preserves the current deployment behavior of existing
	// clusters.
	CniPluginUnset CNIPlugin = ""
	//  CniPluginNone signals MicroShift to not deploy the LVMS components. Setting the value for a cluster that has already
	//  deployed LVMS will not cause LVMS to be deleted. Otherwise, volumes already deployed on the cluster would be
	//  orphaned once their workloads stop or restart.
	CniPluginNone CNIPlugin = "none"
	// CniPluginOVNK is equivalent to CniPluginUnset, and explicitly tells MicroShift to deploy OVNK. This option exists to
	// provide a differentiation between OVNK and potential future CNI options.
	CniPluginOVNK CNIPlugin = "ovnk"

	// MultusEnabled signals MicroShift to deploy Multus CNI.
	MultusEnabled MultusStatusEnum = "Enabled"

	// MultusEnabled signals MicroShift to not deploy Multus CNI.
	MultusDisabled MultusStatusEnum = "Disabled"
)

type Multus struct {
	// Status controls the deployment of the Multus CNI.
	// Changing from "Enabled" to "Disabled" will not cause Multus CNI to be deleted.
	// Allowed values are: unset (disabled), "Enabled", or "Disabled"
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=Disabled
	Status MultusStatusEnum `json:"status"`
}

type Network struct {
	// CNIPlugin is a user defined string value matching one of the above CNI values. MicroShift uses this
	// value to decide whether to deploy the OVN-K as default CNI. An unset field defaults to "" during yaml parsing, and thus
	// could mean that the cluster has been upgraded. In order to support the existing out-of-box behavior, MicroShift
	// assumes an empty string to mean the OVN-K should be deployed.
	// Allowed values are: unset or one of ["", "ovnk", "none"]
	// +kubebuilder:validation:Optional
	CNIPlugin CNIPlugin `json:"cniPlugin,omitempty"`

	// IP address pool to use for pod IPs.
	// This field is immutable after installation.
	// +kubebuilder:default={"10.42.0.0/16"}
	ClusterNetwork []string `json:"clusterNetwork"`

	// IP address pool for services.
	// Currently, we only support a single entry here.
	// This field is immutable after installation.
	// +kubebuilder:default={"10.43.0.0/16"}
	ServiceNetwork []string `json:"serviceNetwork"`

	// The port range allowed for Services of type NodePort.
	// If not specified, the default of 30000-32767 will be used.
	// Such Services without a NodePort specified will have one
	// automatically allocated from this range.
	// This parameter can be updated after the cluster is
	// installed.
	// +kubebuilder:validation:Pattern=`^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])-([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`
	// +kubebuilder:default="30000-32767"
	ServiceNodePortRange string `json:"serviceNodePortRange"`

	Multus Multus `json:"multus"`

	// The DNS server to use
	DNS string `json:"-"`
}

func (c *Config) computeClusterDNS() (string, error) {
	if len(c.Network.ServiceNetwork) == 0 {
		return "", fmt.Errorf("network.serviceNetwork not filled in")
	}

	clusterDNS, err := getClusterDNS(c.Network.ServiceNetwork[0])
	if err != nil {
		return "", fmt.Errorf("failed to get DNS IP: %v", err)
	}
	return clusterDNS, nil
}

// getClusterDNS returns cluster DNS IP that is 10th IP of the ServiceNetwork
func getClusterDNS(serviceCIDR string) (string, error) {
	_, service, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		return "", fmt.Errorf("invalid service cidr %v: %v", serviceCIDR, err)
	}
	dnsClusterIP, err := cidr.Host(service, 10)
	if err != nil {
		return "", fmt.Errorf("service cidr must have at least 10 distinct host addresses %v: %v", serviceCIDR, err)
	}

	return dnsClusterIP.String(), nil
}

func isValidIPAddress(ipAddress string) bool {
	ip := net.ParseIP(ipAddress)
	return ip != nil
}

func (n Network) validCNIPlugin() (isSupported bool) {
	return sets.New[CNIPlugin](CniPluginUnset, CniPluginOVNK, CniPluginNone).Has(n.CNIPlugin)
}

// IsEnabled returns false only when .network.cniPlugin: "none". An empty value is considered "enabled"
// for backwards compatibility. Otherwise, the meaning of the config would silently change after an
// upgrade from enabled-by-default to disabled-by-default.
func (n Network) IsEnabled() bool {
	return n.CNIPlugin != CniPluginNone
}

func (m Multus) IsEnabled() bool {
	return m.Status == MultusEnabled
}
