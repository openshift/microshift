# Install MicroShift on RHEL for Edge
To test MicroShift in a setup similar to the production environment, it is necessary to create a RHEL for Edge ISO installer with all the necessary components preloaded on the image.

The procedures described in this document require the following setup:
* A `physical hypervisor host` with the [libvirt](https://libvirt.org/) virtualization platform, to be used for starting virtual machines that run RHEL for Edge OS containing MicroShift binaries
* A `development virtual machine` set up according to the [MicroShift Development Environment](./devenv_setup.md) instructions, to be used for building a RHEL for Edge ISO installer

## Build RHEL for Edge Installer ISO
Log into the `development virtual machine` with the `microshift` user credentials.

Follow the instructions in the [RPM Packages](./devenv_setup.md#rpm-packages) section to create MicroShift RPM packages.

### Prerequisites
Execute the `scripts/devenv-builder/configure-composer.sh` script to install the tools necessary for building the installer image.

Download the OpenShift pull secret from the https://console.redhat.com/openshift/downloads#tool-pull-secret page and save it into the `~/.pull-secret.json` file.

Make sure there is more than 20GB of free disk space necessary for the build artifacts. Run the following command to free the space if necessary.
```bash
~/microshift/scripts/devenv-builder/cleanup-composer.sh -full
```

### Building Installer
> TODO: The `image-builder/build.sh` script has been deprecated.
> This section will be rewritten in the context of [USHIFT-4299](https://issues.redhat.com/browse/USHIFT-4299).

### Disk Partitioning
The `kickstart.ks` file is configured to partition the main disk using `Logical Volume Manager` (LVM). Such partitioning is required for the data volume to be utilized by the MicroShift CSI driver and it allows for flexible file system customization if the disk space runs out.

By default, the following partition layout is created and formatted with the `XFS` file system:
* BIOS boot partition (1MB)
  * It is required to boot ISO to systems with legacy BIOS while using GPT as the partitioning scheme.
* EFI partition with EFI file system (200MB)
* Boot partition is allocated on a 1GB volume
* The rest of the disk is managed by the `LVM` in a single volume group named `rhel`
  * System root partition is allocated on a 10GB volume (minimal recommended size for a root partition)
  * The remainder of the volume group will be used by the CSI driver for storing data (no need to format and mount it)

> The swap partition is not created as it is not required by MicroShift.
> The system root partition size should be specified in megabytes.

As an example, a 20GB disk is partitioned in the following manner by default.
```
$ lsblk /dev/sda
NAME          MAJ:MIN RM  SIZE RO TYPE MOUNTPOINT
sda             8:0    0   20G  0 disk
├─sda1          8:1    0    1M  0 part
├─sda2          8:2    0  200M  0 part /boot/efi
├─sda3          8:3    0  800M  0 part /boot
└─sda4          8:4    0   19G  0 part
  └─rhel-root 253:0    0   10G  0 lvm  /sysroot

$ sudo vgdisplay -s
  "rhel" <19.02 GiB [10.00 GiB  used / <9.02 GiB free]
```

> Unallocated disk space of 9GB size remains in the `rhel` volume group to be used by the CSI driver.

## Install MicroShift for Edge
Log into the `physical hypervisor host` using your user credentials. The remainder of this section describes how to install a virtual machine running RHEL for Edge OS containing MicroShift binaries.

Start by copying the installer image from the `development virtual machine` to the host file system.
```bash
sudo scp microshift@microshift-dev:/home/microshift/microshift/_output/image-builder/microshift-installer-*.$(uname -m).iso /var/lib/libvirt/images/
```

Run the following commands to create a virtual machine using the installer image.
```bash
VMNAME="microshift-edge"
NETNAME="default"
sudo bash -c " \
cd /var/lib/libvirt/images/ && \
virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 3072 \
    --disk path=./${VMNAME}.qcow2,size=20 \
    --network network=${NETNAME},model=virtio \
    --events on_reboot=restart \
    --cdrom ./microshift-installer-*.$(uname -m).iso \
    --noautoconsole \
    --wait \
"
```

Watch the OS console to see the progress of the installation, waiting until the machine is rebooted and the login prompt appears.

Note that it may be more convenient to access the machine using SSH. Run the following command to get its IP address and use it to remotely connect to the system.
```bash
sudo virsh domifaddr microshift-edge
```

Log into the system using `redhat:redhat` credentials and run the following commands to configure MicroShift access.
```bash
mkdir ~/.kube
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
```

Finally, check if MicroShift is up and running by executing `oc` commands.
```bash
oc get cs
oc get pods -A
```
