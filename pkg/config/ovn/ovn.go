package ovn

import (
	"errors"
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

const (
	ConfigFileName = "ovn.yaml"
)

type OVNKubernetesConfig struct {
	// disable microshift-ovs-init.service.
	// OVS bridge "br-ex" needs to be configured manually when disableOVSInit is true.
	DisableOVSInit bool `json:"disableOVSInit,omitempty"`
	// MTU to use for the geneve tunnel interface.
	// This must be 100 bytes smaller than the uplink mtu.
	// Default is 1400.
	MTU uint32 `json:"mtu,omitempty"`
}

func (o *OVNKubernetesConfig) ValidateOVSBridge(bridge string) error {
	_, err := net.InterfaceByName(bridge)
	if err != nil {
		return err
	}
	return nil
}

func (o *OVNKubernetesConfig) withDefaults() *OVNKubernetesConfig {
	o.DisableOVSInit = false
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

func NewOVNKubernetesConfigFromFileOrDefault(path string) (*OVNKubernetesConfig, error) {
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
