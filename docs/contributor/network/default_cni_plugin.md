# MicroShift CNI Plugin Overview

> **IMPORTANT!** The default CNI configuration is intended to match the developer environment described in [MicroShift Development Environment](../devenv_setup.md).

MicroShift uses Red Hat OpenShift Networking CNI driver, based on [ovn-kubernetes](https://github.com/ovn-org/ovn-kubernetes.git).

## Design

### Systemd Services

#### OpenvSwitch

OpenvSwitch is a core component to ovn-kubernetes CNI plugin, it runs as a systemd service on the MicroShift node.
OpenvSwitch rpm package is installed as a dependency to microshift-networking rpm package.

By default, three performance optimizations are applied to openvswitch services to minimize the resource consumption:

1. CPU affinity to ovs-vswitchd.service and ovsdb-server.service
2. No-mlockall to openvswitch.service
3. Limit handler and revalidator threads to ovs-vswitchd.service

OpenvSwitch service is enabled and started immediately after installing microshift-networking package.

#### NetworkManager

NetworkManager is required by ovn-kubernetes to setup initial gateway bridge on the MicroShift node.
NetworkManager and NetworkManager-ovs rpm packages are installed as dependencies to microshift-networking rpm.
NetworkManager is configured to use `keyfile` plugin and is restarted immediately after installing microshift-networking package to take in the config change.

#### microshift-ovs-init

microshift-ovs-init.service is installed by microshift-networking rpm as oneshot systemd service.
microshift-ovs-init.service executes configure-ovs.sh script which uses NetworkManager commands to setup OVS gateway bridge.

### OVN Containers

Ovn-kubernetes cluster manifests can be found in [microshift/assets/components/ovn](../../../assets/components/ovn).

Two ovn-kubernetes daemonsets are rendered and applied by MicroShift binary.

1. ovnkube-master: includes northd, nbdb, sbdb and ovnkube-master containers
2. ovnkube-node: includes ovn-controller container

Ovn-kubernetes daemonsets are deployed in the `openshift-ovn-kubernetes` namespace, after MicroShift boots.

## Packaging

Ovn-kubernetes manifests and startup logic are built into MicroShift main binary (microshift rpm).
Systemd services and configurations are included in microshift-networking rpm package:
1. microshift-nm.conf for NetworkManager.service
2. microshift-cpuaffinity.conf for ovs-vswitchd.service
3. microshift-cpuaffinity.conf for ovsdb-server.service
4. microshift-ovs-init.service
5. configure-ovs.sh for microshift-ovs-init.service
6. configure-ovs-microshift.sh for microshift-ovs-init.service

## Configurations

### Configuring ovn-kubernetes

The user provided ovn-kubernetes config should be written to `/etc/microshift/ovn.yaml`.
MicroShift will assume default ovn-kubernetes config values if ovn-kubernetes config file is not provided.

The following configs are supported in ovn-kubernetes config file:

|Field                            |Required |Type    |Default |Description                                                                  |Example|
|:--------------------------------|:--------|:-------|:-------|:----------------------------------------------------------------------------|:------|
|mtu                              |N        |int     |*auto*  |MTU value to be used for the Pods, must be less than or equal to the MTU of default route interface|1500|

> When `mtu` is not provided, it is set to the default route MTU.

Below is an example of `ovn.yaml`:

```yaml
mtu: 1500
```
**NOTE:* The change of `mtu` configuration in `ovn.yaml` requires node reboot to take effect. <br>

## Network Features

A wide range of networking features are available with MicroShift and ovn-kubernetes, including but not limited to:

* Network policy
* Dynamic node IP
* Custom gateway interface
* Second gateway interface
* Blocking external access to NodePort service on specific host interfaces

### Network Policy

Network Policy restricts network traffic to and/or from kubernetes pods.
The ovn-kubernetes implementation of network policy supports pod, namespace and ipBlock based identifiers as well as Ingress and Egress isolation types.
See [ovn-kubernetes network policy](https://github.com/ovn-org/ovn-kubernetes/blob/master/docs/network-policy.md) doc for detailed design and configurations.

### Dynamic node IP

MicroShift is able to detect node IP change and restarts itself to take in the new IP address.
Upon restarting, it recreates ovnkube-master daemonset with updated IP address in openshift-ovn-kubernetes namespace.

### Blocking external access to NodePort service on specific host interfaces

ovn-kubernetes doesn't restrict the host interfaces where NodePort service can be accessed from outside MicroShift node. The following `nft` instructions block NodePort service on a specific host interface. <br>

Insert a new rule in table `ip nat` chain `PREROUTING` to drop the packet with matching destination port and ip:
```text
(host)$ NODEPORT=30700
(host)$ INTERFACE_IP=192.168.150.33
(host)$ nft -a insert rule ip nat PREROUTING tcp dport $NODEPORT ip daddr $INTERFACE_IP drop
```
> Replace value of NODEPORT variable with the host port number assigned to kubernetes NodePort service <br>
> Replace value of INTERFACE_IP with the IP address from the host interface where you'd like to block the NodePort service <br>

List the newly added nftable rule:
```text
(host)$ nft -a list chain ip nat PREROUTING
table ip nat {
	chain PREROUTING { # handle 1
		type nat hook prerouting priority dstnat; policy accept;
		tcp dport 30700 ip daddr 192.168.150.33 drop # handle 134
		counter packets 108 bytes 18074 jump OVN-KUBE-ETP # handle 116
		counter packets 108 bytes 18074 jump OVN-KUBE-EXTERNALIP # handle 114
		counter packets 108 bytes 18074 jump OVN-KUBE-NODEPORT # handle 112
	}
}
```

> Record the `handle` number of the newly added rule (for removal)<br>

Remove the custom nftable rule:
```text
(host)$ nft -a delete rule ip nat PREROUTING handle 134
```

Use [nftables systemd service](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/9/html/configuring_firewalls_and_packet_filters/getting-started-with-nftables_firewall-packet-filters#automatically-loading-nftables-rules-when-the-system-boots_writing-and-executing-nftables-scripts) to persist and automatically load nftable rules when the system boots
