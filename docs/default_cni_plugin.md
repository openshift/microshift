# MicroShift CNI Plugin Overview

> **IMPORTANT!** The default CNI configuration is intended to match the developer environment described in [MicroShift Development Environment on RHEL 8](./devenv_rhel8.md).

MicroShift uses Red Hat OpenShift Networking CNI driver, based on [ovn-kubernetes](https://github.com/ovn-org/ovn-kubernetes.git).

## Design

### Systemd Services

#### OpenvSwitch

OpenvSwitch is a core component to ovn-kubernetes CNI plugin, it runs as a systemd service on the MicroShift node.
OpenvSwitch rpm package is installed as a dependency to microshift-networking rpm package.

By default, three performance optimizations are applied to openvswitch services to minimize the resource consumption:

1. CPU affinity to ovs-vswitchd.service and ovsdb-server.service
2. no-mlockall to openvswitch.service
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

Ovn-kubernetes cluster manifests can be found in [microshift/assets/components/ovn](../assets/components/ovn).

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

|Field          |Required |Type    |Default |Description                                                       |Example|
|:--------------|:--------|:-------|:-------|:-----------------------------------------------------------------|:------|
|disableOVSInit |N        |bool    |false   |Skip configuring OVS bridge "br-ex" in microshift-ovs-init.service|true   |
|mtu            |N        |uint32  |1400    |MTU value to be used for the Pods                                 |1300   |

> When `disableOVSInit` is true, OVS bridge "br-ex" needs to be configured manually. This OVS bridge is required by ovn-kubernetes CNI. See section [OVS bridge](#ovs-bridge) for guidance on configuring the OVS gateway bridge manually.

Below is an example of `ovn.yaml`:

```yaml
disableOVSInit: true
mtu: 1300
```

### Configuring Host

#### OVS bridge

[comment]: # (TODO: replace OVS commands with nmcli which can be easily installed under /etc)

When `disableOVSInit` is set to true in ovn-kubernetes CNI config file, OVS bridge "br-ex" needs to be manually configured:

```bash
sudo systemctl enable openvswitch --now
sudo ovs-vsctl add-br br-ex
sudo ovs-vsctl add-port br-ex <physical-interface-name>
sudo ip link set br-ex up
```

Replace `<physical-interface-name>` with the network interface name where node IP address is assigned to.
Once br-ex up is up, assign the node IP address to br-ex bridge.

> Adding physical interface to br-ex bridge will disconnect ssh connection on node IP address.

#### Firewalld

Some ovn-kubernetes traffic needs to be explicitly allowed when `firewalld` service is running:

1. CNI pod to CNI pod
2. CNI pod to Host-Network pod
3. Host-Network pod to Host-Network pod

> CNI pod: kubernetes pod that uses CNI network
> Host-Network pod: kubernetes pod that uses host network

Insert and reload the following firewall rules to allow these ovn-kubernetes traffic:

```bash
sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
sudo firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1
sudo firewall-cmd --reload
```

> Known issue: ovn-kubernetes makes use of iptable rules for some traffic flows (such as nodePort service), these iptable rules are generated and insertd by ovn-kubernetes (ovnkube-master container), but can be removed by reloading firewall rules, which in turn break the traffic flows. To avoid such situation, make sure to execute firewall commands before starting ovn-kubernetes pods. If firewall commands have to be executed after ovn-kubernetes pods have started, manually restart the ovnkube-master pod to trigger reinsertion of ovn-kubernetes iptable rules.

#### IP Forward

Host network `sysctl net.ipv4.ip_forward` is automatically enabled by ovn-kubernetes (ovnkube-master container) when it starts. This is needed to forward incoming traffic to ovn-kubernetes CNI network. For example, accessing node port service from outside of cluster fails if `ip_forward` is disabled.

#### Host Ports

Ingress routes are exposed on host port `80` and `443` automatically by `route-default` deployment in `openshift-ingress` namespace.
Data profiling with pprof is exposed on host port `29500` when starting MicroShift with `--debug.pprof`, disabled by default.

## Network Topology

ovn-kubernetes provides an overlay based networking implementation for Kubernetes, including OVS based implementation of Service/ServiceIP and NetworkPolicy.
The overlay network is using geneve tunnel, so the pod mtu is set to smaller than that of physical interface on the host to remove the overhead of tunnel header.

There are at least two OVS bridges required by ovn-kubernetes:
1. gateway bridge `br-ex`
2. integration bridge `br-int`

The first bridge is created by `microshift-ovs-init.service` or manually.
The second bridge is created by ovn-kubernetes (ovnkube-master container).

`br-ex` contains statically programmed openflow rules which distinguish traffic to/from host network (underlay) and ovn network (overlay).
`br-int` contains dynamically programmed openflow rules which handles cluster network traffic.
`br-ex` and `br-int` bridges are connected via OVS patch ports. Traffic from external to cluster network traves from `br-ex` to `br-int` via patch port, vice versa.
Kubernetes pods are connected to `br-int` bridge via veth pair, one end of the veth pair is attached to pod namespace, the other end is attached to `br-int` bridge.

A snapshot of OVS interfaces from running MicroShift cluster:

```bash
# ovs-vsctl show
9d9f5ea2-9d9d-4e34-bbd2-dbac154fdc93
    Bridge br-ex
        Port enp1s0
            Interface enp1s0
                type: system
        Port br-ex
            Interface br-ex
                type: internal
        Port patch-br-ex_localhost.localdomain-to-br-int
            Interface patch-br-ex_localhost.localdomain-to-br-int
                type: patch
                options: {peer=patch-br-int-to-br-ex_localhost.localdomain}
    Bridge br-int
        fail_mode: secure
        datapath_type: system
        Port patch-br-int-to-br-ex_localhost.localdomain
            Interface patch-br-int-to-br-ex_localhost.localdomain
                type: patch
                options: {peer=patch-br-ex_localhost.localdomain-to-br-int}
        Port eebee1ce5568761
            Interface eebee1ce5568761
        Port b47b1995ada84f4
            Interface b47b1995ada84f4
        Port "3031f43d67c167f"
            Interface "3031f43d67c167f"
        Port br-int
            Interface br-int
                type: internal
        Port ovn-k8s-mp0
            Interface ovn-k8s-mp0
                type: internal
    ovs_version: "2.17.3"
```

> `patch-br-ex_localhost.localdomain-to-br-int` and `patch-br-int-to-br-ex_localhost.localdomain` are OVS patch ports that connect `br-ex` and `br-int`.
> `eebee1ce5568761`, `b47b1995ada84f4` and `3031f43d67c167f` are pod interfaces plugged in `br-int` bridge, they are named with the first 15 bits of pod sandbox ID.
> `ovn-k8s-mp0` is an OVS internal port for hairpin traffic, created by ovn-kubernetes (ovnkube-master container).

## Network Features

A wide range of networking features are available with ovn-kubernetes, including but not limited to:

* Kubernetes network policy
* Dynamic node IP
* Cluster network on specified host interface
* Secondary gateway interface
* Dual stack
* Service idling (disabled)
* Egress firewall/QOS/IP (disabled)

### Cluster network on specified host interface

microshift-ovs-init.service is able to use user specified host interface for cluster network.
This is done by specifying the desired host interface name in a hint file `/var/lib/ovnk/iface_default_hint`.
The specified interface will be added in OVS bridge `br-ex` which acts as gateway bridge for ovn-kubernetes CNI network.


### Secondary gateway interface

microshift-ovs-init.service is able to setup one additional host interface for cluster ingress/egress traffic.
This is done by adding the additional host interface name in another hint file `/etc/ovnk/extra_bridge`.
The additional interface will be added in a second OVS bridge `br-ex1`. Cluster pod traffic destinated to additional host subnet will be routed through `br-ex1`.

## Troubleshooting

### NodePort service

Ovn-kubernetes sets up iptable chain in NAT table to handle incoming traffic destinated to node port service.
In the case that node port service is not reachable or connection refused, check the iptable rules on the host to make sure the relevant rules are properly inserted.

An example of iptable rules for node port service is as below:

```bash
$ iptables-save
[...]
-A OUTPUT -j OVN-KUBE-NODEPORT
-A OVN-KUBE-NODEPORT -p tcp -m addrtype --dst-type LOCAL -m tcp --dport 30326 -j DNAT --to-destination 10.43.95.170:80
[...]
```

ovn-kubernetes configures OVN-KUBE-NODEPORT chain in iptable NAT table to match on destination port and DNATs the packet to the backend clusterIP service.
The DNATed packet is then routed to OVN network through gateway bridge `br-ex` via routing rules on the host:

```bash
$ ip route
[...]
10.43.0.0/16 via 192.168.122.1 dev br-ex mtu 1400
[...]
```

The above routing rule matches the kubernetes service IP range and forward packet to gateway bridge `br-ex`. It is therefor required to enable `ip_forward` on the host.
Once the packet is forwarded to OVS bridge `br-ex`, it is handled by openflow rules in OVS which steers the packet to OVN network and eventually to the Pod.
