package options

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server"
)

// this file allows us to configure a second kube-apiserver port using the "normal" tls configuration.

func (s *SecureServingOptions) AddReadyOnlyFlags(fs *pflag.FlagSet) {
	fs.IPVar(&s.ReadyOnlyBindAddress, "ready-only-bind-address", s.ReadyOnlyBindAddress, ""+
		"The IP address on which to listen for the --secure-port port. The "+
		"associated interface(s) must be reachable by the rest of the cluster, and by CLI/web "+
		"clients. If blank or an unspecified address (0.0.0.0 or ::), all interfaces will be used.")

	desc := "The port on which to serve HTTPS with authentication and authorization."
	if s.Required {
		desc += " It cannot be switched off with 0."
	} else {
		desc += " If 0, don't serve HTTPS at all."
	}
	fs.IntVar(&s.ReadyOnlyBindPort, "ready-only-secure-port", s.ReadyOnlyBindPort, desc)
}

func (s *SecureServingOptions) ApplyReadyOnlyTo(config **server.SecureServingInfo) error {
	if s.Listener != nil {
		return fmt.Errorf("cannot ApplyReadyOnlyTo with a listener set")
	}
	realBindAddress := s.BindAddress
	realBindPort := s.BindPort
	defer func() {
		s.BindAddress = realBindAddress
		s.BindPort = realBindPort
	}()

	return s.ApplyTo(config)
}