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

	// Use controlplane node IP as APIServer backend (instead of 10.44.0.0)
	c.ApiServer.AdvertiseAddress = c.Node.NodeIP
	// Don't configure 10.44.0.0 on lo device
	c.ApiServer.SkipInterface = true

	return c
}
