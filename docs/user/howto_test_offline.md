# Running MicroShift Fully Offline

There are several steps needed to run MicroShift in a fully offline
setting, with no network access at all. This document describes a
process for setting up a host to test MicroShift with no network.

## Image Management

Refer to [Embedding MicroShift Container Images for Offline
Deployments](howto_offline_containers.md) for instructions for
preparing the ostree image with the necessary container images
embedded when they cannot be pulled over the network.

## Setup

The commands below are meant to be exeuted as the root user. Please
login as root. Make sure the appropriate kubeconfig file is available
in your environment so the oc client commands can run as expected.

Warning: These instructions involve disabling the network interface on
the host. You will need to access the host directly, not via ssh. For
VMs, `virsh console` is useful.

## MicroShift Configuration

Kubernetes expects to have some basic network settings on the host for
node identification and DNS. When real settings are not going to be
available, alternatives need to be provided for robust disconnected
operation.

1. Start by preparing a VM or host in the normal way and installing
   the version of MicroShift you want to test.

2. Stop MicroShift and clean up the OVN data.

```
/usr/bin/microshift-cleanup-data --ovn
```

3. Determine the interface on the host that has the default route.

```
IFACE=$(ip route | grep default | awk 'NR==1 {print $5}')
```

(This is usually enp1s0 in a VM)

4. Stop external network access by bringing the NIC down.

```
nmcli connection down $IFACE
nmcli connection modify $IFACE connection.autoconnect no
```

5. Ensure the host has a persistent name

```
NAME=hostname.example.com
hostnamectl hostname $NAME
```

6. Add a fake IP on the loopback interface.

```
IP=10.44.0.1
nmcli con add type loopback con-name stable-microshift ifname lo ip4 $IP/32
```

7. Configure DNS to look at the local name server (note this should
   not be the same IP, it's on a different subnet).

```
nmcli conn modify stable-microshift ipv4.ignore-auto-dns yes
nmcli conn modify stable-microshift ipv4.dns "10.44.1.1"
```

8. Add an entry for the node's hostname from step 5 in /etc/hosts
   using the IP from step 6.

```
echo “$IP $NAME” >> /etc/hosts
```

9. Add the name from step 5 to the MicroShift config file as the
   node.hostnameOverride setting.

10. Add the IP from step 6 to the MicroShift config file as the
    node.nodeIP setting.

11. Restart microshift

```
systemctl enable --now microshift
```

12. Verify it works

```
oc get pods -A
```

13. Reboot the host

14. Verify it works (watch for a while, because the initial pod
    information is cached from before the reboot and is not accurate)

```
oc get pods -A
```

15. Re-enable the NIC

```
IFACE=$(ip route | grep default | awk 'NR==1 {print $5}')
nmcli connection up $IFACE
```

16. Verify that MicroShift is still ok

```
oc get pods -A
```

17. Set the NIC to autoconnect on startup

```
nmcli connection modify $IFACE connection.autoconnect yes
```

18. Reboot the host

19. Verify it works (watch for a while, because the initial pod
    information is cached from before the reboot and is not accurate)

```
watch oc get pods -A
```
