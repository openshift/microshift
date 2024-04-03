# Multus in MicroShift

MicroShift includes an optional Multus CNI.
It can be used to attach additional network interfaces to Pods.

## How to install

Manifests to deploy Multus CNI are part of `microshift-multus` RPM.
`microshift-multus` RPM resides in the same repository as `microshift` RPM.
Installation depends on source of the RPMs (local build, OpenShift mirror, or RHOCP),
just remember that version of `microshift` and `microshift-multus` need to match.

## Supported plugins on MicroShift

> See section "More information" for external resources.

CNI:
- Bridge
- macvlan
- ipvlan

> Fun fact: ipvlan and macvlan are not related to the VLAN.

IPAM (IP Address Management):
- host-local
- static
- dhcp

## How to use

> Following is a simple example based on one of the CI tests.

It is assumed that the MicroShift cluster is already up and running (including Multus).

```
$ oc get pods -n openshift-multus 
NAME                READY   STATUS    RESTARTS   AGE
dhcp-daemon-dfbzw   1/1     Running   0          5h
multus-rz8xc        1/1     Running   0          5h
```

First step is creation of NetworkAttachmentDefinition.
It is basically a wrapper of raw CNI config in a form of Kubernetes' custom resource.

```sh
$ oc apply -f ./test/assets/multus/bridge-nad.yaml
```

Here's content of the file:
```yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: bridge-conf
spec:
  config: '{
      "cniVersion": "0.4.0",
      "type": "bridge",
      "bridge": "br-test",
      "mode": "bridge",
      "ipam": {
        "type": "host-local",
        "ranges": [
          [
            {
              "subnet": "10.10.0.0/24",
              "rangeStart": "10.10.0.20",
              "rangeEnd": "10.10.0.50",
              "gateway": "10.10.0.254"
            }
          ]
        ],
        "dataDir": "/var/lib/cni/br-test"
      }
    }'
```

Summary of what happens:
- `type` specifies a name of the CNI plugin
- `bridge` is name of the bridge on the MicroShift host that will be used - Pod's additional interface will be connected to that bridge.
  > This option is specific to `bridge` type of plugin. Other plugins have different fields, for example `macvlan` and `ipvlan` use `master` to specify host's interface to "enslave".
- `ipam` specifies which IP Address Management plugin should be used. Three plugins are supported on MicroShift: `host-local`, `static`, and `dhcp`.
  - In the example we're using `host-local` which is sort of like dhcp but without a server and local to the node and assigned IPs are stored in the file on the disk.

Next, we need to create a Pod that will use that NetworkAttachmentDefinition.

```sh
$ oc apply -f ./test/assets/multus/bridge-pod.yaml
```

Most important part are the annotations which specify which NADs should be used.

```yaml
metadata:
  annotations:
    k8s.v1.cni.cncf.io/networks: bridge-conf
```

Pretty soon the Pod should be running:
```sh
$ oc get pod
NAME          READY   STATUS    RESTARTS   AGE
test-bridge   1/1     Running   0          81s
```

And, if it wasn't present before, the bridge will be created on the host:
```sh
$ ip a show br-test
22: br-test: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether 96:bf:ca:be:1d:15 brd ff:ff:ff:ff:ff:ff
    inet6 fe80::34e2:bbff:fed2:31f2/64 scope link 
       valid_lft forever preferred_lft forever
```

Let's configure an IP for the bridge:
```sh
$ sudo ip addr add 10.10.0.10/24 dev br-test
$ ip a show br-test
22: br-test: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether 96:bf:ca:be:1d:15 brd ff:ff:ff:ff:ff:ff
    inet 10.10.0.10/24 scope global br-test
       valid_lft forever preferred_lft forever
    inet6 fe80::34e2:bbff:fed2:31f2/64 scope link 
       valid_lft forever preferred_lft forever
```

Now, let's obtain the IP of the Pod:
```sh
$ oc get pod test-bridge --output=jsonpath='{.metadata.annotations.k8s\.v1\.cni\.cncf\.io/network-status}'
[{
    "name": "ovn-kubernetes",
    "interface": "eth0",
    "ips": [
        "10.42.0.17"
    ],
    "mac": "0a:58:0a:2a:00:11",
    "default": true,
    "dns": {}
},{
    "name": "default/bridge-conf",
    "interface": "net1",
    "ips": [
        "10.10.0.20"
    ],
    "mac": "82:01:98:e5:0c:b7",
    "dns": {}
```

We can also exec into the Pod and double check its interfaces using the `ip` command:
```sh
$ oc exec -ti test-bridge -- ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host 
       valid_lft forever preferred_lft forever
2: eth0@if21: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue 
    link/ether 0a:58:0a:2a:00:11 brd ff:ff:ff:ff:ff:ff
    inet 10.42.0.17/24 brd 10.42.0.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::858:aff:fe2a:11/64 scope link 
       valid_lft forever preferred_lft forever
3: net1@if23: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue 
    link/ether 82:01:98:e5:0c:b7 brd ff:ff:ff:ff:ff:ff
    inet 10.10.0.20/24 brd 10.10.0.255 scope global net1
       valid_lft forever preferred_lft forever
    inet6 fe80::8001:98ff:fee5:cb7/64 scope link 
       valid_lft forever preferred_lft forever
```

Looks like Pod got IP 10.10.0.20 on net1 interface. Now we can check if it actually works.
Let's access the HTTP server that is in the Pod from the MicroShift host:

```sh
$ curl 10.10.0.20:8080
Hello MicroShift
```


## Caveats

- `dhcp-daemon` running inside `openshift-multus` namespace **is not** a DHCP server. It's a daemon that DHCP IPAM Client talks to over filesystem socket, and in turn that daemon is making DHCP requests to the DHCP server.
  - It means that when using `bridge` + `dhcp` a DHCP server listening on the bridge interface is required.
    firewalld configuration may also be required on that interface/zone to allow DHCP traffic.
- Above is not required when using `dhcp` with `macvlan` plugin.
  `macvlan` type interface is essentially accessing the network that the host is connected to,
  so it will receive IP from network's DHCP server (example: in testing, `macvlan` interface will get IP from the range of `192.168.122.0/24` meaning it will be served by DHCP server from libvirt's network).
- Currently `ipvlan` cannot be used with `dhcp` plugin - `ipvlan` interface also has direct access to host's network but it shares the MAC address with the host interface, so DHCP could be used with `ClientID` but it's not supported by the DHCP daemon.


## More information

- [Multus tests in MicroShift's test harness](../../../test/suites/optional/multus.robot)
- ["Multus CNI for MicroShift" enhancement](https://github.com/openshift/enhancements/blob/master/enhancements/microshift/multus-cni-for-microshift.md)
- [OpenShift documentation on using Multus](https://docs.openshift.com/container-platform/latest/networking/multiple_networks/understanding-multiple-networks.html)
- [Upstream CNI documentation](https://www.cni.dev/plugins/current/)
