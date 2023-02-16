package ovn

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

const (
	ovnConfigFileName           = "ovn.yaml"
	GeneveHeaderLengthIPv4      = 58
	OVNGatewayInterface         = "br-ex"
	OVNExternalGatewayInterface = "br-ex1"
)

type OVNKubernetesConfig struct {
	// Configuration for microshift-ovs-init.service
	OVSInit OVSInit `json:"ovsInit,omitempty"`
	// MTU to use for the geneve tunnel interface.
	// This must be 100 bytes smaller than the uplink mtu.
	// Default is 1400.
	MTU int `json:"mtu,omitempty"`
}

type OVSInit struct {
	// disable microshift-ovs-init.service.
	// OVS bridge "br-ex" needs to be configured manually when disableOVSInit is true.
	DisableOVSInit bool `json:"disableOVSInit,omitempty"`
	// Uplink interface for OVS bridge "br-ex"
	GatewayInterface string `json:"gatewayInterface,omitempty"`
	// Uplink interface for OVS bridge "br-ex1"
	ExternalGatewayInterface string `json:"externalGatewayInterface,omitempty"`
}

func (o *OVNKubernetesConfig) Validate() error {
	// br-ex is required to run ovn-kubernetes
	err := o.validateOVSBridge()
	if err != nil {
		return err
	}
	err = o.validateConfig()
	if err != nil {
		return err
	}
	return nil
}

// validateOVSBridge validates the existence of ovn-kubernetes br-ex bridge
func (o *OVNKubernetesConfig) validateOVSBridge() error {
	_, err := net.InterfaceByName(OVNGatewayInterface)
	return err
}

// validateConfig validates the user defined configuration in /etc/microshift/ovn.yaml
func (o *OVNKubernetesConfig) validateConfig() error {
	// validate gateway interfaces conf
	if o.OVSInit.GatewayInterface != "" {
		_, err := net.InterfaceByName(o.OVSInit.GatewayInterface)
		if err != nil {
			return fmt.Errorf("gateway interface %s not found", o.OVSInit.GatewayInterface)
		}
	}
	if o.OVSInit.ExternalGatewayInterface != "" {
		_, err := net.InterfaceByName(o.OVSInit.ExternalGatewayInterface)
		if err != nil {
			return fmt.Errorf("external gateway interface %s not found", o.OVSInit.ExternalGatewayInterface)
		}
		_, err = net.InterfaceByName(OVNExternalGatewayInterface)
		if err != nil {
			return fmt.Errorf("external gateway interface %s is configured, but external gateway bridge %s not found",
				o.OVSInit.ExternalGatewayInterface, OVNExternalGatewayInterface)
		}
	}

	// validate MTU conf
	iface, err := net.InterfaceByName(OVNGatewayInterface)
	if err != nil {
		return err
	}
	requiredMTU := o.MTU + GeneveHeaderLengthIPv4

	if iface.MTU < requiredMTU {
		return fmt.Errorf("interface MTU (%d) is too small for specified overlay MTU (%d)", iface.MTU, requiredMTU)
	}
	return nil
}

func (o *OVNKubernetesConfig) withDefaults() *OVNKubernetesConfig {
	o.OVSInit.DisableOVSInit = false
	o.MTU = 1400
	return o
}

func newOVNKubernetesConfigFromFile(path string) (*OVNKubernetesConfig, error) {
	o := new(OVNKubernetesConfig)
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buf, &o)
	if err != nil {
		return nil, fmt.Errorf("parsing OVNKubernetes config: %v", err)
	}
	return o, nil
}

func NewOVNKubernetesConfigFromFileOrDefault(dir string) (*OVNKubernetesConfig, error) {
	path := filepath.Join(dir, ovnConfigFileName)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			klog.Info("OVNKubernetes config file not found, assuming default values")
			return new(OVNKubernetesConfig).withDefaults(), nil
		}
		return nil, fmt.Errorf("failed to get OVNKubernetes config file: %v", err)
	}

	o, err := newOVNKubernetesConfigFromFile(path)
	if err == nil {
		klog.Info("got OVNKubernetes config from file %q", path)
		return o, nil
	}
	return nil, fmt.Errorf("getting OVNKubernetes config: %v", err)
}

func GetOVNGatewayIP() (string, error) {
	iface, err := net.InterfaceByName(OVNGatewayInterface)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ip := addr.(*net.IPNet).IP
		// return the first available addr, ipv4 takes precedence in ip.String()
		return ip.String(), nil
	}
	return "", fmt.Errorf("failed to get ovn gateway IP address")
}
