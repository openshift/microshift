package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
)

type Node struct {
	// If non-empty, will use this string to identify the node instead of the hostname
	HostnameOverride string `json:"hostnameOverride"`

	// IP address of the node, passed to the kubelet.
	// If not specified, kubelet will use the node's default IP address.
	NodeIP string `json:"nodeIP"`
}

// Determine if the config file specified a NodeName (by default it's assigned the hostname)
func (c *Config) isDefaultNodeName() bool {
	hostname, err := os.Hostname()
	if err != nil {
		klog.Fatalf("Failed to get hostname %v", err)
	}
	return c.CanonicalNodeName() == strings.ToLower(hostname)
}

// CanonicalNodeName returns the name to use for the node. The value
// is taken from either the HostnameOverride provided by the user in
// the config file, or the host name.
func (c *Config) CanonicalNodeName() string {
	return strings.ToLower(c.Node.HostnameOverride)
}

// Read or set the NodeName that will be used for this MicroShift instance
func (c *Config) establishNodeName(dataDir string) (string, error) {
	name := c.CanonicalNodeName()
	filePath := filepath.Join(dataDir, ".nodename")
	contents, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		// ensure that dataDir exists
		if err := os.MkdirAll(dataDir, 0700); err != nil {
			return "", fmt.Errorf("failed to create data dir: %w", err)
		}
		if err := os.WriteFile(filePath, []byte(name), 0400); err != nil {
			return "", fmt.Errorf("failed to write nodename file %q: %v", filePath, err)
		}
		return name, nil
	} else if err != nil {
		return "", err
	}
	return string(contents), nil
}

// Validate the NodeName to be used for this MicroShift instances
func (c *Config) validateNodeName(isDefaultNodeName bool, dataDir string) error {
	currentNodeName := c.CanonicalNodeName()
	if addr := net.ParseIP(currentNodeName); addr != nil {
		return fmt.Errorf("NodeName can not be an IP address: %q", currentNodeName)
	}

	establishedNodeName, err := c.establishNodeName(dataDir)
	if err != nil {
		return fmt.Errorf("failed to establish NodeName: %v", err)
	}

	if establishedNodeName != currentNodeName {
		if !isDefaultNodeName {
			return fmt.Errorf("configured NodeName %q does not match previous NodeName %q , NodeName cannot be changed for a device once established",
				currentNodeName, establishedNodeName)
		} else {
			c.Node.HostnameOverride = establishedNodeName
			klog.Warningf("NodeName has changed due to a host name change, using previously established NodeName %q."+
				"Please consider using a static NodeName in configuration", establishedNodeName)
		}
	}

	return nil
}

func (c *Config) EnsureNodeNameHasNotChanged() error {
	// Validate NodeName in config file, node-name should not be changed for an already
	// initialized MicroShift instance. This can lead to Pods being re-scheduled, storage
	// being orphaned or lost, and other side effects.
	return c.validateNodeName(c.isDefaultNodeName(), DataDir)
}
