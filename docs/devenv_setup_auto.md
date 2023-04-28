# Automating Development Environment Setup
This document describes how to automate the setup of MicroShift development environment per instructions in the following sections:
* [Create Development Virtual Machine](./devenv_setup.md#create-development-virtual-machine)
* [Build MicroShift](./devenv_setup.md#build-microshift)
* [Run MicroShift Executable](./devenv_setup.md#run-microshift-executable)

It is recommended to review the above mentioned documentation sections before proceeding with the automation instructions in the current document.

## Manage Virtual Machine

Use the `manage-vm.sh` script on the hypervisor to manipulate the virtual machine where MicroShift will run.

Running it without any arguments prints the usage message.
```bash
$ ./scripts/devenv-builder/manage-vm.sh
```

Use the `create` subcommand to create a new VM.

Use the `delete` subcommand to remove a VM.

Use the `ip` subcommand to find the IP address of a VM.

If your host is configured with VM disk storage somewhere other than
the default, set `MICROSHIFT_VMDISKDIR` to point to the location, for
example "$HOME/VMs".

Set `MICROSHIFT_SSH_KEY_FILE` to the path to an ssh public key to have
that key added to the VM for easier login.

## Configure Virtual Machine
Log into the hypervisor host and run the following command to get the IP address of the development virtual machine and use it to remotely connect to the system.
```bash
sudo virsh domifaddr microshift-dev
...
...
VMIPADDR=192.168.122.29
```

Configure SSH not to require a password when logging into the system.
```bash
ssh-copy-id -f microshift@${VMIPADDR}
```

Copy the configuration script and your OpenShift pull secret file to the virtual machine using `microshift:microshift` credentials.
```bash
scp scripts/devenv-builder/configure-vm.sh microshift@${VMIPADDR}:
scp ~/.pull-secret.json microshift@${VMIPADDR}:
```

Log into the development virtual machine using `microshift:microshift` credentials and run the following command to configure MicroShift.
```bash
~/configure-vm.sh ~/.pull-secret.json
```

> When configuring a RHEL OS and the machine is not yet registered to the Red Hat subscription management service,
> the script prompts for the user name and password to be used for subscribing.

The configuration script should run unattended and finish with the following message.
```
The configuration phase completed. Run the following commands to:
 - Wait until all MicroShift pods are running
 - Clean up MicroShift service configuration

watch sudo $(which oc) --kubeconfig /var/lib/microshift/live/resources/kubeadmin/kubeconfig get pods -A
echo 1 | sudo /usr/bin/microshift-cleanup-data --all

Done
```

## Create Virtual Machine (low level)
Log into the hypervisor host and run the `create-vm.sh` script without arguments to see its usage.
```bash
$ ./scripts/devenv-builder/create-vm.sh
Usage: create-vm.sh [<VMNAME> <VMDISKDIR> <ISOFILE> <NCPUS> <RAMSIZE> <DISKSIZE> <SWAPSIZE> <DATAVOLSIZE>]
INFO: Specify 0 swap size to disable swap partition
INFO: Positional arguments also can be specified using environment variables
INFO: All sizes in GB

ERROR: Invalid VM name: ''
```

> Use the DVD (not boot) image for bootstrapping the virtual machine running a RHEL OS.
> This is necessary to allow for a minimal environment setup before the OS is registered
> with the Red Hat Subscription Management in the subsequent configuration phase.

As an example, run the following command to create a virtual machine named `microshift-dev` with 4 CPUs, 6GB of RAM, 50GB of total disk space, 6GB of swap and 2GB disk allocated to the data volume.
> The recommended swap partition size depends on the amount of RAM.
> See the [Recommended system swap size](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/managing_storage_devices/getting-started-with-swap_managing-storage-devices#recommended-system-swap-space_getting-started-with-swap) document for more information.

```bash
ISONAME=rhel-baseos-9.1-$(uname -m)-dvd.iso

./scripts/devenv-builder/create-vm.sh microshift-dev \
    /var/lib/libvirt/images \
    /var/lib/libvirt/images/${ISONAME} \
    4 6 50 6 2
```

or

```bash
ISONAME=rhel-baseos-9.1-$(uname -m)-dvd.iso

export VMNAME=microshift-dev
export VMDISKDIR=/var/lib/libvirt/images
export ISOFILE=/var/lib/libvirt/images/${ISONAME}
export NCPUS=4
export RAMSIZE=6
export DISKSIZE=50
export SWAPSIZE=6
export DATAVOLSIZE=2
./scripts/devenv-builder/create-vm.sh
```

Watch the output in the `virt-viewer` popup, waiting for a successful completion of the installation procedure.
