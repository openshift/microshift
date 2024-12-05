# Host Networking in MicroShift

## Host Ports

The following host ports are allocated for microshift components by default:

|port    |namespace          |pod                |comment                                |
|:-------|:------------------|:------------------|:--------------------------------------|
|80      |openshift-ingress  |router-default-xxx |http                                   |
|443     |openshift-ingress  |router-default-xxx |https                                  |

Host ports exposure are implemented via iptable rules, see section [iptable -> cri-o](#cri-o) for details on how iptable rules are configured.<br>

**NOTE:** These host ports shall be reserved for MicroShift.

## IPs

Besides the cluster network subnet (default: 10.42.0.0/16) and cluster service subnet (default: 10.43.0.0/16), ovn-kubernetes uses the following IPs in its internal implementation:

|ip                     |proto      | comment                                                          |attached interface   |
|:----------------------|:----------|:-----------------------------------------------------------------|:--------------------|
|169.254.169.1          |v4         |used for ipv4 hairpin traffic from host to hostservice            |no                   |
|169.254.169.2          |v4         |used for ipv4 traffic from host to service                        |br-ex                |
|169.254.169.3          |v4         |used for ipv4 nodeport service with externalTrafficPolicy=Local   |no                   |
|169.254.169.4          |v4         |used as default gateway IP when ipv4 default route doesn't exist  |ovn logical interface|
|100.64.0.0/16          |v4         |used in ovn logical switch                                        |ovn logical interface|
|fd69::1                |v6         |used for ipv6 hairpin traffic from host to hostservice            |no                   |
|fd69::2                |v6         |used for ipv6 traffic from host to service                        |br-ex                |
|fd69::3                |v6         |used for ipv6 nodeport service with externalTrafficPolicy=Local   |no                   |
|fd69::4                |v6         |used as default gateway IP when ipv6 default route doesn't exist  |ovn logical interface|
|fd98::/64              |v6         |used in ovn logical switch                                        |ovn logical interface|

Not all IP addresses are attached to specific physical interfaces, some of them are only used for intermediate packet processing, others are attached to ovn logical interfaces visible in ovn logical network. See section [Interfaces](#interfaces) for details on different interface types.<br>

**NOTE:** These IPs shall be reserved for MicroShift.

**NOTE:** There is another special IP reserved for MicroShift's apiserver. In order to allow external access using host IPs a new local IP must be allocated for the apiserver. It defaults to the first IP in the next available subnet from the service CIDR. If service CIDR is 10.43.0.0/16, then the new IP will be 10.44.0.0/32. This default IP is added to the loopback interface to allow connectivity without ovnk.

## Interfaces

The following physical network interfaces are created or modified by ovn-kubernetes:

|name               |type              |description            |comment                                                                                             |
|:------------------|:-----------------|:----------------------|:---------------------------------------------------------------------------------------------------|
|br-ex              |OVS bridge        |gateway bridge         |created by microshift-ovs-init.service or manually                                                  |
|br-int             |OVS bridge        |integration bridge     |created by ovnkube-master container                                                                 |
|patch-br-ex        |OVS patch port    |                       |created by ovnkube-master container, connect br-ex to br-int                                        |
|patch-br-int       |OVS patch port    |                       |created by ovnkube-master container, connect br-int to br-ex                                        |
|ovn-k8s-mp0        |OVS internal port |management port        |created by ovnkube-master container, assigned with the second IP address from cluster network subnet|
|enp1s0             |physical          |uplink interface       |added into br-ex by microshift-ovs-init.service, IP moved from enp1s0 to br-ex                      |
|7ea12e348b34f1e    |veth              |pod veth interface     |created and plugged to br-int by ovnkube-master container, the other end connects to pod namespace  |

- `7ea12e348b34f1e` is one end of veth pair that connects pod to br-int, it is named after the first 15 bits of pod sandbox ID. The other end of veth pair is in pod network namespace (named `eth0` inside pod). There could be as many veth pairs as the number of pods. <br>

A snapshot of OVS interfaces from running MicroShift cluster:

```text
(host)$ ovs-vsctl show
9d9f5ea2-9d9d-4e34-bbd2-dbac154fdc93
    Bridge br-int
        fail_mode: secure
        datapath_type: system
        Port patch-br-int-to-br-ex_openshift.default.svc.cluster.local
            Interface patch-br-int-to-br-ex_openshift.default.svc.cluster.local
                type: patch
                options: {peer=patch-br-ex_openshift.default.svc.cluster.local-to-br-int}
        Port ovn-k8s-mp0
            Interface ovn-k8s-mp0
                type: internal
        Port br-int
            Interface br-int
                type: internal
        Port "7ea12e348b34f1e"
            Interface "7ea12e348b34f1e"
        Port "11372c66e9363c5"
            Interface "11372c66e9363c5"
        Port "714eb2a1b67ccd5"
            Interface "714eb2a1b67ccd5"
    Bridge br-ex
        Port patch-br-ex_openshift.default.svc.cluster.local-to-br-int
            Interface patch-br-ex_openshift.default.svc.cluster.local-to-br-int
                type: patch
                options: {peer=patch-br-int-to-br-ex_openshift.default.svc.cluster.local}
        Port enp1s0
            Interface enp1s0
                type: system
        Port br-ex
            Interface br-ex
                type: internal
    ovs_version: "2.17.3"
```

`br-ex` is assigned with a second IP address `169.254.169.2` to facilitate kubernetes service traffic.

```text
(host)$ ip a s br-ex
7: br-ex: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UNKNOWN group default qlen 1000
    link/ether 52:54:00:e2:ed:d0 brd ff:ff:ff:ff:ff:ff
    inet 192.168.122.14/24 brd 192.168.122.255 scope global dynamic noprefixroute br-ex
       valid_lft 2991sec preferred_lft 2991sec
    inet 169.254.169.2/29 brd 169.254.169.7 scope global br-ex
       valid_lft forever preferred_lft forever
    inet6 fe80::5686:d56:4f90:af7/64 scope link noprefixroute
       valid_lft forever preferred_lft forever
```

Special routes are associated to `br-ex` interface to facilitate traffic destinated to kubernetes service (10.43.0.0/16) from host:

```text
(host)$ ip route show 10.43.0.0/16
10.43.0.0/16 via 169.254.169.4 dev br-ex mtu 1500

(host)$ ip route show 169.254.169.1
169.254.169.1 dev br-ex src 192.168.122.14 mtu 1500

(host)$ ip route show 169.254.169.3
169.254.169.3 via 10.42.0.1 dev ovn-k8s-mp0
```

`169.254.169.1` is used for service hairpin traffic from host to host endpoint. Packet is SNATed to `169.254.169.1` right before reaching host endpoint, the reply packet follows this route.<br>
`169.254.169.3` is used for nodeport service when `externalPolicyType` is set to `Local`, see the [OVN-KUBE-ETP chain example](#ovn-kube-etp-chain-example) for details.<br>
`169.254.169.4` is the fake next hop address used for service traffic routing, it is assigned to the router port `rtoe-GR_<nodename>` in ovn logical gateway router in the case that default route doesn't exist, for example:

```text
(northd)$ ovn-nbctl lr-list
9009dd05-e6e6-42d0-85f0-239d78397f89 (GR_openshift.default.svc.cluster.local)

(northd)$ ovn-nbctl list logical_router GR_openshift.default.svc.cluster.local
ports               : [42380108-59e4-46ca-b5ea-e564bb265d19, e5e51d55-933d-4896-96e1-464a72e59415]
[...]

(northd)$ ovn-nbctl list logical_router_port 42380108-59e4-46ca-b5ea-e564bb265d19
_uuid               : 42380108-59e4-46ca-b5ea-e564bb265d19
enabled             : []
external_ids        : {gateway-physical-ip=yes}
gateway_chassis     : []
ha_chassis_group    : []
ipv6_prefix         : []
ipv6_ra_configs     : {}
mac                 : "52:54:00:e2:ed:d0"
name                : rtoe-GR_openshift.default.svc.cluster.local
networks            : ["169.254.169.4/24"]
options             : {}
peer                : []
```

## IPTable

### ovn-kubernetes

The following iptable chains are added in host network by ovn-kubernetes to facilitate host to pod traffic or service traffic to endpoints:

|chain                  |used for                                 |table      | comment                                  |applicable|
|:----------------------|:----------------------------------------|:----------|:-----------------------------------------|:---------|
|OVN-KUBE-SNAT-MGMTPORT |host to pod traffic                      |nat        |called from nat-POSTROUTING only          |Yes       |
|OVN-KUBE-ITP           |service with internalTrafficPolicy=Local |nat/mangle |called from mangle-OUTPUT and nat-OUTPUT  |Yes       |
|OVN-KUBE-EGRESS-SVC    |egress Service                           |nat        |called from nat-POSTROUTING               |No        |
|OVN-KUBE-NODEPORT      |nodeport Service                         |nat        |called from nat-PREROUTING and nat-OUTPUT |Yes       |
|OVN-KUBE-EXTERNALIP    |external IP Service                      |nat        |called from nat-PREROUTING and nat-OUTPUT |Yes       |
|OVN-KUBE-ETP           |service with externalTrafficPolicy=Local |nat        |called from nat-PREROUTING only           |Yes       |

#### OVN-KUBE-SNAT-MGMTPORT

```text
(host)$ iptables-save -t nat | grep OVN-KUBE-SNAT-MGMTPORT
:OVN-KUBE-SNAT-MGMTPORT - [0:0]
-A POSTROUTING -o ovn-k8s-mp0 -j OVN-KUBE-SNAT-MGMTPORT
-A OVN-KUBE-SNAT-MGMTPORT -o ovn-k8s-mp0 -m comment --comment "OVN SNAT to Management Port" -j SNAT --to-source 10.42.0.2
```

**NOTE:** the above iptable rule SNAT's packet entering ovn network via ovn-k8s-mp0. It applies to traffic from host to pod.

#### OVN-KUBE-ITP (ClusterService)

Example of OVN-KUBE-ITP for service with pod endpoints.

```text
(host)$ oc get svc
NAME            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
svc-itp-local   ClusterIP   10.43.181.31    <none>        9000/TCP         26s

(host)$ oc get endpoints
NAME            ENDPOINTS             AGE
svc-itp-local   10.42.0.18:8080       28s

(host)$ iptables-save -t mangle | grep OVN-KUBE-ITP
:OVN-KUBE-ITP - [0:0]
-A OUTPUT -j OVN-KUBE-ITP
-A OVN-KUBE-ITP -d 10.43.181.31/32 -p tcp -m tcp --dport 9000 -j MARK --set-xmark 0x1745ec/0xffffffff

(host)$ ip rule
30:	from all fwmark 0x1745ec lookup 7

(host)$ ip route show table 7
0.43.0.0/16 via 10.42.0.1 dev ovn-k8s-mp0
```

#### OVN-KUBE-ITP (HostService)

Example of OVN-KUBE-ITP for service with host endpoints.

```text
(host)$ oc get svc
NAME            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
svc-itp-local   ClusterIP   10.43.181.31    <none>        9000/TCP         17m

(host)$ oc get endpoints
NAME            ENDPOINTS                             AGE
svc-itp-local   192.168.122.14:8080                   18m

(host)$ iptables-save -t nat | grep ITP
:OVN-KUBE-ITP - [0:0]
-A OUTPUT -j OVN-KUBE-ITP
-A OVN-KUBE-ITP -d 10.43.181.31/32 -p tcp -m tcp --dport 9000 -j REDIRECT --to-ports 8080
```

#### OVN-KUBE-NODEPORT

```text
(host)$ oc get svc svc-nodeport
NAME            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
svc-nodeport    NodePort    10.43.19.182    <none>        7000:30700/TCP   21h

(host)$ oc get endpoints svc-nodeport
NAME            ENDPOINTS             AGE
svc-nodeport    10.42.0.13:8080       21h

(host)$ iptables-save  | grep OVN-KUBE-NODEPORT
:OVN-KUBE-NODEPORT - [0:0]
-A PREROUTING -j OVN-KUBE-NODEPORT
-A OUTPUT -j OVN-KUBE-NODEPORT
-A OVN-KUBE-NODEPORT -p tcp -m addrtype --dst-type LOCAL -m tcp --dport 30700 -j DNAT --to-destination 10.43.19.182:7000
```

#### OVN-KUBE-EXTERNALIP

```text
(host)$ oc get svc
NAME              TYPE        CLUSTER-IP      EXTERNAL-IP       PORT(S)          AGE
svc-external-ip   ClusterIP   10.43.127.6     192.168.122.150   6000/TCP         3m5s

(host)$ oc get endpoints
NAME              ENDPOINTS                             AGE
svc-external-ip   10.42.0.19:8080                       4m3s

(host)$ iptables-save -t nat | grep EXTERNALIP
:OVN-KUBE-EXTERNALIP - [0:0]
-A PREROUTING -j OVN-KUBE-EXTERNALIP
-A OUTPUT -j OVN-KUBE-EXTERNALIP
-A OVN-KUBE-EXTERNALIP -d 192.168.122.150/32 -p tcp -m tcp --dport 6000 -j DNAT --to-destination 10.43.127.6:6000
```

#### OVN-KUBE-ETP

```text
(host)$ oc get svc svc-etp-local
NAME            TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
svc-etp-local   NodePort   10.43.169.254   <none>        8000:30800/TCP   106s

(host)$ oc get endpoints svc-etp-local
NAME            ENDPOINTS         AGE
svc-etp-local   10.42.0.17:8080   116s

(host)$ iptables-save -t nat | grep OVN-KUBE-SNAT-MGMTPORT
:OVN-KUBE-SNAT-MGMTPORT - [0:0]
-A POSTROUTING -o ovn-k8s-mp0 -j OVN-KUBE-SNAT-MGMTPORT
-A OVN-KUBE-SNAT-MGMTPORT -p tcp -m tcp --dport 30800 -j RETURN
-A OVN-KUBE-SNAT-MGMTPORT -o ovn-k8s-mp0 -m comment --comment "OVN SNAT to Management Port" -j SNAT --to-source 10.42.0.2

(host)$ iptables-save -t nat | grep OVN-KUBE-ETP
:OVN-KUBE-ETP - [0:0]
-A PREROUTING -j OVN-KUBE-ETP
-A OVN-KUBE-ETP -p tcp -m addrtype --dst-type LOCAL -m tcp --dport 30800 -j DNAT --to-destination 169.254.169.3:30800

(host)$  ip route
169.254.169.3 via 10.42.0.1 dev ovn-k8s-mp0
```

**NOTE:** The `-j RETURN` rule in OVN-KUBE-SNAT-MGMTPORT prevents client packet from being SNATed by the next rule in the same chain. This is to preserve client IP address for externalTrafficPolicy=Local service.

### cri-o

The following iptable chain is added by cri-o to expose pod traffic to local host port.

|chain                  |description                                  |table      | comment                                  |applicable|
|:----------------------|:--------------------------------------------|:----------|:-----------------------------------------|:---------|
|KUBE-HOSTPORTS         |pod with containers[x].ports.hostPort defined|nat        |called from nat-PREROUTING and nat-OUTPUT |Yes       |


#### KUBE-HOSTPORTS

```text
(host)$ oc -n openshift-ingress get deployment router-default -o json | jq -r '.spec.template.spec.containers[0].ports'
[
  {
    "containerPort": 80,
    "hostPort": 80,
    "name": "http",
    "protocol": "TCP"
  },
  {
    "containerPort": 443,
    "hostPort": 443,
    "name": "https",
    "protocol": "TCP"
  },
  {
    "containerPort": 1936,
    "name": "metrics",
    "protocol": "TCP"
  }
]

(host)$ oc get pods -n openshift-ingress -o wide
NAME                             READY   STATUS    RESTARTS   AGE    IP          NODE
router-default-ddc545d88-9bfcb   1/1     Running   0          7d6h   10.42.0.5   openshift.default.svc.cluster.local

(host)$ iptables-save -t nat | grep KUBE-HOSTPORTS
:KUBE-HOSTPORTS - [0:0]
-A PREROUTING -m comment --comment "kube hostport portals" -m addrtype --dst-type LOCAL -j KUBE-HOSTPORTS
-A OUTPUT -m comment --comment "kube hostport portals" -m addrtype --dst-type LOCAL -j KUBE-HOSTPORTS
-A KUBE-HOSTPORTS -p tcp -m comment --comment "k8s_router-default-ddc545d88-9bfcb_openshift-ingress_978b6825-cbab-4dc5-af57-afea77d72b67_0_ hostport 443" -m tcp --dport 443 -j KUBE-HP-Z4QL6F3XYFQDKKHE
-A KUBE-HOSTPORTS -p tcp -m comment --comment "k8s_router-default-ddc545d88-9bfcb_openshift-ingress_978b6825-cbab-4dc5-af57-afea77d72b67_0_ hostport 80" -m tcp --dport 80 -j KUBE-HP-RUFPQQH54SC4QBY6
```

## Firewalld

Some ovn-kubernetes traffic needs to be explicitly allowed when `firewalld` service is running:

- Pod to host
- Pod to host service (kubernetes service backed by host endpoints)

Insert and reload the following firewall rules to allow these ovn-kubernetes traffic (note the `clusterNetwork` must be added to the rules, this example shows defaults):

```text
sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
sudo firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1
sudo firewall-cmd --permanent --zone=trusted --add-source=fd01::/48
sudo firewall-cmd --reload
```

## IP Forward

Host network `sysctl net.ipv4.ip_forward` is automatically enabled by ovn-kubernetes ovnkube-master container when it starts. This is needed to forward incoming traffic to ovn network. For example, node port service from outside of cluster fails if `ip_forward` is disabled.
