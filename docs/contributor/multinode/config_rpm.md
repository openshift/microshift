# RHEL with MicroShift RPMs
Log into the `hypervisor host` and follow the instructions described in the
[Getting Started with MicroShift](../../user/getting_started.md) document to create
and configure the `microshift-pri` and `microshift-sec` virtual machines for
running primary and secondary instances.

> Until the MicroShift 4.14 software is released, it is necessary to compile the
> MicroShift RPMs from the latest sources on the `development host` and copy
> them to the `microshift-pri` and `microshift-sec` virtual machines.
>
> See the [RPM Packages](../devenv_setup.md#rpm-packages) documentation
> for more information on building MicroShift RPMs.
>
> Run the following command on the `microshift-pri` and `microshift-sec`
> virtual machines to upgrade the software.
> ```
> sudo dnf localinstall -y microshift*.rpm
> ```

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
