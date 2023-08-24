package config

// CSI provides configuration options for the inbuilt CSI infrastructure.
type CSI struct {
	// Disable completely disables all internal CSI implementations and allows for Plug & Play of own CSI
	// Plugins instead.
	Disable bool `json:"disable"`
}
