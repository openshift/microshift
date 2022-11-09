# Automating Development Environment Setup
This document describes how to automate the setup of MicroShift development environment per instructions in the following sections:
* [Create Development Virtual Machine](./devenv_rhel8.md#create-development-virtual-machine)
* [Build MicroShift](./devenv_rhel8.md#build-microshift)
* [Run MicroShift Executable](./devenv_rhel8.md#run-microshift-executable)

It is recommended to review the above mentioned documentation sections before proceeding with the automation instructions in the current document.

## Create Virtual Machine
Log into the hypervisor host and run the `create-vm.sh` script without arguments to see its usage.
```bash
$ ./scripts/devenv-builder/create-vm.sh
Usage: create-vm.sh [<VMNAME> <VMDISKDIR> <RHELISO> <NCPUS> <RAMSIZE> <DISKSIZE> <SWAPSIZE> <DATAVOLSIZE>]
INFO: Specify 0 swap size to disable swap partition
INFO: Positional arguments also can be specified using environment variables
INFO: All sizes in GB

ERROR: Invalid VM name: ''
```
> It is mandatory to use the RHEL 8 DVD (not boot) image for installing the virtual machine.

As an example, run the following command to create a virtual machine named `microshift-dev` with 4 CPUs, 6GB of RAM, 50GB of total disk space, 6GB of swap and 2GB disk allocated to the data volume.
> The recommended swap partition size depends on the amount of RAM.
> See the [Recommended system swap size](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/managing_storage_devices/getting-started-with-swap_managing-storage-devices#recommended-system-swap-space_getting-started-with-swap) document for more information.

```bash
./scripts/devenv-builder/create-vm.sh microshift-dev /var/lib/libvirt/images /var/lib/libvirt/images/rhel-8.7-x86_64-dvd.iso 4 6 50 6 2
```

or

```bash
export VMNAME=microshift-dev
export VMDISKDIR=/var/lib/libvirt/images
export RHELISO=/var/lib/libvirt/images/rhel-8.7-x86_64-dvd.iso
export NCPUS=4
export RAMSIZE=6
export DISKSIZE=50
export SWAPSIZE=6
export DATAVOLSIZE=2
./scripts/devenv-builder/create-vm.sh
```

Watch the output in the `Virt Viewer` popup, waiting for a successful completion of the installation procedure.

## Configure Virtual Machine
Log into the hypervisor host and run the following command to get the IP address of the development virtual machine and use it to remotely connect to the system.
```bash
sudo virsh domifaddr microshift-dev
```

Copy the configuration script and your OpenShift pull secret file to the virtual machine using `microshift:microshift` credentials.
```bash
VMIPADDR=192.168.122.29
scp scripts/devenv-builder/configure-vm.sh microshift@${VMIPADDR}:
scp ~/.pull-secret.json microshift@${VMIPADDR}:
```

Log into the development virtual machine using `microshift:microshift` credentials and run the following command to configure MicroShift.
```bash
~/configure-vm.sh ~/.pull-secret.json
```

If the machine is not yet registered to the Red Hat subscription management service, the script prompts for the user name and password to be used for subscribing.

In the event of a successful subscription, the configuration script should run unattended and finish with the following message.
```
The configuration phase completed. Run the following commands to:
 - Wait until all MicroShift pods are running
 - Clean up MicroShift service configuration

watch sudo /usr/local/bin/oc --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig get pods -A
echo 1 | /usr/bin/cleanup-all-microshift-data

Done
```
