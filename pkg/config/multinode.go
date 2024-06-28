package config

type MultiNodeConfig struct {
	Enabled bool `json:"enabled"`
	// only one controlplane node is supported
	// IP address of control plane node
	Controlplane string `json:"controlplane"`
}

// ConfigMultiNode populates multinode configurations to Config.MultiNode
func ConfigMultiNode(c *Config, enabled bool) *Config {
	if !enabled {
		return c
	}
	c.MultiNode.Enabled = enabled
	c.MultiNode.Controlplane = c.Node.NodeIP

	// Use controlplane node IP as APIServer backend (instead of next available
	// IP from service network)
	c.ApiServer.AdvertiseAddress = c.Node.NodeIP
	// Don't configure the advertise address on the device
	c.ApiServer.SkipInterface = true

	return c
}
