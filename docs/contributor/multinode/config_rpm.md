# RHEL with MicroShift RPMs
Log into the `hypervisor host` and follow the instructions described in the
[Getting Started with MicroShift](../../user/getting_started.md) document to create
and configure the `microshift-pri` and `microshift-sec` virtual machines for
running primary and secondary instances.

After the virtual machines are up and running, use the `virsh` command to determine their IP addresses.
```
sudo virsh domifaddr microshift-pri
sudo virsh domifaddr microshift-sec
```

The multinode configuration procedure uses SSH for copying files among the
nodes and remotely running scripts on the primary and secondary hosts.
Having your SSH keys authorized in the nodes would make their configuration
procedure more convenient and streamlined.
```
PRI_ADDR=192.168.122.118
SEC_ADDR=192.168.122.70

# Use 'redhat' without quotes as a password when prompted
ssh-copy-id redhat@${PRI_ADDR}
ssh-copy-id redhat@${SEC_ADDR}
```
