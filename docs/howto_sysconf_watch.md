# MicroShift Mitigation of System Configuration Changes

MicroShift depends on the following system settings to remain consistent during its runtime:
- Device IP address
- System-wide clock settings
- Iptable configurations
However, these settings may occasionally change on edge devices (i.e. DHCP or NTP updates). When such changes occur, some MicroShift components may stop functioning properly. To mitigate this situation, MicroShift monitors the mentioned system configuration settings, restarts or reloads components if a setting change is detected.

This document describes how to simulate system configuration changes in a virtual environment and verify that MicroShift service reacts by restarting when necessary.

## Create MicroShift Server
Use the instructions in the [Install MicroShift on RHEL for Edge](./rhel4edge_iso.md) document to configure a virtual machine running MicroShift. 

Log into the virtual machine and run the following commands to configure the MicroShift access and check if the PODs are up and running.

```
mkdir ~/.kube
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
oc get pods -A
```

MicroShift startups and restarts can be detected by examining the service output.

```bash
sudo journalctl -xu microshift | egrep 'Starting MicroShift|restarting MicroShift'
```

## IP Address Changes
Log into the hypervisor host and examine the `libvirt` settings to select a new IP address not conflicting with the existing allocations.

Examine the hypervisor DHCP server range.

```bash
sudo virsh net-dumpxml default | grep '<range'
      <range start='192.168.122.2' end='192.168.122.35'/>
```

List the IP addresses already allocated from the DHCP pool.

```bash
sudo virsh net-dhcp-leases default
 Expiry Time           MAC address         Protocol   IP address          Hostname         Client ID or DUID
-----------------------------------------------------------------------------------------------------------------
 2022-07-05 16:32:35   52:54:00:75:23:32   ipv4       192.168.122.21/24   -                01:52:54:00:75:23:32
```

> Pick an IP address from the DHCP range that does *not* already appear in the DHCP lease pool (i.e. 192.168.122.22 based on the above output)

Proceed by logging into the virtual machine **console** using `virt-manager` or `cockpit` interfaces.
> Remote login connections are awkward to use for this experiment because they will be interrupted after the IP address change.

Set the variables denoting the current and the new IP addresses.

```
IPCUR=192.168.122.21
IPNEW=192.168.122.22
```

Run the following command to get the name of the network interface for the current IP address.

```bash
IFACE=$(ip route show | grep $IPCUR | awk '{print $3}')
```

Replace the IP address on the network interface running the following commands.

```bash
sudo ip addr add $IPNEW dev $IFACE
sudo ip addr del $IPCUR dev $IFACE
```

Run the `journalctl` command to verify that the service was restarted. The logs should contain restart and startup messages.
```
Jul 05 09:54:51 localhost.localdomain microshift[1146]: W0705 09:54:51.834933    5803 sysconfwatch.go:81] IP address has changed from "192.168.122.21" to "192.168.122.22", restarting MicroShift
Jul 05 09:54:51 localhost.localdomain microshift[5345]: I0705 09:54:51.306117    6088 run.go:120] Starting MicroShift
```

To restore the proper IP address setting, reboot the virtual machine so that the address is reset back to normal by the DHCP service.

## System-wide Clock Changes
Log into the virtual machine to simulate discontinuous system-wide clock changes using the `timedatectl` command.

> MicroShift restarts when the time is adjusted by more than 10 seconds in the past or the future. 
> Smaller time drifts are allowed to avoid unnecessary restarts on regular time adjustments performed by the NTP service.

### Clock Update with Restart
Reset the clock with a drift of more than 10 seconds to cause the MicroShift service restart.

```bash
sudo timedatectl set-ntp false
sudo timedatectl set-time 00:00:00
```

Run the `journalctl` command to verify that the service was restarted. The logs should contain restart and startup messages.

```
Jul 05 00:00:03 localhost.localdomain microshift[5803]: W0705 00:00:03.834933    5803 sysconfwatch.go:91] realtime clock change detected, time drifted -48955 seconds, restarting MicroShift
Jul 05 00:00:04 localhost.localdomain microshift[6088]: I0705 00:00:04.306117    6088 run.go:120] Starting MicroShift
```

### Clock Update without Restart
Reset the clock with a drift of less than 10 seconds to cause the MicroShift service to log a warning message, but continue execution.

```bash
sudo timedatectl set-ntp false
sudo timedatectl set-time $(date +%H:%M:%S)
```

Run the `journalctl` command to verify that the service was not restarted. The logs should contain a warning message.

```
W0707 00:17:07.309549  157061 sysconfwatch.go:118] realtime clock change detected, time drifted 0 seconds within the allowed range
```
### Restore Clock Setting
To restore the normal clock setting, re-enable the NTP using the following command.
```bash
sudo timedatectl set-ntp true
```

> MicroShift may be restarted again after the system-wide time got corrected by the NTP.

### Certificate Lifetime and Rotation

Microshift certificates are separated into two basic groups:

- long lived certificates with certificate validity of **10 years**
- short lived certificates with certificate validity of **1 year**

Most of the leaf certificates are short lived.

An example of a long-lived certificate is the client certificate for `system:admin`
user authentication, or the certificate of the signer of the kube-apiserver
external serving certificate.

The below (non-proportional!) graph shows when certificates are rotated.

![Cert Rotation](./images/certrotation.png)

- a certificate in the **green zone** does not get rotated
- a certificate in the **yellow zone** is rotated on Microshift start (or restart)
- Microshift will get restarted should a certificate get to the **red zone**, the
  certificate will be rotated for a new one.

If the rotated certificate is a CA, all of the certificates it signed get rotated
as well.

## Firewall Changes

Reload the firewall rules with the following command to trigger the reloading of MicroShift components.

```bash
sudo firewall-cmd --reload
```

Firewall reload action flushes the iptable configurations which results in failed network traffic. <br>
Run the `journalctl -xu microshift` command to verify that the components are reloaded. The logs should contain reload messages.

```
Dec 21 08:57:01 localhost.localdomain microshift[2005232]: infrastructure-services-manager I1221 08:57:01.567046 2005232 iptables.go:590] iptables canary mangle/MICROSHIFT-CANARY deleted
Dec 21 08:57:01 localhost.localdomain microshift[2005232]: infrastructure-services-manager W1221 08:57:01.582233 2005232 components.go:25] Iptables flush is detected, reloading affected components
Dec 21 08:57:01 localhost.localdomain microshift[2005232]: infrastructure-services-manager I1221 08:57:01.582276 2005232 components.go:64] Reload ingress controller
Dec 21 08:57:01 localhost.localdomain microshift[2005232]: infrastructure-services-manager I1221 08:57:01.644365 2005232 components.go:69] Reload CNI plugin
```
