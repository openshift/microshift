/*
Copyright © 2023 MicroShift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package node

import (
	"context"
	"fmt"
	"net"

	"k8s.io/klog/v2"

	"github.com/vishvananda/netlink"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/config/ovn"
)

const (
	// Network configuration component name
	componentNetworkConfiguration = "network-configuration"
	// Interface name where to add service IP
	loopbackInterface = "lo"
)

type NetworkConfiguration struct {
	kasAdvertiseAddresses      []string
	skipInterfaceConfiguration bool
}

func NewNetworkConfiguration(cfg *config.Config) *NetworkConfiguration {
	n := &NetworkConfiguration{}
	n.configure(cfg)
	return n
}

func (n *NetworkConfiguration) Name() string           { return componentNetworkConfiguration }
func (n *NetworkConfiguration) Dependencies() []string { return []string{} }

func (n *NetworkConfiguration) configure(cfg *config.Config) {
	n.kasAdvertiseAddresses = cfg.ApiServer.AdvertiseAddresses
	n.skipInterfaceConfiguration = cfg.ApiServer.SkipInterface
}

func (n *NetworkConfiguration) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	stopChan := make(chan struct{})

	if !n.skipInterfaceConfiguration {
		if err := n.addServiceIPLoopback(); err != nil {
			return err
		}
		go func() {
			<-ctx.Done()
			if err := n.removeServiceIPLoopback(); err != nil {
				klog.Warningf("failed to remove IP from interface: %v", err)
			}
			close(stopChan)
		}()
	}
	klog.Infof("%q is ready", n.Name())
	close(ready)
	<-stopChan
	return ctx.Err()
}

func (n *NetworkConfiguration) addServiceIPLoopback() error {
	var link netlink.Link
	var err error

	link, err = netlink.LinkByName(ovn.OVNGatewayInterface)
	if _, ok := err.(netlink.LinkNotFoundError); ok {
		link, err = netlink.LinkByName(loopbackInterface)
		if err != nil {
			return err
		}
	}

	for _, entry := range n.kasAdvertiseAddresses {
		prefix := 32
		if net.ParseIP(entry).To4() == nil {
			prefix = 128
		}
		address, err := netlink.ParseAddr(fmt.Sprintf("%s/%d", entry, prefix))
		if err != nil {
			return err
		}
		existing, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return err
		}
		for _, existingAddress := range existing {
			if address.Equal(existingAddress) {
				return nil
			}
		}
		if err := netlink.AddrAdd(link, address); err != nil {
			return err
		}
	}

	return nil
}

func (n *NetworkConfiguration) removeServiceIPLoopback() error {
	var link netlink.Link
	var err error

	link, err = netlink.LinkByName(ovn.OVNGatewayInterface)
	if _, ok := err.(netlink.LinkNotFoundError); ok {
		link, err = netlink.LinkByName(loopbackInterface)
		if err != nil {
			return err
		}
	}
	for _, entry := range n.kasAdvertiseAddresses {
		prefix := 32
		if net.ParseIP(entry).To4() == nil {
			prefix = 128
		}
		address, err := netlink.ParseAddr(fmt.Sprintf("%s/%d", entry, prefix))
		if err != nil {
			return err
		}
		existing, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return err
		}
		for _, existingAddress := range existing {
			if address.Equal(existingAddress) {
				if err := netlink.AddrDel(link, address); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
